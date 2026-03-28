package api

import (
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/likun/invisible-archive/internal/vfs"
)

func TestMain(m *testing.M) {
	// Initialize libvips globally for the package
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)
	code := m.Run()
	vips.Shutdown()
	os.Exit(code)
}

func setupThumbnailerTestEnv(t *testing.T) (*vfs.Manager, string, string) {
	// Temporary lib directory
	libDir, err := os.MkdirTemp("", "thumb-test-lib-*")
	if err != nil {
		t.Fatal(err)
	}

	// Temporary cache directory
	cacheDir, err := os.MkdirTemp("", "thumb-test-cache-*")
	if err != nil {
		t.Fatal(err)
	}

	// Create dummy image (1x1 PNG)
	pngBase64 := "iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAYAAAAfFcSJAAAACklEQVR4nGMAAQAABQABDQottAAAAABJRU5ErkJggg=="
	pngBytes, _ := base64.StdEncoding.DecodeString(pngBase64)
	if err := os.WriteFile(filepath.Join(libDir, "test.png"), pngBytes, 0644); err != nil {
		t.Fatal(err)
	}

	// Create dummy SVG
	svgContent := "<svg></svg>"
	if err := os.WriteFile(filepath.Join(libDir, "test.svg"), []byte(svgContent), 0644); err != nil {
		t.Fatal(err)
	}

	mgr, err := vfs.NewManager(libDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(libDir)
		os.RemoveAll(cacheDir)
	})

	return mgr, cacheDir, libDir
}

func TestNewThumbnailer(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)

	// Test missing directory creation
	newCache := filepath.Join(cacheDir, "new-cache")
	thumb, err := NewThumbnailer(mgr, newCache, 2)
	if err != nil {
		t.Fatalf("Failed to create thumbnailer: %v", err)
	}

	if thumb.cacheDir != newCache {
		t.Errorf("Expected cacheDir %s, got %s", newCache, thumb.cacheDir)
	}

	if _, err := os.Stat(newCache); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestThumbnailer_GenerateInternal_SVG(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)
	thumb, _ := NewThumbnailer(mgr, cacheDir, 2)

	path, err := thumb.generateInternal("test.svg")
	if err != nil {
		t.Fatalf("Expected no error for SVG, got: %v", err)
	}
	if path != "" {
		t.Errorf("Expected empty path for SVG bypass, got: %s", path)
	}
}

func TestThumbnailer_GenerateInternal_Image(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)
	thumb, _ := NewThumbnailer(mgr, cacheDir, 2)

	path, err := thumb.generateInternal("test.png")
	if err != nil {
		t.Fatalf("Failed to generate image: %v", err)
	}
	if path == "" {
		t.Fatal("Expected path for generated thumbnail, got empty")
	}
	if filepath.Ext(path) != ".webp" {
		t.Errorf("Expected .webp extension, got %s", filepath.Ext(path))
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Errorf("Thumbnail file not found at %s", path)
	}
}

func TestThumbnailer_QueueBackground(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)
	thumb, _ := NewThumbnailer(mgr, cacheDir, 2)

	thumb.QueueBackground("test.png")

	// Wait for background worker to process the queue (polling with timeout)
	timeout := time.After(1 * time.Second)
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	found := false
loop:
	for {
		select {
		case <-timeout:
			break loop
		case <-ticker.C:
			// Check if any .webp files exist in cache directory
			files, err := os.ReadDir(cacheDir)
			if err != nil {
				t.Fatal(err)
			}

			for _, f := range files {
				if filepath.Ext(f.Name()) == ".webp" {
					found = true
					break loop
				}
			}
		}
	}

	if !found {
		t.Error("Background queue did not generate thumbnail within timeout")
	}
}

func TestThumbnailer_GetThumbnail_SVG(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)
	thumb, _ := NewThumbnailer(mgr, cacheDir, 2)

	req := httptest.NewRequest("GET", "/thumb?path=test.svg", nil)
	w := httptest.NewRecorder()

	thumb.GetThumbnail(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status OK, got %v", w.Code)
	}
	if w.Header().Get("Content-Type") != "image/svg+xml" {
		t.Errorf("Expected Content-Type image/svg+xml, got %s", w.Header().Get("Content-Type"))
	}
}

func TestThumbnailer_GetThumbnail_MissingPath(t *testing.T) {
	mgr, cacheDir, _ := setupThumbnailerTestEnv(t)
	thumb, _ := NewThumbnailer(mgr, cacheDir, 2)

	req := httptest.NewRequest("GET", "/thumb", nil)
	w := httptest.NewRecorder()

	thumb.GetThumbnail(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status Bad Request, got %v", w.Code)
	}
}
