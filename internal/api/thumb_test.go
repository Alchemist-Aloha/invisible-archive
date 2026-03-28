package api

import (
	"image"
	"image/color"
	"image/png"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/likun/invisible-archive/internal/vfs"
)

func TestMain(m *testing.M) {
	vips.LoggingSettings(nil, vips.LogLevelError)
	vips.Startup(nil)

	code := m.Run()

	vips.Shutdown()
	os.Exit(code)
}

type thumbTestEnv struct {
	vfsMgr   *vfs.Manager
	cacheDir string
	rootDir  string
	thumb    *Thumbnailer
}

func setupThumbTest(t *testing.T) *thumbTestEnv {
	rootDir, err := os.MkdirTemp("", "thumb-test-root-*")
	if err != nil {
		t.Fatal(err)
	}

	cacheDir, err := os.MkdirTemp("", "thumb-test-cache-*")
	if err != nil {
		t.Fatal(err)
	}

	mgr, err := vfs.NewManager(rootDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	thumb, err := NewThumbnailer(mgr, cacheDir, 2)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		os.RemoveAll(rootDir)
		os.RemoveAll(cacheDir)
	})

	return &thumbTestEnv{
		vfsMgr:   mgr,
		cacheDir: cacheDir,
		rootDir:  rootDir,
		thumb:    thumb,
	}
}

func TestNewThumbnailer(t *testing.T) {
	rootDir, err := os.MkdirTemp("", "thumb-test-root-*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(rootDir)

	cacheDir := rootDir + "/cache"

	mgr, err := vfs.NewManager(rootDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	thumb, err := NewThumbnailer(mgr, cacheDir, 2)
	if err != nil {
		t.Fatalf("Failed to create thumbnailer: %v", err)
	}

	// Verify directory was created
	stat, err := os.Stat(cacheDir)
	if err != nil {
		t.Fatalf("Cache directory was not created: %v", err)
	}
	if !stat.IsDir() {
		t.Errorf("Cache path is not a directory")
	}

	if thumb.vfs != mgr {
		t.Errorf("VFS manager not set correctly")
	}
	if thumb.cacheDir != cacheDir {
		t.Errorf("Cache dir not set correctly")
	}
}

func TestThumbnailer_QueueBackground(t *testing.T) {
	env := setupThumbTest(t)

	// Queue size is 1000, queue 1050 to ensure it doesn't block and drops
	for i := 0; i < 1050; i++ {
		env.thumb.QueueBackground("path/to/test.jpg")
	}

	// Add one valid path and ensure no panic or blocking
	env.thumb.QueueBackground("path/to/valid.jpg")
}

func createTestImage(path string) error {
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for x := 0; x < 100; x++ {
		for y := 0; y < 100; y++ {
			img.Set(x, y, color.RGBA{uint8(x), uint8(y), 255, 255})
		}
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}

func TestThumbnailer_GetThumbnail(t *testing.T) {
	env := setupThumbTest(t)

	// Create test files
	imgName := "test_image.png"
	imgPath := filepath.Join(env.rootDir, imgName)
	if err := createTestImage(imgPath); err != nil {
		t.Fatalf("Failed to create test image: %v", err)
	}

	svgName := "test_icon.svg"
	svgPath := filepath.Join(env.rootDir, svgName)
	svgContent := `<svg width="100" height="100"><circle cx="50" cy="50" r="40" fill="red"/></svg>`
	if err := os.WriteFile(svgPath, []byte(svgContent), 0644); err != nil {
		t.Fatalf("Failed to create test svg: %v", err)
	}

	t.Run("Missing path", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/thumb", nil)
		w := httptest.NewRecorder()

		env.thumb.GetThumbnail(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status Bad Request, got %v", w.Code)
		}
	})

	t.Run("SVG Bypass", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/thumb?path="+svgName, nil)
		w := httptest.NewRecorder()

		env.thumb.GetThumbnail(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status OK, got %v", w.Code)
		}

		if w.Header().Get("Content-Type") != "image/svg+xml" {
			t.Errorf("Expected Content-Type image/svg+xml, got %v", w.Header().Get("Content-Type"))
		}

		if w.Body.String() != svgContent {
			t.Errorf("Expected original SVG body content")
		}
	})

	t.Run("Generate WebP Thumbnail and Retrieve from Cache", func(t *testing.T) {
		// First request to generate the thumbnail
		req1 := httptest.NewRequest("GET", "/api/thumb?path="+imgName, nil)
		w1 := httptest.NewRecorder()

		env.thumb.GetThumbnail(w1, req1)

		if w1.Code != http.StatusOK {
			t.Errorf("Generate: Expected status OK, got %v", w1.Code)
		}

		if w1.Header().Get("Content-Type") != "image/webp" {
			t.Errorf("Generate: Expected Content-Type image/webp, got %v", w1.Header().Get("Content-Type"))
		}

		// Second request should serve from cache
		req2 := httptest.NewRequest("GET", "/api/thumb?path="+imgName, nil)
		w2 := httptest.NewRecorder()

		env.thumb.GetThumbnail(w2, req2)

		if w2.Code != http.StatusOK {
			t.Errorf("Cache: Expected status OK, got %v", w2.Code)
		}

		if w2.Header().Get("Content-Type") != "image/webp" {
			t.Errorf("Cache: Expected Content-Type image/webp, got %v", w2.Header().Get("Content-Type"))
		}
	})
}
