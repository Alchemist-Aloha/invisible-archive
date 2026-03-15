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
	queue      chan string   // Background generation queue
}

func NewThumbnailer(vfs *vfs.Manager, cacheDir string, maxWorkers int) (*Thumbnailer, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	t := &Thumbnailer{
		vfs:       vfs,
		cacheDir:  cacheDir,
		concurSem: make(chan struct{}, maxWorkers),
		queue:     make(chan string, 1000), // Buffer for background tasks
	}

	// Start background worker
	go t.backgroundWorker()

	return t, nil
}

// QueueBackground adds a path to the background generation queue
func (t *Thumbnailer) QueueBackground(path string) {
	select {
	case t.queue <- path:
		// Queued successfully
	default:
		// Queue full, skip background task
	}
}

func (t *Thumbnailer) backgroundWorker() {
	log.Printf("THUMB: Background worker started")
	for path := range t.queue {
		_, _ = t.generateInternal(path)
	}
}

func (t *Thumbnailer) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	// SVG Bypass: SVGs are their own thumbnails
	if strings.HasSuffix(strings.ToLower(path), ".svg") {
		reader, closer, err := t.vfs.GetRawReader(path)
		if err != nil {
			http.Error(w, "failed to read svg", http.StatusInternalServerError)
			return
		}
		defer closer.Close()
		
		stat, err := t.vfs.Stat(path)
		var modTime time.Time
		if err == nil {
			modTime = stat.ModTime()
		} else {
			modTime = time.Now()
		}

		w.Header().Set("Content-Type", "image/svg+xml")
		http.ServeContent(w, r, "thumb.svg", modTime, reader)
		return
	}

	thumbPath, err := t.generateInternal(path)
	if err != nil {
		log.Printf("THUMB: Error for %s: %v", path, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, thumbPath)
}

func (t *Thumbnailer) generateInternal(path string) (string, error) {
	// 1. FAST-PATH: Try to get metadata from SQLite first to avoid VFS Stat
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
			return "", err
		}
		size = stat.Size()
		modTime = stat.ModTime().Unix()
	}

	// SVG Bypass in internal: just skip
	if strings.HasSuffix(strings.ToLower(path), ".svg") {
		return "", nil
	}

	// 2. Generate Identity Cache Key
	id := fmt.Sprintf("%s-%d-%d", path, size, modTime)
	hash := fmt.Sprintf("%x", md5.Sum([]byte(id)))
	thumbPath := filepath.Join(t.cacheDir, hash+".webp")

	// 3. Check Cache
	if _, err := os.Stat(thumbPath); err == nil {
		return thumbPath, nil
	}

	// 4. SINGLE-FLIGHT: Ensure only one worker generates this specific thumbnail
	_, err, _ := t.processing.Do(hash, func() (interface{}, error) {
		// Re-check cache inside singleflight
		if _, err := os.Stat(thumbPath); err == nil {
			return nil, nil
		}

		// Throttle global concurrent generations
		t.concurSem <- struct{}{}
		defer func() { <-t.concurSem }()

		// Generate
		reader, closer, err := t.vfs.GetRawReader(path)
		if err != nil {
			return nil, err
		}
		defer closer.Close()

		img, err := vips.NewImageFromReader(reader)
		if err != nil {
			return nil, err
		}
		defer img.Close()

		// Resize using Thumbnail (shrink-on-load)
		if err := img.Thumbnail(200, 200, vips.InterestingNone); err != nil {
			return nil, err
		}

		// Export to WebP
		wp := vips.NewWebpExportParams()
		wp.Quality = 75
		thumbBytes, _, err := img.ExportWebp(wp)
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(thumbPath, thumbBytes, 0644); err != nil {
			return nil, err
		}

		return nil, nil
	})

	return thumbPath, err
}
