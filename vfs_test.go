package vfs

import (
	"os"
	"reflect"
	"testing"
)

func testVFS(t *testing.T, fs VFS) {
	if err := WriteFile(fs, "a", []byte("A"), 0644); err != nil {
		t.Fatal(err)
	}
	data, err := ReadFile(fs, "a")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "A" {
		t.Errorf("expecting file a to contain \"A\" got %q instead", string(data))
	}
	if err := WriteFile(fs, "b", []byte("B"), 0755); err != nil {
		t.Fatal(err)
	}
	if _, err := fs.OpenFile("b", os.O_CREATE|os.O_TRUNC|os.O_EXCL|os.O_WRONLY, 0755); err == nil || !IsExist(err) {
		t.Errorf("error should be ErrExist, it's %v", err)
	}
	fb, err := fs.OpenFile("b", os.O_TRUNC|os.O_WRONLY, 0755)
	if err != nil {
		t.Fatalf("error opening b: %s", err)
	}
	if _, err := fb.Write([]byte("BB")); err != nil {
		t.Errorf("error writing to b: %s", err)
	}
	if _, err := fb.Seek(0, os.SEEK_SET); err != nil {
		t.Errorf("error seeking b: %s", err)
	}
	if _, err := fb.Read(make([]byte, 2)); err == nil {
		t.Error("allowed reading WRONLY file b")
	}
	if err := fb.Close(); err != nil {
		t.Errorf("error closing b: %s", err)
	}
	files, err := fs.ReadDir("/")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Errorf("expecting 2 files, got %d", len(files))
	}
	if n := files[0].Name(); n != "a" {
		t.Errorf("expecting first file named \"a\", got %q", n)
	}
	if n := files[1].Name(); n != "b" {
		t.Errorf("expecting first file named \"b\", got %q", n)
	}
	for ii, v := range files {
		es := int64(ii + 1)
		if s := v.Size(); es != s {
			t.Errorf("expecting file %s to have size %d, has %d", v.Name(), es, s)
		}
	}
	if err := MkdirAll(fs, "a/b/c/d", 0); err == nil {
		t.Error("should not allow dir over file")
	}
	if err := MkdirAll(fs, "c/d", 0755); err != nil {
		t.Fatal(err)
	}
	// Idempotent
	if err := MkdirAll(fs, "c/d", 0755); err != nil {
		t.Fatal(err)
	}
	if err := fs.Mkdir("c", 0755); err == nil || !IsExist(err) {
		t.Errorf("err should be ErrExist, it's %v", err)
	}
	// Should fail to remove, c is not empty
	if err := fs.Remove("c"); err == nil {
		t.Fatalf("removed non-empty directory")
	}
	var walked []os.FileInfo
	var walkedNames []string
	err = Walk(fs, "c", func(fs VFS, path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		walked = append(walked, info)
		walkedNames = append(walkedNames, path)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if exp := []string{"c", "c/d"}; !reflect.DeepEqual(exp, walkedNames) {
		t.Error("expecting walked names %v, got %v", exp, walkedNames)
	}
	for _, v := range walked {
		if !v.IsDir() {
			t.Errorf("%s should be a dir", v.Name())
		}
	}
	if err := RemoveAll(fs, "c"); err != nil {
		t.Fatal(err)
	}
	err = Walk(fs, "c", func(fs VFS, path string, info os.FileInfo, err error) error {
		return err
	})
	if err == nil || !IsNotExist(err) {
		t.Errorf("error should be ErrNotExist, it's %v", err)
	}
}

func TestMapFS(t *testing.T) {
	fs, err := Map(nil)
	if err != nil {
		t.Fatal(err)
	}
	testVFS(t, fs)
}

func TestPopulatedMap(t *testing.T) {
	files := map[string]*File{
		"a/1": &File{},
		"a/2": &File{},
	}
	fs, err := Map(files)
	if err != nil {
		t.Fatal(err)
	}
	infos, err := fs.ReadDir("a")
	if err != nil {
		t.Fatal(err)
	}
	if c := len(infos); c != 2 {
		t.Fatalf("expecting 2 files in a, got %d", c)
	}
	if infos[0].Name() != "1" || infos[1].Name() != "2" {
		t.Errorf("expecting names 1, 2 got %q, %q", infos[0].Name(), infos[1].Name())
	}
}

func TestBadPopulatedMap(t *testing.T) {
	// 1 can't be file and directory
	files := map[string]*File{
		"a/1":   &File{},
		"a/1/2": &File{},
	}
	_, err := Map(files)
	if err == nil {
		t.Fatal("Map should not work with a path as both file and directory")
	}
}

func TestTmpFS(t *testing.T) {
	fs, err := TmpFS("vfs-test")
	if err != nil {
		t.Fatal(err)
	}
	defer fs.Close()
	testVFS(t, fs)
}
