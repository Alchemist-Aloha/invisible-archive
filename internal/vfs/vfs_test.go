package vfs

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestManager(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vfs_manager_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a real ZIP with a file inside
	zipPath := filepath.Join(tmpDir, "archive.zip")
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	// Explicitly create directory entry
	_, _ = zw.Create("inner/")
	f, _ := zw.Create("inner/hello.txt")
	f.Write([]byte("content inside zip"))
	zw.Close()
	os.WriteFile(zipPath, buf.Bytes(), 0644)

	mgr, err := NewManager(tmpDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("List ZIP as directory (Auto-enter single folder)", func(t *testing.T) {
		files, effectivePath, err := mgr.ReadDir("archive.zip")
		if err != nil {
			t.Errorf("ReadDir(archive.zip) error: %v", err)
		} else {
			// ZIP contains only "inner/" so it should auto-enter it
			if effectivePath != "archive.zip/inner" {
				t.Errorf("expected effective path 'archive.zip/inner', got '%s'", effectivePath)
			}
			found := false
			for _, fi := range files {
				if fi.Name() == "hello.txt" {
					found = true
				}
			}
			if !found {
				t.Error("expected to find 'hello.txt' inside zip root (via auto-enter)")
			}
		}
	})

	t.Run("List inner directory inside ZIP", func(t *testing.T) {
		files, _, err := mgr.ReadDir("archive.zip/inner")
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, fi := range files {
			if fi.Name() == "hello.txt" {
				found = true
			}
		}
		if !found {
			t.Error("expected to find 'hello.txt' inside 'inner' directory")
		}
	})

	t.Run("Read file inside ZIP", func(t *testing.T) {
		reader, closer, err := mgr.GetRawReader("archive.zip/inner/hello.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer closer.Close()

		data := make([]byte, 18)
		n, err := reader.Read(data)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(data[:n]) != "content inside zip" {
			t.Errorf("expected 'content inside zip', got '%s'", string(data[:n]))
		}
	})

	t.Run("Seek inside ZIP", func(t *testing.T) {
		reader, closer, err := mgr.GetRawReader("archive.zip/inner/hello.txt")
		if err != nil {
			t.Fatal(err)
		}
		defer closer.Close()

		// Read first 7 bytes ("content")
		data := make([]byte, 7)
		n, err := reader.Read(data)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(data[:n]) != "content" {
			t.Errorf("expected 'content', got '%s'", string(data[:n]))
		}

		// Seek to "zip" (offset 15)
		newOffset, err := reader.Seek(15, 0) // SeekStart
		if err != nil {
			t.Fatal(err)
		}
		if newOffset != 15 {
			t.Errorf("expected offset 15, got %d", newOffset)
		}

		data = make([]byte, 3)
		n, err = reader.Read(data)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(data[:n]) != "zip" {
			t.Errorf("expected 'zip', got '%s'", string(data[:n]))
		}

		// Seek back to "inside" (offset 8)
		newOffset, err = reader.Seek(8, 0) // SeekStart
		if err != nil {
			t.Fatal(err)
		}
		if newOffset != 8 {
			t.Errorf("expected offset 8, got %d", newOffset)
		}

		data = make([]byte, 6)
		n, err = reader.Read(data)
		if err != nil && err.Error() != "EOF" {
			t.Fatalf("unexpected err: %v", err)
		}
		if string(data[:n]) != "inside" {
			t.Errorf("expected 'inside', got '%s'", string(data[:n]))
		}

		// Seek to end
		newOffset, err = reader.Seek(0, 2) // SeekEnd
		if err != nil {
			t.Fatal(err)
		}
		if newOffset != 18 {
			t.Errorf("expected offset 18, got %d", newOffset)
		}
	})
}
