package vfs

import (
	"os"
	"path"
	"sync"
	"time"
)

type EntryType uint8

const (
	EntryTypeFile EntryType = iota + 1
	EntryTypeDir
)

type Entry interface {
	Type() EntryType
	Size() int64
	FileMode() os.FileMode
	ModificationTime() time.Time
}

// Type File represents an in-memory file or
// directory. Most in-memory VFS should use this
// structure to represent their files, in order to
// save work.
type File struct {
	sync.RWMutex
	// Data contains the file data.
	Data []byte
	// Mode is the file or directory mode. Note that some filesystems
	// might ignore the permission bits.
	Mode os.FileMode
	// ModTime represents the last modification time to the file.
	ModTime time.Time
}

func (f *File) Type() EntryType {
	return EntryTypeFile
}

func (f *File) Size() int64 {
	f.RLock()
	defer f.RUnlock()
	return int64(len(f.Data))
}

func (f *File) FileMode() os.FileMode {
	return f.Mode
}

func (f *File) ModificationTime() time.Time {
	f.RLock()
	defer f.RUnlock()
	return f.ModTime
}

type Dir struct {
	sync.RWMutex
	// Mode is the file or directory mode. Note that some filesystems
	// might ignore the permission bits.
	Mode os.FileMode
	// ModTime represents the last modification time to directory.
	ModTime time.Time
	// Entry names in this directory, in order.
	EntryNames []string
	// Entries in the same order as EntryNames.
	Entries []Entry
}

func (d *Dir) Type() EntryType {
	return EntryTypeDir
}

func (d *Dir) Size() int64 {
	return 0
}

func (d *Dir) FileMode() os.FileMode {
	return d.Mode
}

func (d *Dir) ModificationTime() time.Time {
	d.RLock()
	defer d.RUnlock()
	return d.ModTime
}

func (d *Dir) Add(name string, entry Entry) error {
	// TODO: Binary search
	for ii, v := range d.EntryNames {
		if v > name {
			names := make([]string, len(d.EntryNames)+1)
			copy(names, d.EntryNames[:ii])
			names[ii] = name
			copy(names[ii+1:], d.EntryNames[ii:])
			d.EntryNames = names

			entries := make([]Entry, len(d.Entries)+1)
			copy(entries, d.Entries[:ii])
			entries[ii] = entry
			copy(entries[ii+1:], d.Entries[ii:])

			d.Entries = entries
			return nil
		}
		if v == name {
			return os.ErrExist
		}
	}
	// Not added yet, put at the end
	d.EntryNames = append(d.EntryNames, name)
	d.Entries = append(d.Entries, entry)
	return nil
}

func (d *Dir) Find(name string) (Entry, int, error) {
	for ii, v := range d.EntryNames {
		if v == name {
			return d.Entries[ii], ii, nil
		}
	}
	return nil, -1, os.ErrNotExist
}

// EntryInfo implements the os.FileInfo interface wrapping
// a given File and its Path in its VFS.
type EntryInfo struct {
	// Path is the full path to the entry in its VFS.
	Path string
	// Entry is the instance used by the VFS to represent
	// the in-memory entry.
	Entry Entry
}

func (info *EntryInfo) Name() string {
	return path.Base(info.Path)
}

func (info *EntryInfo) Size() int64 {
	return info.Entry.Size()
}

func (info *EntryInfo) Mode() os.FileMode {
	return info.Entry.FileMode()
}

func (info *EntryInfo) ModTime() time.Time {
	return info.Entry.ModificationTime()
}

func (info *EntryInfo) IsDir() bool {
	return info.Entry.Type() == EntryTypeDir
}

// Sys returns the underlying Entry.
func (info *EntryInfo) Sys() interface{} {
	return info.Entry
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
