package vfs

import (
	"archive/zip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/likun/invisible-archive/internal/data"
	"github.com/spf13/afero"
)

// Manager coordinates path peeling and archive mounting
type Manager struct {
	basePath   string
	mountTable *MountTable
	osFs       afero.Fs
	indexer    *data.Indexer
}

// NewManager creates a new VFS Manager
func NewManager(basePath string, cacheSize int, indexer *data.Indexer) (*Manager, error) {
	mt, err := NewMountTable(cacheSize)
	if err != nil {
		return nil, err
	}

	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, err
	}

	return &Manager{
		basePath:   absPath,
		mountTable: mt,
		osFs:       afero.NewReadOnlyFs(afero.NewBasePathFs(afero.NewOsFs(), absPath)),
		indexer:    indexer,
	}, nil
}

// Open resolves the path and returns an afero.File
func (m *Manager) Open(path string) (afero.File, *CachedArchive, error) {
	res, err := PeelPath(m.basePath, path)
	if err != nil {
		return nil, nil, err
	}

	if !res.IsArchive {
		relPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
		f, err := m.osFs.Open(relPath)
		return f, nil, err
	}

	ca, err := m.mountTable.Get(res.PhysicalPath)
	if err != nil {
		return nil, nil, err
	}

	vPath := strings.Trim(filepath.ToSlash(res.VirtualPath), "/")
	if vPath == "" {
		return &zipFile{name: "/", isDir: true, ca: ca}, ca, nil
	}

	for _, f := range ca.Reader.File {
		fPath := strings.Trim(f.Name, "/")
		if fPath == vPath {
			rc, err := f.Open()
			if err != nil {
				ca.Close()
				return nil, nil, err
			}
			return &zipFile{
				name:  filepath.Base(fPath),
				isDir: f.FileInfo().IsDir(),
				rc:    rc,
				ca:    ca,
				file:  f,
			}, ca, nil
		}
		if strings.HasPrefix(fPath, vPath+"/") {
			return &zipFile{name: filepath.Base(vPath), isDir: true, ca: ca}, ca, nil
		}
	}

	ca.Close()
	return nil, nil, os.ErrNotExist
}

// Stat returns FileInfo for a given path
func (m *Manager) Stat(path string) (os.FileInfo, error) {
	f, ca, err := m.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if ca != nil {
			ca.Close()
		}
		f.Close()
	}()
	return f.Stat()
}

// ReadDir returns a list of directory entries and the effective path (might be deeper for ZIPs)
func (m *Manager) ReadDir(path string) ([]os.FileInfo, string, error) {
	res, err := PeelPath(m.basePath, path)
	if err != nil {
		return nil, "", err
	}

	if !res.IsArchive {
		relPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
		if m.indexer != nil {
			go m.indexer.IndexDirectory(context.Background(), res.PhysicalPath)
		}
		items, err := afero.ReadDir(m.osFs, relPath)
		return items, path, err
	}

	ca, err := m.mountTable.Get(res.PhysicalPath)
	if err != nil {
		return nil, "", err
	}
	defer ca.Close()

	if m.indexer != nil {
		relZipPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
		go m.indexer.IndexZip(context.Background(), res.PhysicalPath, relZipPath)
	}

	// Auto-enter logic
	effectiveVPath := res.VirtualPath
	for {
		items, _ := m.readZipDir(ca.Reader, effectiveVPath)
		if len(items) == 1 && items[0].IsDir() && res.VirtualPath == "" {
			// Only auto-enter if we are at the root of the ZIP and there is exactly one folder
			effectiveVPath = filepath.ToSlash(filepath.Join(effectiveVPath, items[0].Name()))
			continue
		}
		break
	}

	finalItems, err := m.readZipDir(ca.Reader, effectiveVPath)
	relZipPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
	effectivePath := "/" + filepath.ToSlash(filepath.Join(relZipPath, effectiveVPath))
	return finalItems, effectivePath, err
}

// parseZipEntryName extracts the first path component and whether it should be
// treated as a directory. If no further path separator is found, it returns
// the full path and uses fallbackIsDir to determine directory-ness.
func parseZipEntryName(path string, fallbackIsDir bool) (name string, isDir bool) {
	idx := strings.IndexByte(path, '/')
	if idx >= 0 {
		return path[:idx], true
	}
	return path, fallbackIsDir
}

func (m *Manager) readZipDir(r *zip.ReadCloser, vPath string) ([]os.FileInfo, error) {
	vPath = strings.Trim(filepath.ToSlash(vPath), "/")
	seen := make(map[string]bool)
	var items []os.FileInfo

	for _, f := range r.File {
		fPath := strings.Trim(f.Name, "/")
		var name string
		var isDir bool

		// Use IndexByte instead of Split for a zero-allocation fast path
		if vPath == "" {
			name, isDir = parseZipEntryName(fPath, f.FileInfo().IsDir())
		} else if strings.HasPrefix(fPath, vPath+"/") {
			subPath := strings.TrimPrefix(fPath, vPath+"/")
			name, isDir = parseZipEntryName(subPath, f.FileInfo().IsDir())
		} else {
			continue
		}

		if !seen[name] {
			seen[name] = true
			items = append(items, &zipFileInfo{name: name, isDir: isDir, file: f})
		}
	}

	return items, nil
}

type zipFileInfo struct {
	name  string
	isDir bool
	file  *zip.File
}

func (z *zipFileInfo) Name() string       { return z.name }
func (z *zipFileInfo) Size() int64        { if z.isDir { return 0 }; return int64(z.file.UncompressedSize64) }
func (z *zipFileInfo) Mode() os.FileMode  { if z.isDir { return os.ModeDir | 0555 }; return 0444 }
func (z *zipFileInfo) ModTime() time.Time { return z.file.Modified }
func (z *zipFileInfo) IsDir() bool        { return z.isDir }
func (z *zipFileInfo) Sys() interface{}   { return nil }

func (m *Manager) Search(ctx context.Context, pattern string) ([]data.Item, error) {
	if m.indexer == nil {
		return nil, fmt.Errorf("indexer not initialized")
	}
	return m.indexer.GetQueries().SearchItems(ctx, data.SearchItemsParams{
		Name: pattern,
		Path: pattern,
	})
}

func (m *Manager) GetIndexer() *data.Indexer {
	return m.indexer
}

func (m *Manager) Random(ctx context.Context, pathPrefix string, limit int) ([]data.Item, error) {
	if m.indexer == nil {
		return nil, fmt.Errorf("indexer not initialized")
	}
	if !strings.HasPrefix(pathPrefix, "/") {
		pathPrefix = "/" + pathPrefix
	}
	return m.indexer.GetQueries().RandomItemsByPathPrefix(ctx, data.RandomItemsByPathPrefixParams{
		PathPrefix: sql.NullString{String: pathPrefix, Valid: true},
		Limit:      int64(limit),
	})
}

func (m *Manager) GetRawReader(path string) (io.ReadSeeker, io.Closer, error) {
	res, err := PeelPath(m.basePath, path)
	if err != nil {
		return nil, nil, err
	}

	if !res.IsArchive {
		relPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
		f, err := m.osFs.Open(relPath)
		return f, f, err
	}

	ca, err := m.mountTable.Get(res.PhysicalPath)
	if err != nil {
		return nil, nil, err
	}

	vPath := strings.Trim(filepath.ToSlash(res.VirtualPath), "/")
	for _, f := range ca.Reader.File {
		if strings.Trim(f.Name, "/") == vPath {
			if f.Method == zip.Store {
				relPath, _ := filepath.Rel(m.basePath, res.PhysicalPath)
				osFile, err := m.osFs.Open(relPath)
				if err != nil {
					ca.Close()
					return nil, nil, err
				}

				off, err := f.DataOffset()
				if err != nil {
					osFile.Close()
					ca.Close()
					return nil, nil, err
				}

				ss := &storeStreamSeeker{
					sr: io.NewSectionReader(osFile, off, int64(f.UncompressedSize64)),
					osFile: osFile,
					ca: ca,
				}
				return ss, ss, nil
			}

			zs := &zipStreamSeeker{
				f:  f,
				ca: ca,
			}
			return zs, zs, nil
		}
	}

	ca.Close()
	return nil, nil, os.ErrNotExist
}

type storeStreamSeeker struct {
	sr     *io.SectionReader
	osFile afero.File
	ca     *CachedArchive
}

func (s *storeStreamSeeker) Read(p []byte) (n int, err error) {
	return s.sr.Read(p)
}

func (s *storeStreamSeeker) Seek(offset int64, whence int) (int64, error) {
	return s.sr.Seek(offset, whence)
}

func (s *storeStreamSeeker) Close() error {
	var err error
	if s.osFile != nil {
		err = s.osFile.Close()
	}
	if caErr := s.ca.Close(); caErr != nil && err == nil {
		err = caErr
	}
	return err
}

type zipStreamSeeker struct {
	f      *zip.File
	ca     *CachedArchive
	rc     io.ReadCloser
	offset int64
}

func (z *zipStreamSeeker) Read(p []byte) (n int, err error) {
	if z.rc == nil {
		rc, err := z.f.Open()
		if err != nil {
			return 0, err
		}
		z.rc = rc
		if z.offset > 0 {
			if _, err := io.CopyN(io.Discard, z.rc, z.offset); err != nil {
				z.rc.Close()
				z.rc = nil
				return 0, err
			}
		}
	}
	n, err = z.rc.Read(p)
	z.offset += int64(n)
	return n, err
}

func (z *zipStreamSeeker) Seek(offset int64, whence int) (int64, error) {
	var newOffset int64
	switch whence {
	case io.SeekStart:
		newOffset = offset
	case io.SeekCurrent:
		newOffset = z.offset + offset
	case io.SeekEnd:
		newOffset = int64(z.f.UncompressedSize64) + offset
	default:
		return 0, fmt.Errorf("invalid whence: %d", whence)
	}

	if newOffset < 0 {
		return 0, fmt.Errorf("negative seek offset")
	}

	if newOffset == z.offset {
		return newOffset, nil
	}

	// Optimization: if we are seeking forward and have an open reader, we can just discard bytes
	if newOffset > z.offset && z.rc != nil {
		discard := newOffset - z.offset
		n, err := io.CopyN(io.Discard, z.rc, discard)
		z.offset += n
		if err != nil {
			return z.offset, err
		}
		return z.offset, nil
	}

	z.offset = newOffset
	if z.rc != nil {
		z.rc.Close()
		z.rc = nil
	}

	return z.offset, nil
}

func (z *zipStreamSeeker) Close() error {
	var err error
	if z.rc != nil {
		err = z.rc.Close()
	}
	if caErr := z.ca.Close(); caErr != nil && err == nil {
		err = caErr
	}
	return err
}

type zipFile struct {
	name  string
	isDir bool
	rc    io.ReadCloser
	ca    *CachedArchive
	file  *zip.File
	afero.File
}

func (z *zipFile) Name() string { return z.name }
func (z *zipFile) Close() error {
	if z.rc != nil {
		return z.rc.Close()
	}
	return nil
}
func (z *zipFile) Read(p []byte) (n int, err error) {
	if z.isDir {
		return 0, fmt.Errorf("is a directory")
	}
	return z.rc.Read(p)
}
func (z *zipFile) Stat() (os.FileInfo, error) {
	if z.isDir {
		return &zipFileInfo{name: z.name, isDir: true, file: z.file}, nil
	}
	return z.file.FileInfo(), nil
}
func (z *zipFile) Readdir(count int) ([]os.FileInfo, error) {
	return nil, fmt.Errorf("use Manager.ReadDir")
}
