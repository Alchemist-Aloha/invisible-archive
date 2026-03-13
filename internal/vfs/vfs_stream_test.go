package vfs

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestStoreStreamSeeker(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "vfs_test_*")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "test.zip")

	// Create a zip with one Store file
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)

	f, err := w.CreateHeader(&zip.FileHeader{
		Name:   "video.mp4",
		Method: zip.Store,
	})
	if err != nil {
		t.Fatal(err)
	}

	data := []byte("0123456789")
	f.Write(data)
	w.Close()

	if err := os.WriteFile(zipPath, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	manager, err := NewManager(tmpDir, 10, nil)
	if err != nil {
		t.Fatal(err)
	}

	reader, closer, err := manager.GetRawReader("test.zip/video.mp4")
	if err != nil {
		t.Fatal(err)
	}
	defer closer.Close()

	// Type assertion to ensure it's storeStreamSeeker
	_, isStoreSeeker := reader.(*storeStreamSeeker)
	if !isStoreSeeker {
		t.Fatalf("expected storeStreamSeeker, got %T", reader)
	}

	seeker := reader.(io.ReadSeeker)

	// Seek forward
	off, err := seeker.Seek(5, io.SeekStart)
	if err != nil || off != 5 {
		t.Fatalf("expected offset 5, got %d (err: %v)", off, err)
	}

	d := make([]byte, 5)
	n, err := seeker.Read(d)
	if err != nil || n != 5 {
		t.Fatalf("expected to read 5 bytes, got %d (err: %v)", n, err)
	}
	if string(d) != "56789" {
		t.Fatalf("expected '56789', got '%s'", string(d))
	}

	// Seek backward
	off, err = seeker.Seek(2, io.SeekStart)
	if err != nil || off != 2 {
		t.Fatalf("expected offset 2, got %d (err: %v)", off, err)
	}

	d = make([]byte, 3)
	n, err = seeker.Read(d)
	if err != nil || n != 3 {
		t.Fatalf("expected to read 3 bytes, got %d (err: %v)", n, err)
	}
	if string(d) != "234" {
		t.Fatalf("expected '234', got '%s'", string(d))
	}
}
