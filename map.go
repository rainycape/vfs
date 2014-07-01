package vfs

import (
	"fmt"
	"os"
	pathpkg "path"
	"sort"
	"strings"
	"sync"
	"time"
)

type MapFileSystem struct {
	mu    sync.RWMutex
	files map[string]*File
}

func (fs *MapFileSystem) file(path string) (*File, error) {
	fs.mu.RLock()
	f := fs.files[path]
	fs.mu.RUnlock()
	if f == nil {
		return nil, os.ErrNotExist
	}
	return f, nil
}

// dir must always be called with the lock held
func (fs *MapFileSystem) dir(p string) (*File, string, error) {
	dir := pathpkg.Dir(p)
	dir = strings.TrimSuffix(dir, "/")
	if dir == "." {
		dir = ""
	}
	f := fs.files[dir]
	if f == nil {
		return nil, "", os.ErrNotExist
	}
	if f.Mode&os.ModeDir == 0 {
		return nil, "", fmt.Errorf("%s is not a directory", dir)
	}
	return f, dir, nil
}

func (fs *MapFileSystem) sanitize(path string) string {
	path = pathpkg.Clean("/" + path)
	return strings.Trim(path, "/")
}

func (fs *MapFileSystem) Open(path string) (RFile, error) {
	path = fs.sanitize(path)
	f, err := fs.file(path)
	if err != nil {
		return nil, err
	}
	return &file{f: f, readable: true}, nil
}

func (fs *MapFileSystem) OpenFile(path string, flag int, mode os.FileMode) (WFile, error) {
	path = fs.sanitize(path)
	if mode&os.ModeType != 0 {
		return nil, fmt.Errorf("%T does not support special files", fs)
	}
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	f := fs.files[path]
	if f == nil && flag&os.O_CREATE == 0 {
		return nil, os.ErrNotExist
	}
	// Read only file?
	if flag&os.O_WRONLY == 0 && flag&os.O_RDWR == 0 {
		if f == nil {
			return nil, os.ErrNotExist
		}
		return &file{f: f, readable: true}, nil
	}
	// Write file, either f != nil or flag&os.O_CREATE
	if _, _, err := fs.dir(path); err != nil {
		// No parent dir
		return nil, os.ErrNotExist
	}
	if f != nil {
		if flag&os.O_EXCL != 0 {
			return nil, os.ErrExist
		}
		if f.Mode&os.ModeDir != 0 {
			return nil, fmt.Errorf("%s is a directory", path)
		}
	} else {
		f = &File{ModTime: time.Now()}
		// Balance with the deferred RUnlock()
		fs.mu.RUnlock()
		fs.mu.Lock()
		fs.files[path] = f
		fs.mu.Unlock()
		fs.mu.RLock()
	}
	// Check if we should truncate
	if flag&os.O_TRUNC != 0 {
		f.Lock()
		f.ModTime = time.Now()
		f.Data = nil
		f.Unlock()
	}
	return &file{f: f, readable: (flag&os.O_RDWR != 0), writable: true}, nil
}

func (fs *MapFileSystem) Lstat(path string) (os.FileInfo, error) {
	return fs.Stat(path)
}

func (fs *MapFileSystem) Stat(path string) (os.FileInfo, error) {
	path = fs.sanitize(path)
	f, err := fs.file(path)
	if err != nil {
		return nil, err
	}
	return &FileInfo{Path: path, File: f}, nil
}

func (fs *MapFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()
	return fs.readDir(path)
}

func (fs *MapFileSystem) readDir(path string) ([]os.FileInfo, error) {
	path = fs.sanitize(path)
	d := fs.files[path]
	if d == nil {
		return nil, os.ErrNotExist
	}
	if d.Mode&os.ModeDir == 0 {
		return nil, fmt.Errorf("%s is not a directory", path)
	}
	var infos []os.FileInfo
	for k, v := range fs.files {
		if v == d {
			continue
		}
		if dir, _, _ := fs.dir(k); dir == d {
			infos = append(infos, &FileInfo{Path: k, File: v})
		}
	}
	sort.Sort(FileInfos(infos))
	return infos, nil
}

func (fs *MapFileSystem) Mkdir(path string, perm os.FileMode) error {
	path = fs.sanitize(path)
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if f, ok := fs.files[path]; ok {
		if f.Mode&os.ModeDir != 0 {
			return os.ErrExist
		}
		return fmt.Errorf("%s is a file, can't create a directory with the same name", path)
	}
	// Check if parent exists
	if path != "" {
		_, _, err := fs.dir(path)
		if err != nil {
			return os.ErrNotExist
		}
	}
	fs.files[path] = &File{Mode: perm | os.ModeDir}
	return nil
}

func (fs *MapFileSystem) Remove(path string) error {
	path = fs.sanitize(path)
	f, _ := fs.file(path)
	if f == nil {
		return os.ErrNotExist
	}
	fs.mu.Lock()
	defer fs.mu.Unlock()
	if f.Mode&os.ModeDir != 0 {
		infos, err := fs.readDir(path)
		if err != nil {
			return err
		}
		if len(infos) > 0 {
			return fmt.Errorf("directory %s not empty", path)
		}
	}
	delete(fs.files, path)
	return nil
}

func (fs *MapFileSystem) String() string {
	return fmt.Sprintf("MapFileSystem: %d files", len(fs.files))
}

// Map returns a MapFileSystem using the given files argument to
// prepopulate it (which might be nil). Note that the files map does
// not need to contain any directories, they will be created automatically.
// If the files contain conflicting paths (e.g. files named a and a/b, thus
// making "a" both a file and a directory), an error will be returned.
func Map(files map[string]*File) (*MapFileSystem, error) {
	cpy := make(map[string]*File, len(files))
	for k, v := range files {
		cpy[k] = v
	}
	fs := &MapFileSystem{files: cpy}
	if err := fs.Mkdir("/", 0755); err != nil {
		return nil, err
	}
	for k := range files {
		if err := MkdirAll(fs, pathpkg.Dir(k), 0755); err != nil {
			return nil, err
		}
	}
	return fs, nil
}

func mapCompileTimeCheck() VFS {
	return &MapFileSystem{}
}
