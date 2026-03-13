package data

import (
	"archive/zip"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
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

	return &Indexer{
		db:      db,
		queries: New(db),
		watcher: watcher,
		library: filepath.Clean(library),
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

	// Calculate paths relative to system root for VFS consistency
	relParent, _ := filepath.Rel("/", physicalPath)
	if relParent == "." {
		relParent = ""
	}

	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		relPath := "/" + filepath.Join(relParent, entry.Name())
		isZip := !info.IsDir() && strings.ToLower(filepath.Ext(entry.Name())) == ".zip"
		
		// Use a simple bitmask for capabilities
		var caps int64 = 0
		if info.IsDir() || isZip {
			caps |= 1 // can_browse
		}
		// ... (rest of logic for other capabilities should be here or handled in API)

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
		parts := strings.Split(strings.TrimSuffix(f.Name, "/"), "/")
		name := parts[len(parts)-1]
		parentInZip := ""
		if len(parts) > 1 {
			parentInZip = strings.Join(parts[:len(parts)-1], "/")
		}

		// ZIP paths are stored absolute-looking relative to VFS root
		parentPath := "/" + filepath.Join(relZipPath, parentInZip)
		fullPath := "/" + filepath.Join(relZipPath, f.Name)
		isDir := f.FileInfo().IsDir()

		var caps int64 = 0
		if isDir || strings.HasSuffix(strings.ToLower(name), ".zip") {
			caps |= 1 // can_browse
		}

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
