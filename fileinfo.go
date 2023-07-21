//go:build windows

package winio

import (
	"os"

	"github.com/Microsoft/go-winio/internal/fs"
)

// FileBasicInfo contains file access time and file attributes information.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.FileBasicInfo instead.
type FileBasicInfo = fs.FileBasicInfo

// GetFileBasicInfo retrieves times and attributes for a file.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.GetFileBasicInfo instead.
func GetFileBasicInfo(f *os.File) (*FileBasicInfo, error) {
	return fs.GetFileBasicInfo(f)
}

// SetFileBasicInfo sets times and attributes for a file.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.SetFileBasicInfo instead.
func SetFileBasicInfo(f *os.File, bi *FileBasicInfo) error {
	return fs.SetFileBasicInfo(f, bi)
}

// FileStandardInfo contains extended information for the file.
// FILE_STANDARD_INFO in WinBase.h
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_standard_info
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.FileStandardInfo instead.
type FileStandardInfo = fs.FileStandardInfo

// GetFileStandardInfo retrieves ended information for the file.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.GetFileStandardInfo instead.
func GetFileStandardInfo(f *os.File) (*FileStandardInfo, error) {
	return fs.GetFileStandardInfo(f)
}

// FileIDInfo contains the volume serial number and file ID for a file. This pair should be
// unique on a system.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.FileIDInfo instead.
type FileIDInfo = fs.FileIDInfo

// GetFileID retrieves the unique (volume, file ID) pair for a file.
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/fs.GetFileID instead.
func GetFileID(f *os.File) (*FileIDInfo, error) {
	return fs.GetFileID(f)
}
