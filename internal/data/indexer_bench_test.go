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
	relZipPath := "test/archive.zip"
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			slashIdx := strings.LastIndexByte(cleanName, '/')
			var parentInZip string
			if slashIdx != -1 {
				parentInZip = cleanName[:slashIdx]
			} else {
				parentInZip = ""
			}

			_ = "/" + filepath.Join(relZipPath, parentInZip)
			_ = "/" + filepath.Join(relZipPath, path)
		}
	}
}

func BenchmarkPathConcat(b *testing.B) {
	relZipPath := "test/archive.zip"
	for i := 0; i < b.N; i++ {
		for _, path := range testPaths {
			cleanName := strings.TrimSuffix(path, "/")
			slashIdx := strings.LastIndexByte(cleanName, '/')
			var parentInZip string
			if slashIdx != -1 {
				parentInZip = cleanName[:slashIdx]
			} else {
				parentInZip = ""
			}

			var parentPath string
			if parentInZip == "" {
				parentPath = "/" + relZipPath
			} else {
				parentPath = "/" + relZipPath + "/" + parentInZip
			}
			fullPath := "/" + relZipPath + "/" + cleanName

			_ = parentPath
			_ = fullPath
		}
	}
}
