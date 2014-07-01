package vfs

import (
	"os"
	"path"
	"sync"
	"time"
)

// Type File represents an in-memory file or
// directory. Most in-memory VFS should use this
// structure to represent their files, in order to
// save work.
type File struct {
	sync.RWMutex
	// Data contains the file data. For directories, it's empty.
	Data []byte
	// Mode is the file or directory mode. Note that some filesystems
	// might ignore the permission bits.
	Mode os.FileMode
	// ModTime represents the last modification time to the file or
	// directory.
	ModTime time.Time
}

// FileInfo implements the os.FileInfo interface wrapping
// a given File and its Path in its VFS.
type FileInfo struct {
	// Path is the full path to the file in its VFS.
	Path string
	// File is the *File instance used by the VFS to represent
	// the in-memory file.
	File *File
}

func (info *FileInfo) Name() string {
	return path.Base(info.Path)
}

func (info *FileInfo) Size() int64 {
	info.File.RLock()
	defer info.File.RUnlock()
	return int64(len(info.File.Data))
}

func (info *FileInfo) Mode() os.FileMode {
	info.File.RLock()
	defer info.File.RUnlock()
	return info.File.Mode
}

func (info *FileInfo) ModTime() time.Time {
	info.File.RLock()
	defer info.File.RUnlock()
	return info.File.ModTime
}

func (info *FileInfo) IsDir() bool {
	return info.Mode().IsDir()
}

// Sys returns the underlying *File.
func (info *FileInfo) Sys() interface{} {
	return info.File
}

// FileInfos represents an slice of os.FileInfo which
// implements the sort.Sort protocol. This type is only
// exported for users who want to implement their own
// filesystems, since VFS.ReadDir requires the returned
// []os.FileInfo to be sorted by name.
type FileInfos []os.FileInfo

func (f FileInfos) Len() int {
	return len(f)
}

func (f FileInfos) Less(i, j int) bool {
	return f[i].Name() < f[j].Name()
}

func (f FileInfos) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}
