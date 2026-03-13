package api

import (
	"archive/zip"
	"encoding/json"
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

func TestListHandlerAutoEnterZip(t *testing.T) {
	// Setup temporary library with a ZIP
	tempDir, err := os.MkdirTemp("", "archive-test-list-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	// Create a zip file with single folder root: test.zip/folder/file.txt
	zipPath := filepath.Join(tempDir, "test.zip")
	f, err := os.Create(zipPath)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(f)
	zf, _ := zw.Create("folder/file.txt")
	zf.Write([]byte("content"))
	zw.Close()
	f.Close()

	// Initialize VFS Manager
	mgr, err := vfs.NewManager(tempDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	h := NewHandler(mgr)
	r := chi.NewRouter()
	r.Get("/ls", h.List)

	// List the zip root
	req := httptest.NewRequest("GET", "/ls?path=test.zip", nil)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %v", w.Code)
	}

	// Verify it auto-entered 'folder'
	var resp ListResponse
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if resp.EffectivePath != "test.zip/folder" {
		t.Errorf("Expected effective path 'test.zip/folder', got '%s'", resp.EffectivePath)
	}

	if len(resp.Items) != 1 || resp.Items[0].Name != "file.txt" {
		t.Errorf("Expected 1 item 'file.txt', got %v", resp.Items)
	}
}
