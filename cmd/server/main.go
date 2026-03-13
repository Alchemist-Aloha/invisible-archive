package main

import (
	"context"
	"database/sql"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/likun/invisible-archive/internal/api"
	"github.com/likun/invisible-archive/internal/data"
	"github.com/likun/invisible-archive/internal/vfs"
)

func main() {
	// Register additional mime types
	mime.AddExtensionType(".m4v", "video/mp4") // m4v is basically mp4, browsers handle video/mp4 better
	mime.AddExtensionType(".mkv", "video/x-matroska")

	libraryPath := os.Getenv("LIBRARY_PATH")
	if libraryPath == "" {
		libraryPath = "./library" // Default for local dev
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "./archive.db"
	}

	// Ensure library path exists
	if err := os.MkdirAll(libraryPath, 0755); err != nil {
		log.Fatalf("failed to create library path: %v", err)
	}

	// Proper schema initialization
	initDB(dbPath, "internal/data/schema.sql")

	// Initialize Indexer
	indexer, err := data.NewIndexer(dbPath, libraryPath)
	if err != nil {
		log.Fatalf("failed to initialize indexer: %v", err)
	}
	defer indexer.Close()

	// Start watch loop
	go indexer.WatchLoop(context.Background())

	// Initialize VFS Manager rooted at libraryPath
	vfsMgr, err := vfs.NewManager(libraryPath, 50, indexer)

	if err != nil {
		log.Fatalf("failed to initialize VFS: %v", err)
	}

	// Initialize Thumbnailer
	cacheDir := os.Getenv("CACHE_DIR")
	if cacheDir == "" {
		cacheDir = "./cache"
	}
	thumb, err := api.NewThumbnailer(vfsMgr, filepath.Join(cacheDir, "thumbs"), 2)
	if err != nil {
		log.Fatalf("failed to initialize thumbnailer: %v", err)
	}

	h := api.NewHandler(vfsMgr)

	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	
	// Simple CORS middleware
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, Range")
			w.Header().Set("Access-Control-Expose-Headers", "Content-Range, Content-Length, Accept-Ranges, Content-Disposition")
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	// API Routes
	r.Route("/api", func(r chi.Router) {
		r.Get("/ls", h.List)
		r.Get("/search", h.Search)
		r.Get("/thumb", thumb.GetThumbnail)
		r.Get("/raw/*", h.Raw)
		r.Head("/raw/*", h.Raw)
	})

	// Static Frontend
	workDir, _ := os.Getwd()
	filesDir := filepath.Join(workDir, "public")
	if _, err := os.Stat(filesDir); err == nil {
		r.Handle("/*", http.FileServer(http.Dir(filesDir)))
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s, serving %s", port, libraryPath)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatal(err)
	}
}

func initDB(dbPath, schemaPath string) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		log.Fatal(err)
	}

	// Execute entire schema as one block if possible, or split more carefully
	// modernc.org/sqlite supports multiple statements in one Exec if they don't return rows
	if _, err := db.Exec(string(schema)); err != nil {
		// Ignore errors if table already exists, but log others
		if !strings.Contains(err.Error(), "already exists") {
			log.Printf("Note: Initial DB exec: %v", err)
		}
	}
}
