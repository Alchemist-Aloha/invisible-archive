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

	t.Run("List ZIP as directory", func(t *testing.T) {
		files, err := mgr.ReadDir("archive.zip")
		if err != nil {
			t.Errorf("ReadDir(archive.zip) error: %v", err)
		} else {
			found := false
			for _, fi := range files {
				if fi.Name() == "inner" && fi.IsDir() {
					found = true
				}
			}
			if !found {
				t.Error("expected to find 'inner' directory inside zip root")
			}
		}
	})

	t.Run("List inner directory inside ZIP", func(t *testing.T) {
		files, err := mgr.ReadDir("archive.zip/inner")
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
		_, err = reader.Read(data)
		if err != nil {
			t.Fatal(err)
		}
		if string(data) != "content inside zip" {
			t.Errorf("expected 'content inside zip', got '%s'", string(data))
		}
	})
}
