package vfs

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkVFSOpenInner(b *testing.B) {
	tmpDir, err := os.MkdirTemp("", "vfs_bench_test")
	if err != nil {
		b.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	zipPath := filepath.Join(tmpDir, "archive.zip")
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Create 1000 files in inner directory
	for i := 0; i < 1000; i++ {
		f, _ := zw.Create(filepath.Join("inner", string(rune('a'+i%26)), "file.txt"))
		f.Write([]byte("content inside zip"))
	}
	// The target file at the end
	f, _ := zw.Create("inner/target.txt")
	f.Write([]byte("content inside zip"))

	zw.Close()
	os.WriteFile(zipPath, buf.Bytes(), 0644)

	mgr, err := NewManager(tmpDir, 10, nil)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		file, ca, err := mgr.Open("archive.zip/inner/target.txt")
		if err != nil {
			b.Fatal(err)
		}
		file.Close()
		ca.Close()
	}
}
