//go:build windows

package fs

import (
	"io"

	"golang.org/x/sys/windows"

	"github.com/Microsoft/go-winio/internal/fs"
)

var (
	ErrFileClosed = fs.ErrFileClosed
	ErrTimeout    = fs.ErrTimeout
)

func MakeOpenFile(h windows.Handle) (io.ReadWriteCloser, error) {
	// If we return the result of makeFile directly, it can result in an
	// interface-wrapped nil, rather than a nil interface value.
	f, err := fs.MakeFile(h)
	if err != nil {
		return nil, err
	}
	return f, nil
}
