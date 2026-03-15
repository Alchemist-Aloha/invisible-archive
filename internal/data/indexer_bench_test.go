package data

import (
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
