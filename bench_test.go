package vfs

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

var (
	goTestFile      = filepath.Join("testdata", "go1.3.src.tar.gz")
	errNoGoTestFile = fmt.Errorf("%s not found, use testdata/download-data.sh to fetch it", goTestFile)
)

func BenchmarkLoadGoSrc(b *testing.B) {
	f, err := os.Open(goTestFile)
	if err != nil {
		b.Skip(errNoGoTestFile)
	}
	defer f.Close()
	// Decompress to avoid measuring the time to gunzip
	zr, err := gzip.NewReader(f)
	if err != nil {
		b.Fatal(err)
	}
	defer zr.Close()
	data, err := ioutil.ReadAll(zr)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		if _, err := Tar(bytes.NewReader(data)); err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkWalkGoSrc(b *testing.B) {
	f, err := os.Open(goTestFile)
	if err != nil {
		b.Skip(errNoGoTestFile)
	}
	defer f.Close()
	fs, err := TarGzip(f)
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for ii := 0; ii < b.N; ii++ {
		Walk(fs, "/", func(_ VFS, _ string, _ os.FileInfo, _ error) error { return nil })
	}
}
