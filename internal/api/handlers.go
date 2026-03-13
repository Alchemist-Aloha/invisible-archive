package api

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/likun/invisible-archive/internal/vfs"
)

type FileItem struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IsDir        bool   `json:"is_dir"`
	Size         int64  `json:"size"`
	ModTime      int64  `json:"mod_time"`
	Capabilities uint32 `json:"capabilities"`
}

const (
	CapBrowse uint32 = 1 << iota // 1
	CapStream                    // 2
	CapRender                    // 4
	CapEdit                      // 8
)

type Handler struct {
	vfs *vfs.Manager
}

func NewHandler(vfs *vfs.Manager) *Handler {
	return &Handler{vfs: vfs}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Query().Get("path")
	if path == "" {
		path = "."
	}
	log.Printf("API: Listing path: %s", path)

	files, err := h.vfs.ReadDir(path)
	if err != nil {
		log.Printf("API: Error listing path %s: %v", path, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("API: Found %d items in %s", len(files), path)

	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		items = append(items, FileItem{
			Name:         f.Name(),
			Path:         filepath.Join(path, f.Name()),
			IsDir:        f.IsDir(),
			Size:         f.Size(),
			ModTime:      f.ModTime().Unix(),
			Capabilities: getCapabilities(f),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) Raw(w http.ResponseWriter, r *http.Request) {
	path := chi.URLParam(r, "*")
	if path == "" {
		path = r.URL.Query().Get("path")
	}

	reader, closer, err := h.vfs.GetRawReader(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer closer.Close()

	// Use Stat to get size and modtime for ServeContent
	stat, err := h.vfs.Stat(path)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.ServeContent(w, r, stat.Name(), stat.ModTime(), reader)
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		http.Error(w, "query parameter 'q' is required", http.StatusBadRequest)
		return
	}

	pattern := "%" + q + "%"
	files, err := h.vfs.Search(r.Context(), pattern)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		items = append(items, FileItem{
			Name:         f.Name,
			Path:         f.Path,
			IsDir:        f.IsDir,
			Size:         f.Size,
			ModTime:      f.ModTime,
			Capabilities: uint32(f.Capabilities),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func getCapabilities(f os.FileInfo) uint32 {
	var caps uint32
	name := strings.ToLower(f.Name())

	if f.IsDir() || strings.HasSuffix(name, ".zip") {
		caps |= CapBrowse
	}

	ext := filepath.Ext(name)
	switch ext {
	case ".mp4", ".mkv", ".webm", ".mp3", ".wav":
		caps |= CapStream
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf":
		caps |= CapRender
	case ".txt", ".md", ".go", ".js", ".ts", ".vue", ".css", ".json", ".py", ".cpp", ".h":
		caps |= CapEdit
	}

	return caps
}
