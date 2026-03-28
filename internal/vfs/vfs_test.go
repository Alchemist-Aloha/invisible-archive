package vfs

import (
	"archive/zip"
	"bytes"
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/likun/invisible-archive/internal/data"
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
		files, effectivePath, err := mgr.ReadDir("archive.zip", "", false)
		if err != nil {
			t.Errorf("ReadDir(archive.zip) error: %v", err)
		} else {
			// ZIP contains only "inner/" so it should auto-enter it
			if effectivePath != "/archive.zip/inner" {
				t.Errorf("expected effective path '/archive.zip/inner', got '%s'", effectivePath)
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
		files, _, err := mgr.ReadDir("archive.zip/inner", "", false)
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

func TestManager_Search(t *testing.T) {
	// 1. Test when indexer is nil
	t.Run("Nil Indexer", func(t *testing.T) {
		mgr, err := NewManager(".", 10, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = mgr.Search(context.Background(), "anything")
		if err == nil {
			t.Fatal("expected error when indexer is nil, got none")
		}
		if err.Error() != "indexer not initialized" {
			t.Fatalf("expected error 'indexer not initialized', got '%v'", err)
		}
	})

	// 2. Test with actual indexer and sqlite db
	t.Run("Valid Search", func(t *testing.T) {
		dbPath := filepath.Join(t.TempDir(), "test_vfs_search.db")

		// Init DB
		indexer, err := data.NewIndexer(dbPath, ".")
		if err != nil {
			t.Fatalf("failed to init indexer: %v", err)
		}
		defer indexer.Close()

		// Access db and init schema
		db, err := sql.Open("sqlite", "file:"+dbPath)
		if err != nil {
			t.Fatalf("failed to open db: %v", err)
		}
		defer db.Close()

		schemaPath := filepath.Join("..", "data", "schema.sql")
		schema, err := os.ReadFile(schemaPath)
		if err != nil {
			t.Fatalf("failed to read schema: %v", err)
		}

		if _, err := db.Exec(string(schema)); err != nil {
			t.Fatalf("failed to init schema: %v", err)
		}

		mgr, err := NewManager(".", 10, indexer)
		if err != nil {
			t.Fatalf("failed to create manager: %v", err)
		}

		// Insert mock items
		err = indexer.GetQueries().UpsertItem(context.Background(), data.UpsertItemParams{
			ParentPath:   "root",
			Name:         "testfile.txt",
			Path:         "root/testfile.txt",
			IsDir:        false,
			Size:         100,
			ModTime:      12345,
			Capabilities: 0,
			IsInsideZip:  false,
		})
		if err != nil {
			t.Fatalf("failed to insert item 1: %v", err)
		}

		err = indexer.GetQueries().UpsertItem(context.Background(), data.UpsertItemParams{
			ParentPath:   "root/sub",
			Name:         "another.png",
			Path:         "root/sub/another.png",
			IsDir:        false,
			Size:         200,
			ModTime:      12345,
			Capabilities: 0,
			IsInsideZip:  false,
		})
		if err != nil {
			t.Fatalf("failed to insert item 2: %v", err)
		}

		// Test search by name match
		res, err := mgr.Search(context.Background(), "%testfile%")
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 result, got %d", len(res))
		}
		if res[0].Name != "testfile.txt" {
			t.Errorf("expected result to be 'testfile.txt', got '%s'", res[0].Name)
		}

		// Test search by path match
		res, err = mgr.Search(context.Background(), "%sub/another%")
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(res) != 1 {
			t.Fatalf("expected 1 result, got %d", len(res))
		}
		if res[0].Name != "another.png" {
			t.Errorf("expected result to be 'another.png', got '%s'", res[0].Name)
		}

		// Test search no match
		res, err = mgr.Search(context.Background(), "%nonexistent%")
		if err != nil {
			t.Fatalf("search failed: %v", err)
		}
		if len(res) != 0 {
			t.Fatalf("expected 0 results, got %d", len(res))
		}
	})
}
