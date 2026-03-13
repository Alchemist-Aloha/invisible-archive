package vfs

import (
	"log"
	"os"
	"path/filepath"
	"strings"
)

// PathResult represents the split path
type PathResult struct {
	PhysicalPath string // Path to the actual file or directory on disk
	VirtualPath  string // Path within the archive (if any)
	IsArchive    bool   // True if PhysicalPath points to a ZIP file
}

// PeelPath identifies where the physical filesystem ends and virtual path begins.
// It follows the "Longest Physical Match" strategy.
func PeelPath(basePath, requestPath string) (*PathResult, error) {
	fullPath := filepath.Join(basePath, filepath.Clean("/"+requestPath))
	log.Printf("VFS: Peeling path. Base: %s, Request: %s, Full: %s", basePath, requestPath, fullPath)
	
	// We start from the longest path and peel back segments
	current := fullPath
	var virtualSegments []string

	for {
		info, err := os.Stat(current)
		if err == nil {
			// Found a physical match
			nameLower := strings.ToLower(current)
			isZip := !info.IsDir() && strings.HasSuffix(nameLower, ".zip")
			
			virtualPath := filepath.Join(virtualSegments...)
			if virtualPath != "" {
				// Don't prepend slash, just keep the joined segments, it's relative
				virtualPath = virtualPath
			}

			res := &PathResult{
				PhysicalPath: current,
				VirtualPath:  virtualPath,
				IsArchive:    isZip,
			}
			log.Printf("VFS: Found physical match: %s (IsArchive: %v, Virtual: %s)", current, isZip, res.VirtualPath)
			return res, nil
		}

		// Peel off one segment
		parent := filepath.Dir(current)
		if parent == current {
			// Reached root without finding a match
			break
		}
		
		virtualSegments = append([]string{filepath.Base(current)}, virtualSegments...)
		current = parent
	}

	log.Printf("VFS: No physical match found for %s", fullPath)
	return nil, os.ErrNotExist
}
