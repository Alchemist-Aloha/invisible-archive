package util

import (
	"testing"
)

func TestGetCapabilities(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		isDir    bool
		expected uint32
	}{
		// Directory tests
		{"Directory", "folder", true, CapBrowse},
		{"Directory with extension", "folder.mp4", true, CapBrowse | CapStream},

		// ZIP tests
		{"ZIP file", "archive.zip", false, CapBrowse},
		{"Uppercase ZIP", "ARCHIVE.ZIP", false, CapBrowse},

		// Stream tests
		{"MP4 file", "video.mp4", false, CapStream},
		{"Uppercase MKV", "video.MKV", false, CapStream},
		{"MP3 file", "audio.mp3", false, CapStream},

		// Render tests
		{"JPG file", "image.jpg", false, CapRender},
		{"Uppercase PNG", "image.PNG", false, CapRender},

		// Edit tests
		{"TXT file", "text.txt", false, CapEdit},
		{"Uppercase GO", "source.GO", false, CapEdit},

		// Unknown extensions
		{"Unknown extension", "file.unknown", false, 0},
		{"No extension", "file", false, 0},

		// Edge cases
		{"Empty string", "", false, 0},
		{"Empty directory", "", true, CapBrowse},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetCapabilities(tt.fileName, tt.isDir)
			if result != tt.expected {
				t.Errorf("GetCapabilities(%q, %v) = %v; want %v", tt.fileName, tt.isDir, result, tt.expected)
			}
		})
	}
}
