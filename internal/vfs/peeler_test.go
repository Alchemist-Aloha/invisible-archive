package vfs

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPeelPath(t *testing.T) {
	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "vfs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	// Setup structure:
	// tmpDir/folder/
	// tmpDir/folder/archive.zip
	// tmpDir/folder/v1.0.release.zip
	// tmpDir/direct.zip
	
	folderPath := filepath.Join(tmpDir, "folder")
	os.Mkdir(folderPath, 0755)
	
	archivePath := filepath.Join(folderPath, "archive.zip")
	os.WriteFile(archivePath, []byte("fake zip content"), 0644)

	dottedArchivePath := filepath.Join(folderPath, "v1.0.release.zip")
	os.WriteFile(dottedArchivePath, []byte("fake zip content"), 0644)
	
	directZipPath := filepath.Join(tmpDir, "direct.zip")
	os.WriteFile(directZipPath, []byte("fake zip content"), 0644)

	tests := []struct {
		name         string
		requestPath  string
		wantPhysical string
		wantVirtual  string
		wantIsArchive bool
	}{
		{
			name:         "Real directory",
			requestPath:  "folder",
			wantPhysical: folderPath,
			wantVirtual:  "",
			wantIsArchive: false,
		},
		{
			name:         "Real archive",
			requestPath:  "folder/archive.zip",
			wantPhysical: archivePath,
			wantVirtual:  "",
			wantIsArchive: true,
		},
		{
			name:         "Archive with multiple dots",
			requestPath:  "folder/v1.0.release.zip",
			wantPhysical: dottedArchivePath,
			wantVirtual:  "",
			wantIsArchive: true,
		},
		{
			name:         "Path inside archive with multiple dots",
			requestPath:  "folder/v1.0.release.zip/data/config.json",
			wantPhysical: dottedArchivePath,
			wantVirtual:  filepath.Join("data", "config.json"),
			wantIsArchive: true,
		},
		{
			name:         "Path inside archive",
			requestPath:  "folder/archive.zip/images/cat.jpg",
			wantPhysical: archivePath,
			wantVirtual:  filepath.Join("images", "cat.jpg"),
			wantIsArchive: true,
		},
		{
			name:         "Deep path inside archive",
			requestPath:  "direct.zip/a/b/c/d.txt",
			wantPhysical: directZipPath,
			wantVirtual:  filepath.Join("a", "b", "c", "d.txt"),
			wantIsArchive: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := PeelPath(tmpDir, tt.requestPath)
			if err != nil {
				t.Errorf("PeelPath() error = %v", err)
				return
			}
			if got.PhysicalPath != tt.wantPhysical {
				t.Errorf("PhysicalPath = %v, want %v", got.PhysicalPath, tt.wantPhysical)
			}
			if got.VirtualPath != tt.wantVirtual {
				t.Errorf("VirtualPath = %v, want %v", got.VirtualPath, tt.wantVirtual)
			}
			if got.IsArchive != tt.wantIsArchive {
				t.Errorf("IsArchive = %v, want %v", got.IsArchive, tt.wantIsArchive)
			}
		})
	}
}
