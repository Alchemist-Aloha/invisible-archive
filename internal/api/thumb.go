package api

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/disintegration/imaging"
	"github.com/likun/invisible-archive/internal/vfs"
)

type Thumbnailer struct {
	vfs        *vfs.Manager
	cacheDir   string
	concurSem  chan struct{} // Semaphore for throttling
	mu         sync.Mutex
	processing map[string]chan struct{} // Track currently processing thumbs
}

func NewThumbnailer(vfs *vfs.Manager, cacheDir string, maxWorkers int) (*Thumbnailer, error) {
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, err
	}

	return &Thumbnailer{
		vfs:        vfs,
		cacheDir:   cacheDir,
		concurSem:  make(chan struct{}, maxWorkers),
		processing: make(map[string]chan struct{}),
	}, nil
}

func (t *Thumbnailer) GetThumbnail(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		http.Error(w, "missing path", http.StatusBadRequest)
		return
	}

	stat, err := t.vfs.Stat(path)
	if err != nil {
		http.Error(w, "file not found", http.StatusNotFound)
		return
	}

	// Fast Identity Cache Key
	id := fmt.Sprintf("%s-%d-%d", path, stat.Size(), stat.ModTime().Unix())
	hash := fmt.Sprintf("%x", md5.Sum([]byte(id)))
	thumbPath := filepath.Join(t.cacheDir, hash+".webp")

	// Check cache
	if _, err := os.Stat(thumbPath); err == nil {
		http.ServeFile(w, r, thumbPath)
		return
	}

	// Throttle and generate
	t.concurSem <- struct{}{}
	defer func() { <-t.concurSem }()

	// Re-check cache inside semaphore (might have been generated while waiting)
	if _, err := os.Stat(thumbPath); err == nil {
		http.ServeFile(w, r, thumbPath)
		return
	}

	// Generate
	reader, closer, err := t.vfs.GetRawReader(path)
	if err != nil {
		http.Error(w, "failed to read file", http.StatusInternalServerError)
		return
	}
	defer closer.Close()

	src, err := imaging.Decode(reader)
	if err != nil {
		http.Error(w, "failed to decode image", http.StatusInternalServerError)
		return
	}

	// Resize to 200px width, preserving aspect ratio
	dst := imaging.Resize(src, 200, 0, imaging.Lanczos)

	// Save to cache
	if err := imaging.Save(dst, thumbPath); err != nil {
		http.Error(w, "failed to save thumbnail", http.StatusInternalServerError)
		return
	}

	http.ServeFile(w, r, thumbPath)
}
