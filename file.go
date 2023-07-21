//go:build windows

package winio

import (
	"io"
	"syscall"

	"github.com/Microsoft/go-winio/pkg/fs"
	"golang.org/x/sys/windows"
)

// Deprecated: use github.com/Microsoft/go-winio/pkg/fs instead.
var (
	ErrFileClosed = fs.ErrFileClosed
	ErrTimeout    = fs.ErrTimeout
)

// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.MakeOpenFile instead.
func MakeOpenFile(h syscall.Handle) (io.ReadWriteCloser, error) {
	return fs.MakeOpenFile(windows.Handle(h))
}
