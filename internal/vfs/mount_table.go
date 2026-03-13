package vfs

import (
	"archive/zip"
	"fmt"
	"sync"
	"sync/atomic"

	lru "github.com/hashicorp/golang-lru/v2"
)

// CachedArchive wraps a zip.ReadCloser with reference counting
type CachedArchive struct {
	Reader     *zip.ReadCloser
	Path       string
	refs       int32 // Atomic reference counter
	evicted    bool  // True if removed from LRU cache
	mu         sync.Mutex
}

// Close decrements the reference count and closes the file if needed
func (ca *CachedArchive) Close() error {
	newRefs := atomic.AddInt32(&ca.refs, -1)
	if newRefs == 0 {
		ca.mu.Lock()
		defer ca.mu.Unlock()
		if ca.evicted {
			return ca.Reader.Close()
		}
	}
	return nil
}

// Acquire increments the reference count
func (ca *CachedArchive) Acquire() {
	atomic.AddInt32(&ca.refs, 1)
}

// MountTable manages open archives
type MountTable struct {
	cache *lru.Cache[string, *CachedArchive]
}

// NewMountTable creates a new MountTable with the specified size
func NewMountTable(size int) (*MountTable, error) {
	// onEvict is called when an item is removed from the LRU
	onEvict := func(key string, ca *CachedArchive) {
		ca.mu.Lock()
		ca.evicted = true
		ca.mu.Unlock()

		// If no one is using it, close it now
		if atomic.LoadInt32(&ca.refs) == 0 {
			ca.Reader.Close()
		}
	}

	cache, err := lru.NewWithEvict(size, onEvict)
	if err != nil {
		return nil, err
	}

	return &MountTable{cache: cache}, nil
}

// Get returns an existing archive or opens a new one
func (mt *MountTable) Get(path string) (*CachedArchive, error) {
	if ca, ok := mt.cache.Get(path); ok {
		ca.Acquire()
		return ca, nil
	}

	// Cache miss, open the archive
	rc, err := zip.OpenReader(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open zip: %w", err)
	}

	ca := &CachedArchive{
		Reader: rc,
		Path:   path,
		refs:   1, // Initial reference for the caller
	}

	mt.cache.Add(path, ca)
	return ca, nil
}
