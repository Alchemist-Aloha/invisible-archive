package data

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
	"github.com/likun/invisible-archive/pkg/util"
	_ "modernc.org/sqlite"
)

type Indexer struct {
	db          *sql.DB
	queries     *Queries
	watcher     *fsnotify.Watcher
	library     string
	zipSem      chan struct{} // Concurrency limit for ZIPs
	dirSem      chan struct{} // Concurrency limit for Dirs
	Discovery   func(path string) // Callback for new images
	activeTasks int32         // Atomic counter for tracking
}

func NewIndexer(dbPath, library string) (*Indexer, error) {
	// Set busy_timeout in DSN for maximum reliability with modernc.org/sqlite
	dsn := fmt.Sprintf("file:%s?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=temp_store(MEMORY)", dbPath)
	db, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(1)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	absLib, err := filepath.Abs(library)
	if err != nil {
		return nil, err
	}

	return &Indexer{
		db:      db,
		queries: New(db),
		watcher: watcher,
		library: absLib,
		zipSem:  make(chan struct{}, 4),
		dirSem:  make(chan struct{}, 8), // Gate both walk and write
	}, nil
}

func (idx *Indexer) Close() error {
	idx.watcher.Close()
	return idx.db.Close()
}

func (idx *Indexer) GetQueries() *Queries {
	return idx.queries
}

// IndexRecursive performs a throttled background crawl
func (idx *Indexer) IndexRecursive(ctx context.Context, physicalPath string) {
	atomic.AddInt32(&idx.activeTasks, 1)
	defer atomic.AddInt32(&idx.activeTasks, -1)

	// CRITICAL: Throttle the WALK itself to prevent FD exhaustion and memory spikes
	select {
	case idx.dirSem <- struct{}{}:
		defer func() { <-idx.dirSem }()
	case <-ctx.Done():
		return
	}

	// 1. Index this directory metadata
	if err := idx.indexDirectoryInternal(ctx, physicalPath); err != nil {
		log.Printf("INDEX: Error indexing %s: %v", physicalPath, err)
		// Continue to children even if this specific dir metadata failed
	}

	// 2. Read subdirectories
	entries, err := os.ReadDir(physicalPath)
	if err != nil {
		log.Printf("INDEX: Failed to read dir %s: %v", physicalPath, err)
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			subPath := filepath.Join(physicalPath, entry.Name())
			if strings.HasPrefix(entry.Name(), ".") {
				continue
			}
			// Recursive call - the semaphore at the top will handle throttling
			// We use a goroutine so multiple branches can wait in the dirSem queue
			go idx.IndexRecursive(ctx, subPath)
		}
	}
}

// IndexDirectory is the public entrypoint for targeted scans
func (idx *Indexer) IndexDirectory(ctx context.Context, physicalPath string) error {
	idx.dirSem <- struct{}{}
	defer func() { <-idx.dirSem }()
	return idx.indexDirectoryInternal(ctx, physicalPath)
}

func (idx *Indexer) indexDirectoryInternal(ctx context.Context, physicalPath string) error {
	entries, err := os.ReadDir(physicalPath)
	if err != nil {
		return err
	}

	absPhysical, _ := filepath.Abs(physicalPath)
	relParent, err := filepath.Rel(idx.library, absPhysical)
	if err != nil {
		return err
	}
	if relParent == "." {
		relParent = ""
	}

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := idx.queries.WithTx(tx)

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath := "/" + filepath.Join(relParent, entry.Name())
		caps := uint32(util.GetCapabilities(entry.Name(), info.IsDir()))

		err = qtx.UpsertItem(ctx, UpsertItemParams{
			ParentPath:  "/" + relParent,
			Name:        entry.Name(),
			Path:        relPath,
			IsDir:       info.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime().Unix(),
			Capabilities: int64(caps),
			IsInsideZip: false,
		})
		if err != nil {
			return err
		}

		if !info.IsDir() && (caps&util.CapRender) != 0 && idx.Discovery != nil {
			idx.Discovery(relPath)
		}

		if !info.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			relZipPath := filepath.ToSlash(filepath.Join(relParent, entry.Name()))
			go idx.IndexZip(ctx, filepath.Join(physicalPath, entry.Name()), relZipPath)
		}
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	_ = idx.watcher.Add(physicalPath)
	return nil
}

func (idx *Indexer) WatchLoop(ctx context.Context) {
	for {
		select {
		case event, ok := <-idx.watcher.Events:
			if !ok {
				return
			}
			parent := filepath.Dir(event.Name)
			go idx.IndexDirectory(ctx, parent)
		case <-ctx.Done():
			return
		}
	}
}

func (idx *Indexer) IndexZip(ctx context.Context, physicalPath, relZipPath string) error {
	idx.zipSem <- struct{}{}
	defer func() { <-idx.zipSem }()

	r, err := zip.OpenReader(physicalPath)
	if err != nil {
		return err
	}
	defer r.Close()

	tx, err := idx.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	qtx := idx.queries.WithTx(tx)

	for _, f := range r.File {
		cleanName := strings.TrimSuffix(f.Name, "/")
		slashIdx := strings.LastIndexByte(cleanName, '/')

		var name, parentInZip string
		if slashIdx != -1 {
			name = cleanName[slashIdx+1:]
			parentInZip = cleanName[:slashIdx]
		} else {
			name = cleanName
			parentInZip = ""
		}

		parentPath := "/" + filepath.Join(relZipPath, parentInZip)
		fullPath := "/" + filepath.Join(relZipPath, f.Name)
		isDir := f.FileInfo().IsDir()

		caps := uint32(util.GetCapabilities(name, isDir))

		err = qtx.UpsertItem(ctx, UpsertItemParams{
			ParentPath:  parentPath,
			Name:        name,
			Path:        fullPath,
			IsDir:       isDir,
			Size:        int64(f.UncompressedSize64),
			ModTime:     f.Modified.Unix(),
			Capabilities: int64(caps),
			IsInsideZip: true,
		})
		if err != nil {
			return err
		}

		if !isDir && (caps&util.CapRender) != 0 && idx.Discovery != nil {
			idx.Discovery(fullPath)
		}
	}

	return tx.Commit()
}

// GetActiveTasks returns the number of directories currently being scanned
func (idx *Indexer) GetActiveTasks() int32 {
	return atomic.LoadInt32(&idx.activeTasks)
}
