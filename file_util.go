package vfs

import (
	"fmt"
	"io"
	"os"
	"time"
)

// NewRFile returns a RFile from a *File.
func NewRFile(f *File) RFile {
	return &file{f: f, readable: true}
}

// NewWFile returns a WFile from a *File.
func NewWFile(f *File, read bool, write bool) WFile {
	return &file{f: f, readable: read, writable: write}
}

type file struct {
	f        *File
	offset   int
	readable bool
	writable bool
}

func (f *file) Read(p []byte) (int, error) {
	if !f.readable {
		return 0, ErrWriteOnly
	}
	f.f.RLock()
	defer f.f.RUnlock()
	if f.offset > len(f.f.Data) {
		return 0, io.EOF
	}
	n := copy(p, f.f.Data[f.offset:])
	f.offset += n
	if n < len(p) {
		return n, io.EOF
	}
	return n, nil
}

func (f *file) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case os.SEEK_SET:
		f.offset = int(offset)
	case os.SEEK_CUR:
		f.offset += int(offset)
	case os.SEEK_END:
		f.offset = len(f.f.Data) + int(offset)
	default:
		panic(fmt.Errorf("Seek: invalid whence %d", whence))
	}
	if f.offset > len(f.f.Data) {
		f.offset = len(f.f.Data)
	} else if f.offset < 0 {
		f.offset = 0
	}
	return int64(f.offset), nil
}

func (f *file) Write(p []byte) (int, error) {
	if !f.writable {
		return 0, ErrReadOnly
	}
	f.f.Lock()
	defer f.f.Unlock()
	count := len(p)
	n := copy(f.f.Data[f.offset:], p)
	if n < count {
		f.f.Data = append(f.f.Data, p[n:]...)
	}
	f.offset += count
	f.f.ModTime = time.Now()
	return count, nil
}

func (f *file) Close() error {
	return nil
}
