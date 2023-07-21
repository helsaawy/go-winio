//go:build windows

package winio

import (
	"github.com/Microsoft/go-winio/pkg/fs/reparse"
)

// ReparsePoint describes a Win32 symlink or mount point.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs/reparse.ReparsePoint instead.
type ReparsePoint = reparse.ReparsePoint

// UnsupportedReparsePointError is returned when trying to decode a non-symlink or
// mount point reparse point.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs/reparse.UnsupportedReparsePointError instead.
type UnsupportedReparsePointError = reparse.UnsupportedReparsePointError

// Deprecated: use github.com/Microsoft/go-winio/pkg/fs/reparse.Decode instead.
func DecodeReparsePoint(b []byte) (*ReparsePoint, error) {
	return reparse.Decode(b)
}

// Deprecated: use github.com/Microsoft/go-winio/pkg/fs/reparse.DecodeData instead.
func DecodeReparsePointData(tag uint32, b []byte) (*ReparsePoint, error) {
	return reparse.DecodeData(reparse.Tag(tag), b)
}

// EncodeReparsePoint encodes a Win32 REPARSE_DATA_BUFFER structure describing a symlink or
// mount point.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs/reparse.Encode instead.
func EncodeReparsePoint(rp *ReparsePoint) []byte {
	return reparse.Encode(rp)
}
