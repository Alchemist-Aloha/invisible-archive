package data

import (
	"path/filepath"
	"strings"
	"testing"
)

var testPaths = []string{
	"file.txt",
	"folder/file.txt",
	"folder/subfolder/file.txt",
	"deep/nested/folder/structure/file.txt",
	"folder_with_slash/",
}

func BenchmarkPathParsingSplit(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			parts := strings.Split(cleanName, "/")
			_ = parts[len(parts)-1]
			if len(parts) > 1 {
				_ = strings.Join(parts[:len(parts)-1], "/")
			}
		}
	}
}

func BenchmarkPathParsingIndexByte(b *testing.B) {
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			slashIdx := strings.LastIndexByte(cleanName, '/')
			var name, parent string
			if slashIdx != -1 {
				name = cleanName[slashIdx+1:]
				parent = cleanName[:slashIdx]
			} else {
				name = cleanName
				parent = ""
			}
			_ = name
			_ = parent
		}
	}
}

func BenchmarkPathJoin(b *testing.B) {
	relZipPath := "photos/2023.zip"
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			slashIdx := strings.LastIndexByte(cleanName, '/')
			var parent string
			if slashIdx != -1 {
				parent = cleanName[:slashIdx]
			} else {
				parent = ""
			}

			_ = "/" + filepath.Join(relZipPath, parent)
			_ = "/" + filepath.Join(relZipPath, cleanName)
		}
	}
}

func BenchmarkPathConcatOptimal(b *testing.B) {
	relZipPath := "photos/2023.zip"
	var baseZipPath string
	if relZipPath == "" || relZipPath == "." {
		baseZipPath = ""
	} else {
		baseZipPath = "/" + relZipPath
	}

	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			slashIdx := strings.LastIndexByte(cleanName, '/')
			var name, parentPath, fullPath string

			if slashIdx != -1 {
				name = cleanName[slashIdx+1:]
				parentPath = baseZipPath + "/" + cleanName[:slashIdx]
			} else {
				name = cleanName
				if baseZipPath == "" {
					parentPath = "/"
				} else {
					parentPath = baseZipPath
				}
			}

			if cleanName == "" {
				fullPath = baseZipPath
				if fullPath == "" {
					fullPath = "/"
				}
			} else {
				if baseZipPath == "" {
					fullPath = "/" + cleanName
				} else {
					fullPath = baseZipPath + "/" + cleanName
				}
			}

			_ = name
			_ = parentPath
			_ = fullPath
		}
	}
}
