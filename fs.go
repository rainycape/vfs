package vfs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
)

type FileSystem struct {
	root      string
	temporary bool
}

func (fs *FileSystem) path(name string) string {
	name = path.Clean("/" + name)
	return filepath.Join(fs.root, filepath.FromSlash(name))
}

// Root returns the root directory of the FileSystem, as an
// absolute path native to the current operating system.
func (fs *FileSystem) Root() string {
	return fs.root
}

// IsTemporary returns wheter the FileSystem is temporary.
func (fs *FileSystem) IsTemporary() bool {
	return fs.temporary
}

func (fs *FileSystem) Open(path string) (RFile, error) {
	return os.Open(fs.path(path))
}

func (fs *FileSystem) OpenFile(path string, flag int, mode os.FileMode) (WFile, error) {
	return os.OpenFile(fs.path(path), flag, mode)
}

func (fs *FileSystem) Lstat(path string) (os.FileInfo, error) {
	return os.Lstat(fs.path(path))
}

func (fs *FileSystem) Stat(path string) (os.FileInfo, error) {
	return os.Stat(fs.path(path))
}

func (fs *FileSystem) ReadDir(path string) ([]os.FileInfo, error) {
	return ioutil.ReadDir(fs.path(path))
}

func (fs *FileSystem) Mkdir(path string, perm os.FileMode) error {
	return os.Mkdir(fs.path(path), perm)
}

func (fs *FileSystem) Remove(path string) error {
	return os.Remove(fs.path(path))
}

func (fs *FileSystem) String() string {
	return fmt.Sprintf("FileSystem: %s", fs.root)
}

// Close is a no-op on non-temporary filesystems. On temporary
// ones (as returned by TmpFS), it removes all the temporary files.
func (f *FileSystem) Close() error {
	if f.temporary {
		return os.RemoveAll(f.root)
	}
	return nil
}

// FS returns a FileSystem at the given path, which must be provided
// as native path of the current operating system. The path might be
// either absolute or relative, but the FileSystem will be anchored
// at the absolute path represented by root at the time of the function
// call.
func FS(root string) (*FileSystem, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	return &FileSystem{root: abs}, nil
}

// TmpFS returns a temporary FileSystem with the given prefix in its root
// directory name, which might be empty. Once you're done with the temporary
// filesystem, you might remove all its files by calling FileSystem.Close.
func TmpFS(prefix string) (*FileSystem, error) {
	dir, err := ioutil.TempDir("", prefix)
	if err != nil {
		return nil, err
	}
	fs, err := FS(dir)
	if err != nil {
		return nil, err
	}
	fs.temporary = true
	return fs, nil
}

func fsCompileTimeCheck() VFS {
	return &FileSystem{}
}
