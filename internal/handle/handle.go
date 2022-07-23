//go:build windows

package handle

import (
	"fmt"

	"golang.org/x/sys/windows"
)

// BOOL CompareObjectHandles(
//   [in] HANDLE hFirstObjectHandle,
//   [in] HANDLE hSecondObjectHandle
// );
//
//sys Compare(one windows.Handle, two windows.Handle) (b bool) = CompareObjectHandles

// BOOL GetHandleInformation(
//   [in]  HANDLE  hObject,
//   [out] LPDWORD lpdwFlags
// );
//
//sys getHandleInformation(h windows.Handle, i *Flags) (err error) [failretval != 0] = GetHandleInformation

const (
	Inherit          = 0x1
	ProtectFromClose = 0x2
)

type Flags uint32

func (f Flags) Inherit() bool {
	return f&Inherit != 0
}

func (f Flags) ProtectFromClose() bool {
	return f&ProtectFromClose != 0
}

func GetInformation(h windows.Handle) (f Flags, err error) {
	err = getHandleInformation(h, &f)
	return f, err
}

const (
	DuplicateCloseSource = 0x1
	DuplicateSameAccess  = 0x2
)

// Duplicate duplicates the provided handle using the same access and inheritance flags
func Duplicate(h windows.Handle) (t windows.Handle, err error) {
	f, err := GetInformation(h)
	if err != nil {
		return windows.InvalidHandle, fmt.Errorf("get handle information: %w", err)
	}
	p := windows.CurrentProcess()
	err = windows.DuplicateHandle(p, h, p, &t, 0 /* dwDesiredAccess */, f.Inherit(), DuplicateSameAccess)
	return t, err
}
