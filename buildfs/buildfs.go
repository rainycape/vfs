// Package buildfs allows plugging a VFS into a go/build.Context.
package buildfs

import (
	"go/build"
	"io"
	"path"
	"path/filepath"
	"strings"

	"github.com/rainycape/vfs"
)

const (
	separator = "/"
)

// Setup configures a *build.Context to use the given VFS
// as its filesystem.
func Setup(ctx *build.Context, fs vfs.VFS) {
	ctx.JoinPath = path.Join
	ctx.SplitPathList = filepath.SplitList
	ctx.IsAbsPath = func(p string) bool {
		return p != "" && p[0] == '/'
	}
	ctx.IsDir = func(p string) bool {
		stat, err := fs.Stat(p)
		return err == nil && stat.IsDir()
	}
	ctx.HasSubdir = func(root, dir string) (string, bool) {
		root = path.Clean(root)
		if !strings.HasSuffix(root, separator) {
			root += separator
		}
		dir = path.Clean(dir)
		if !strings.HasPrefix(dir, root) {
			return "", false
		}
		return dir[len(root):], true
	}
	ctx.ReadDir = fs.ReadDir
	ctx.OpenFile = func(p string) (io.ReadCloser, error) {
		return fs.Open(p)
	}
}
