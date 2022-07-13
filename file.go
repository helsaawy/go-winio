//go:build windows

package winio

import (
	"io"
	"syscall"

	"github.com/Microsoft/go-winio/internal/file"
)

var (
	ErrFileClosed = file.ErrFileClosed
	ErrTimeout    = file.ErrTimeout
)

func MakeOpenFile(h syscall.Handle) (io.ReadWriteCloser, error) {
	// If we return the result of makeWin32File directly, it can result in an
	// interface-wrapped nil, rather than a nil interface value.
	f, err := file.MakeWin32File(h, false)
	if err != nil {
		return nil, err
	}
	return f, nil
}
