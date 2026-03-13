package api

import (
	"encoding/json"
	"log"
	"net/http"
	"path"

	"github.com/go-chi/chi/v5"
	"github.com/likun/invisible-archive/internal/vfs"
	"github.com/likun/invisible-archive/pkg/util"
)

type FileItem struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	IsDir        bool   `json:"is_dir"`
	Size         int64  `json:"size"`
	ModTime      int64  `json:"mod_time"`
	Capabilities uint32 `json:"capabilities"`
}

type Handler struct {
	vfs *vfs.Manager
}

func NewHandler(vfs *vfs.Manager) *Handler {
	return &Handler{vfs: vfs}
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	requestPath := r.URL.Query().Get("path")
	if requestPath == "" || requestPath == "." {
		requestPath = "/"
	}
	log.Printf("API: Listing path: %s", requestPath)

	files, err := h.vfs.ReadDir(requestPath)
	if err != nil {
		log.Printf("API: Error listing path %s: %v", requestPath, err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("API: Found %d items in %s", len(files), requestPath)

	items := make([]FileItem, 0, len(files))
	for _, f := range files {
		items = append(items, FileItem{
			Name:         f.Name(),
			Path:         path.Join(requestPath, f.Name()),
			IsDir:        f.IsDir(),
			Size:         f.Size(),
			ModTime:      f.ModTime().Unix(),
			Capabilities: util.GetCapabilities(f.Name(), f.IsDir()),
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(items)
}

func (h *Handler) Raw(w http.ResponseWriter, r *http.Request) {
	pathParam := chi.URLParam(r, "*")
	if pathParam == "" {
		pathParam = r.URL.Query().Get("path")
	}

	reader, closer, err := h.vfs.GetRawReader(pathParam)
	if err != nil {
		log.Printf("API: Failed to get reader for %s: %v", pathParam, err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	defer closer.Close()

	// Use Stat to get size and modtime for ServeContent
	stat, err := h.vfs.Stat(pathParam)
	if err != nil {
		log.Printf("API: Failed to stat %s: %v", pathParam, err)
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
