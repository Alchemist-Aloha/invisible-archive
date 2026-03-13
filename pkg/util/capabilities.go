package util

import (
	"path/filepath"
	"strings"
)

const (
	CapBrowse uint32 = 1 << iota // 1
	CapStream                    // 2
	CapRender                    // 4
	CapEdit                      // 8
)

func GetCapabilities(name string, isDir bool) uint32 {
	var caps uint32
	nameLower := strings.ToLower(name)

	if isDir || strings.HasSuffix(nameLower, ".zip") {
		caps |= CapBrowse
	}

	ext := filepath.Ext(nameLower)
	switch ext {
	case ".mp4", ".mkv", ".webm", ".mp3", ".wav", ".m4v":
		caps |= CapStream
	case ".jpg", ".jpeg", ".png", ".gif", ".webp", ".pdf":
		caps |= CapRender
	case ".txt", ".md", ".go", ".js", ".ts", ".vue", ".css", ".json", ".py", ".cpp", ".h":
		caps |= CapEdit
	}

	return caps
}
