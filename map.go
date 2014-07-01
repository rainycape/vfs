package vfs

import (
	"path"
	"time"
)

// Map returns a MemoryFileSystem using the given files argument to
// prepopulate it (which might be nil). Note that the files map does
// not need to contain any directories, they will be created automatically.
// If the files contain conflicting paths (e.g. files named a and a/b, thus
// making "a" both a file and a directory), an error will be returned.
func Map(files map[string]*File) (*MemoryFileSystem, error) {
	fs := Memory()
	for k, v := range files {
		if v.Mode == 0 {
			v.Mode = 0644
		}
		if v.ModTime.IsZero() {
			v.ModTime = time.Now()
		}
		fs.files[k] = v
		if err := MkdirAll(fs, path.Dir(k), 0755); err != nil {
			return nil, err
		}
	}
	return fs, nil
}
