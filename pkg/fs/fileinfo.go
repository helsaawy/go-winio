//go:build windows

package fs

import (
	"os"
	"runtime"
	"unsafe"

	"golang.org/x/sys/windows"
)

//	typedef struct _FILE_BASIC_INFO {
//	  LARGE_INTEGER CreationTime;
//	  LARGE_INTEGER LastAccessTime;
//	  LARGE_INTEGER LastWriteTime;
//	  LARGE_INTEGER ChangeTime;
//	  DWORD         FileAttributes;
//	} FILE_BASIC_INFO, *PFILE_BASIC_INFO;
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_basic_info

// FileBasicInfo contains file access time and file attributes information.
type FileBasicInfo struct {
	CreationTime   windows.Filetime
	LastAccessTime windows.Filetime
	LastWriteTime  windows.Filetime
	ChangeTime     windows.Filetime
	FileAttributes uint32
	_              uint32 // padding
}

// GetFileBasicInfo retrieves times and attributes for a file.
func GetFileBasicInfo(f *os.File) (*FileBasicInfo, error) {
	bi := &FileBasicInfo{}
	if err := windows.GetFileInformationByHandleEx(
		windows.Handle(f.Fd()),
		windows.FileBasicInfo,
		(*byte)(unsafe.Pointer(bi)),
		uint32(unsafe.Sizeof(*bi)),
	); err != nil {
		return nil, &os.PathError{Op: "GetFileInformationByHandleEx", Path: f.Name(), Err: err}
	}
	runtime.KeepAlive(f)
	return bi, nil
}

// SetFileBasicInfo sets times and attributes for a file.
func SetFileBasicInfo(f *os.File, bi *FileBasicInfo) error {
	if err := windows.SetFileInformationByHandle(
		windows.Handle(f.Fd()),
		windows.FileBasicInfo,
		(*byte)(unsafe.Pointer(bi)),
		uint32(unsafe.Sizeof(*bi)),
	); err != nil {
		return &os.PathError{Op: "SetFileInformationByHandle", Path: f.Name(), Err: err}
	}
	runtime.KeepAlive(f)
	return nil
}

//  typedef struct _FILE_STANDARD_INFO {
//    LARGE_INTEGER AllocationSize;
//    LARGE_INTEGER EndOfFile;
//    DWORD         NumberOfLinks;
//    BOOLEAN       DeletePending;
//    BOOLEAN       Directory;
//  } FILE_STANDARD_INFO, *PFILE_STANDARD_INFO;
//
// https://docs.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_standard_info

// FileStandardInfo contains extended information for the file.
type FileStandardInfo struct {
	AllocationSize int64
	EndOfFile      int64
	NumberOfLinks  uint32
	DeletePending  bool
	Directory      bool
}

// GetFileStandardInfo retrieves ended information for the file.
func GetFileStandardInfo(f *os.File) (*FileStandardInfo, error) {
	si := &FileStandardInfo{}
	if err := windows.GetFileInformationByHandleEx(windows.Handle(f.Fd()),
		windows.FileStandardInfo,
		(*byte)(unsafe.Pointer(si)),
		uint32(unsafe.Sizeof(*si)),
	); err != nil {
		return nil, &os.PathError{Op: "GetFileInformationByHandleEx", Path: f.Name(), Err: err}
	}
	runtime.KeepAlive(f)
	return si, nil
}

//  typedef struct _FILE_ID_INFO {
//    ULONGLONG   VolumeSerialNumber;
//    FILE_ID_128 FileId;
//  } FILE_ID_INFO, *PFILE_ID_INFO;
//
// https://learn.microsoft.com/en-us/windows/win32/api/winbase/ns-winbase-file_id_info

// FileIDInfo contains the volume serial number and file ID for a file. This pair should be
// unique on a system.
type FileIDInfo struct {
	VolumeSerialNumber uint64
	FileID             [16]byte
}

// GetFileID retrieves the unique (volume, file ID) pair for a file.
func GetFileID(f *os.File) (*FileIDInfo, error) {
	fileID := &FileIDInfo{}
	if err := windows.GetFileInformationByHandleEx(
		windows.Handle(f.Fd()),
		windows.FileIdInfo,
		(*byte)(unsafe.Pointer(fileID)),
		uint32(unsafe.Sizeof(*fileID)),
	); err != nil {
		return nil, &os.PathError{Op: "GetFileInformationByHandleEx", Path: f.Name(), Err: err}
	}
	runtime.KeepAlive(f)
	return fileID, nil
}
