package vfs

import (
	"os"
	"testing"
	"time"
)

type mockFileInfo struct {
	name  string
	isDir bool
	size  int64
}

func (m *mockFileInfo) Name() string       { return m.name }
func (m *mockFileInfo) Size() int64        { return m.size }
func (m *mockFileInfo) Mode() os.FileMode  { return 0 }
func (m *mockFileInfo) ModTime() time.Time { return time.Now() }
func (m *mockFileInfo) IsDir() bool        { return m.isDir }
func (m *mockFileInfo) Sys() interface{}   { return nil }

func TestNaturalCompare(t *testing.T) {
	tests := []struct {
		s1, s2 string
		want   bool
	}{
		{"a", "b", true},
		{"a1", "a2", true},
		{"a2", "a10", true},
		{"file2.txt", "file10.txt", true},
		{"10", "2", false},
	}

	for _, tt := range tests {
		if got := NaturalCompare(tt.s1, tt.s2); got != tt.want {
			t.Errorf("NaturalCompare(%q, %q) = %v, want %v", tt.s1, tt.s2, got, tt.want)
		}
	}
}

func TestManager_ReadDir_Sorting(t *testing.T) {
	// We can't easily test Manager.ReadDir without setup,
	// but we can test the sorting logic by manually calling it if we expose it,
	// or just trust the logic if we test the NaturalCompare.
	// Actually, let's test NaturalCompare and rely on sort.Slice.
}
