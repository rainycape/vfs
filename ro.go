package vfs

import (
	"errors"
	"fmt"
	"os"
)

var (
	// ErrReadOnlyFileSystem is the error returned by ReadOnlyFileSystem
	// from calls which would result in a write operation.
	ErrReadOnlyFileSystem = errors.New("read-only filesystem")
)

// ReadOnlyFileSystem wraps another VFS and intercepts all the write
// calls, making them return an error. Use function ReadOnly to create
// a ReadOnlyFileSystem.
type ReadOnlyFileSystem struct {
	fs VFS
}

// VFS returns the underlying VFS.
func (fs *ReadOnlyFileSystem) VFS() VFS {
	return fs.fs
}

func (fs *ReadOnlyFileSystem) Open(path string) (RFile, error) {
	return fs.fs.Open(path)
}

func (fs *ReadOnlyFileSystem) OpenFile(path string, flag int, perm os.FileMode) (WFile, error) {
	if flag&(os.O_CREATE|os.O_WRONLY|os.O_RDWR) != 0 {
		return nil, ErrReadOnlyFileSystem
	}
	return fs.fs.OpenFile(path, flag, perm)
}

func (fs *ReadOnlyFileSystem) Lstat(path string) (os.FileInfo, error) {
	return fs.fs.Lstat(path)
}

func (fs *ReadOnlyFileSystem) Stat(path string) (os.FileInfo, error) {
	return fs.fs.Stat(path)
}

func (fs *ReadOnlyFileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	return fs.fs.ReadDir(path)
}

func (fs *ReadOnlyFileSystem) Mkdir(path string, perm os.FileMode) error {
	return ErrReadOnlyFileSystem
}

func (fs *ReadOnlyFileSystem) Remove(path string) error {
	return ErrReadOnlyFileSystem
}

func (fs *ReadOnlyFileSystem) String() string {
	return fmt.Sprintf("RO %s", fs.fs.String())
}

// ReadOnly returns a ReadOnlyFileSystem wrapping the given fs.
func ReadOnly(fs VFS) *ReadOnlyFileSystem {
	return &ReadOnlyFileSystem{fs: fs}
}
