package vfs

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func createTestZip(t *testing.T, path string) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	f, err := w.Create("test.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, err = f.Write([]byte("hello world"))
	if err != nil {
		t.Fatal(err)
	}
	err = w.Close()
	if err != nil {
		t.Fatal(err)
	}
	err = os.WriteFile(path, buf.Bytes(), 0644)
	if err != nil {
		t.Fatal(err)
	}
}

func TestMountTable(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "mount_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zip1 := filepath.Join(tmpDir, "1.zip")
	zip2 := filepath.Join(tmpDir, "2.zip")
	createTestZip(t, zip1)
	createTestZip(t, zip2)

	// Create table with size 1 to force eviction
	mt, err := NewMountTable(1)
	if err != nil {
		t.Fatal(err)
	}

	// 1. Get zip1
	ca1, err := mt.Get(zip1)
	if err != nil {
		t.Fatal(err)
	}
	if ca1.refs != 1 {
		t.Errorf("expected 1 ref, got %d", ca1.refs)
	}

	// 2. Get zip2 (should evict zip1)
	ca2, err := mt.Get(zip2)
	if err != nil {
		t.Fatal(err)
	}

	if !ca1.evicted {
		t.Error("expected ca1 to be evicted")
	}

	// ca1 should still be open because we haven't closed our handle
	_, err = ca1.Reader.Open("test.txt")
	if err != nil {
		t.Errorf("failed to read from evicted archive with active ref: %v", err)
	}

	// 3. Close ca1
	ca1.Close()
	if ca1.refs != 0 {
		t.Errorf("expected 0 refs, got %d", ca1.refs)
	}

	// Now ca1 should be closed (internal Reader.Close called)
	// Trying to open a file now should fail or behavior is undefined,
	// but we've verified the logic path.

	// 4. Get zip2 again (cache hit)
	ca2_hit, err := mt.Get(zip2)
	if err != nil {
		t.Fatal(err)
	}
	if ca2_hit != ca2 {
		t.Error("expected cache hit for zip2")
	}
	if ca2.refs != 2 {
		t.Errorf("expected 2 refs for ca2, got %d", ca2.refs)
	}

	ca2.Close()
	ca2_hit.Close()
}
