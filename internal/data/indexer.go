package data

import (
	"archive/zip"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/likun/invisible-archive/pkg/util"
	_ "modernc.org/sqlite"
)

type Indexer struct {
	db      *sql.DB
	queries *Queries
	watcher *fsnotify.Watcher
	library string
}

func NewIndexer(dbPath, library string) (*Indexer, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, err
	}

	// Enable WAL mode for high performance
	if _, err := db.Exec("PRAGMA journal_mode=WAL;"); err != nil {
		return nil, err
	}

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
	}, nil
}

func (idx *Indexer) Close() error {
	idx.watcher.Close()
	return idx.db.Close()
}

func (idx *Indexer) GetQueries() *Queries {
	return idx.queries
}

// IndexDirectory indexes a physical directory non-recursively
func (idx *Indexer) IndexDirectory(ctx context.Context, physicalPath string) error {
	entries, err := os.ReadDir(physicalPath)
	if err != nil {
		return err
	}

	// Calculate paths relative to library root
	absPhysical, _ := filepath.Abs(physicalPath)
	relParent, err := filepath.Rel(idx.library, absPhysical)
	if err != nil {
		return err
	}
	if relParent == "." {
		relParent = ""
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath := "/" + filepath.Join(relParent, entry.Name())
		
		// Use shared capability logic
		caps := int64(util.GetCapabilities(entry.Name(), info.IsDir()))

		err = idx.queries.UpsertItem(ctx, UpsertItemParams{
			ParentPath:  "/" + relParent,
			Name:        entry.Name(),
			Path:        relPath,
			IsDir:       info.IsDir(),
			Size:        info.Size(),
			ModTime:     info.ModTime().Unix(),
			Capabilities: caps,
			IsInsideZip: false,
		})
		if err != nil {
			return err
		}

		// Shallowly index ZIP contents in the background
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".zip") {
			relZipPath := filepath.ToSlash(filepath.Join(relParent, entry.Name()))
			go idx.IndexZip(ctx, filepath.Join(physicalPath, entry.Name()), relZipPath)
		}
	}

	// Start watching this directory
	return idx.watcher.Add(physicalPath)
}

// WatchLoop runs the real-time file watcher
func (idx *Indexer) WatchLoop(ctx context.Context) {
	for {
		select {
		case event, ok := <-idx.watcher.Events:
			if !ok {
				return
			}
			// Handle file system changes (Create, Write, Rename, Remove)
			// For simplicity, re-index the parent directory
			parent := filepath.Dir(event.Name)
			idx.IndexDirectory(ctx, parent)

		case err, ok := <-idx.watcher.Errors:
			if !ok {
				return
			}
			// Log error (should use a logger)
			_ = err
		case <-ctx.Done():
			return
		}
	}
}

// IndexZip indexes the internal structure of a ZIP archive
func (idx *Indexer) IndexZip(ctx context.Context, physicalPath, relZipPath string) error {
	r, err := zip.OpenReader(physicalPath)
	if err != nil {
		return err
	}
	defer r.Close()

	for _, f := range r.File {
		// ZIP paths are always forward-slash separated
		// Performance: Avoid allocation-heavy strings.Split/Join in tight loop.
		// LastIndexByte provides O(1) allocation path parsing.
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

		// ZIP paths are stored absolute-looking relative to VFS root
		parentPath := "/" + filepath.Join(relZipPath, parentInZip)
		fullPath := "/" + filepath.Join(relZipPath, f.Name)
		isDir := f.FileInfo().IsDir()

		caps := int64(util.GetCapabilities(name, isDir))

		err = idx.queries.UpsertItem(ctx, UpsertItemParams{
			ParentPath:  parentPath,
			Name:        name,
			Path:        fullPath,
			IsDir:       isDir,
			Size:        int64(f.UncompressedSize64),
			ModTime:     f.Modified.Unix(),
			Capabilities: caps,
			IsInsideZip: true,
		})
		if err != nil {
			return err
		}
	}
	return nil
}
