package api

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/likun/invisible-archive/internal/vfs"
	"golang.org/x/sync/singleflight"
)

type Thumbnailer struct {
	vfs        *vfs.Manager
	cacheDir   string
	concurSem  chan struct{} // Semaphore for throttling
	mu         sync.Mutex
	processing singleflight.Group
}

func NewThumbnailer(vfs *vfs.Manager, cacheDir string, maxWorkers int) (*Thumbnailer, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &Thumbnailer{
		vfs:       vfs,
		cacheDir:  cacheDir,
		concurSem: make(chan struct{}, maxWorkers),
	}, nil
}

func (t *Thumbnailer) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	// 1. FAST-PATH: Try to get metadata from SQLite first to avoid VFS Stat (slow for ZIPs)
	var size int64
	var modTime int64
	
	if t.vfs.GetIndexer() != nil {
		item, err := t.vfs.GetIndexer().GetQueries().GetItemByPath(context.Background(), path)
		if err == nil {
			size = item.Size
			modTime = item.ModTime
		}
	}

	// Fallback to VFS Stat if not indexed yet
	if size == 0 {
		stat, err := t.vfs.Stat(path)
		if err != nil {
			log.Printf("THUMB: File not found: %s, err: %v", path, err)
			http.Error(w, "file not found", http.StatusNotFound)
			return
		}
		size = stat.Size()
		modTime = stat.ModTime().Unix()
	}

	// SVG Bypass: SVGs are their own thumbnails
	if strings.HasSuffix(strings.ToLower(path), ".svg") {
		reader, closer, err := t.vfs.GetRawReader(path)
		if err != nil {
			http.Error(w, "failed to read svg", http.StatusInternalServerError)
			return
		}
		defer closer.Close()
		w.Header().Set("Content-Type", "image/svg+xml")
		// Using a dummy name since we don't have the stat object here for simplicity
		// ServeContent handles Range and Caching
		http.ServeContent(w, r, "thumb.svg", time.Unix(modTime, 0), reader)
		return
	}

	// 2. Generate Identity Cache Key
	id := fmt.Sprintf("%s-%d-%d", path, size, modTime)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(id)))
	thumbPath := filepath.Join(t.cacheDir, hash+".webp")

	// 3. Check Cache
	if _, err := os.Stat(thumbPath); err == nil {
		http.ServeFile(w, r, thumbPath)
		return
	}

	// 4. SINGLE-FLIGHT: Ensure only one worker generates this specific thumbnail
	_, err, _ := t.processing.Do(hash, func() (interface{}, error) {
		// Re-check cache inside singleflight
		if _, err := os.Stat(thumbPath); err == nil {
			return nil, nil
		}

		log.Printf("THUMB: Generating (WebP) for %s", path)

		// Throttle global concurrent generations
		t.concurSem <- struct{}{}
		defer func() { <-t.concurSem }()

		// Generate
		reader, closer, err := t.vfs.GetRawReader(path)
		if err != nil {
			return nil, fmt.Errorf("failed to get reader: %w", err)
		}
		defer closer.Close()

		img, err := vips.NewImageFromReader(reader)
		if err != nil {
			return nil, fmt.Errorf("failed to decode: %w", err)
		}
		defer img.Close()

		// Resize using Thumbnail (shrink-on-load)
		if err := img.Thumbnail(200, 200, vips.InterestingNone); err != nil {
			return nil, fmt.Errorf("failed to resize: %w", err)
		}

		// Export to WebP (smaller than JPEG)
		wp := vips.NewWebpExportParams()
		wp.Quality = 75
		thumbBytes, _, err := img.ExportWebp(wp)
		if err != nil {
			return nil, fmt.Errorf("failed to export webp: %w", err)
		}

		if err := os.WriteFile(thumbPath, thumbBytes, 0644); err != nil {
			return nil, fmt.Errorf("failed to save: %w", err)
		}

		return nil, nil
	})

	if err != nil {
		log.Printf("THUMB: Error for %s: %v", path, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, thumbPath)
}
