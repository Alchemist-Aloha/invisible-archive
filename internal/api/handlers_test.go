package api

import (
	"archive/zip"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/likun/invisible-archive/internal/vfs"
)

func TestRawHandlerSpecialCharacters(t *testing.T) {
	// Setup temporary library
	tempDir, err := os.MkdirTemp("", "archive-test-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	fileName := "video [2026].txt"
	filePath := filepath.Join(tempDir, fileName)
	content := "test content"
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Initialize VFS Manager
	mgr, err := vfs.NewManager(tempDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := NewHandler(mgr)
	r := chi.NewRouter()
	r.Get("/raw/*", h.Raw)

	// Test with encoded path
	encodedPath := "/raw/video%20%5B2026%5D.txt"
	req := httptest.NewRequest("GET", encodedPath, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if w.Body.String() != content {
		t.Errorf("Expected body %q, got %q", content, w.Body.String())
	}
}

func TestRawHandlerSpecialCharactersInZip(t *testing.T) {
	// Setup temporary library with a ZIP
	tempDir, err := os.MkdirTemp("", "archive-test-zip-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a real zip file for testing
	zipPath := filepath.Join(tempDir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	
	innerFileName := "inner [video].txt"
	content := "inner content"
	zf, err := zw.Create(innerFileName)
	if err != nil {
		t.Fatal(err)
	}
	zf.Write([]byte(content))
	zw.Close()
	f.Close()

	// Initialize VFS Manager
	mgr, err := vfs.NewManager(tempDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := NewHandler(mgr)
	r := chi.NewRouter()
	r.Get("/raw/*", h.Raw)

	// Test with encoded path inside zip
	encodedPath := "/raw/test.zip/inner%20%5Bvideo%5D.txt"
	req := httptest.NewRequest("GET", encodedPath, nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}

	if w.Body.String() != content {
		t.Errorf("Expected body %q, got %q", content, w.Body.String())
	}
}
