//go:build windows

package security

import (
	"errors"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

//sys lookupAccountName(systemName *uint16, accountName string, sid *byte, sidSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) = advapi32.LookupAccountNameW
//sys lookupAccountSid(systemName *uint16, sid *byte, name *uint16, nameSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) = advapi32.LookupAccountSidW
//sys convertSidToStringSid(sid *byte, str **uint16) (err error) = advapi32.ConvertSidToStringSidW
//sys convertStringSidToSid(str string, sid **byte) (err error) = advapi32.ConvertStringSidToSidW
//sys convertStringSecurityDescriptorToSecurityDescriptor(str string, revision uint32, sd *uintptr, size *uint32) (err error) = advapi32.ConvertStringSecurityDescriptorToSecurityDescriptorW
//sys convertSecurityDescriptorToStringSecurityDescriptor(sd *byte, revision uint32, secInfo uint32, sddl **uint16, sddlSize *uint32) (err error) = advapi32.ConvertSecurityDescriptorToStringSecurityDescriptorW
//sys localFree(mem uintptr) = LocalFree
//sys getSecurityDescriptorLength(sd uintptr) (len uint32) = advapi32.GetSecurityDescriptorLength

type AccountLookupError struct {
	Name string
	Err  error
}

func (e *AccountLookupError) Error() string {
	if e.Name == "" {
		return "lookup account: empty account name specified"
	}
	var s string
	switch {
	case errors.Is(e.Err, windows.ERROR_INVALID_SID):
		s = "the security ID structure is invalid"
	case errors.Is(e.Err, windows.ERROR_NONE_MAPPED):
		s = "not found"
	default:
		s = e.Err.Error()
	}
	return "lookup account " + e.Name + ": " + s
}

func (e *AccountLookupError) Unwrap() error { return e.Err }

type SDDLConversionError struct {
	SDDL string
	Err  error
}

func (e *SDDLConversionError) Error() string {
	return "convert " + e.SDDL + ": " + e.Err.Error()
}

func (e *SDDLConversionError) Unwrap() error { return e.Err }

// LookupSIDByName looks up the SID of an account by name
//
//revive:disable-next-line:var-naming SID, not Sid
func LookupSIDByName(name string) (sid string, err error) {
	if name == "" {
		return "", &AccountLookupError{name, windows.ERROR_NONE_MAPPED}
	}

	var sidSize, sidNameUse, refDomainSize uint32
	err = lookupAccountName(nil, name, nil, &sidSize, nil, &refDomainSize, &sidNameUse)
	if err != nil && err != syscall.ERROR_INSUFFICIENT_BUFFER { //nolint:errorlint // err is Errno
		return "", &AccountLookupError{name, err}
	}
	sidBuffer := make([]byte, sidSize)
	refDomainBuffer := make([]uint16, refDomainSize)
	err = lookupAccountName(nil, name, &sidBuffer[0], &sidSize, &refDomainBuffer[0], &refDomainSize, &sidNameUse)
	if err != nil {
		return "", &AccountLookupError{name, err}
	}
	var strBuffer *uint16
	err = convertSidToStringSid(&sidBuffer[0], &strBuffer)
	if err != nil {
		return "", &AccountLookupError{name, err}
	}
	sid = syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(strBuffer))[:])
	localFree(uintptr(unsafe.Pointer(strBuffer)))
	return sid, nil
}

// LookupNameBySID looks up the name of an account by SID
func LookupNameBySID(sid string) (name string, err error) {
	if sid == "" {
		return "", &AccountLookupError{sid, windows.ERROR_NONE_MAPPED}
	}

	var sidPtr *byte
	if err = convertStringSidToSid(sid, &sidPtr); err != nil {
		return "", &AccountLookupError{sid, err}
	}
	defer localFree(uintptr(unsafe.Pointer(sidPtr)))

	var nameSize, refDomainSize, sidNameUse uint32
	err = lookupAccountSid(nil, sidPtr, nil, &nameSize, nil, &refDomainSize, &sidNameUse)
	if err != nil && err != windows.ERROR_INSUFFICIENT_BUFFER { //nolint:errorlint // err is Errno
		return "", &AccountLookupError{sid, err}
	}

	nameBuffer := make([]uint16, nameSize)
	refDomainBuffer := make([]uint16, refDomainSize)
	err = lookupAccountSid(nil, sidPtr, &nameBuffer[0], &nameSize, &refDomainBuffer[0], &refDomainSize, &sidNameUse)
	if err != nil {
		return "", &AccountLookupError{sid, err}
	}

	name = windows.UTF16ToString(nameBuffer)
	return name, nil
}

func SDDLToSecurityDescriptor(sddl string) ([]byte, error) {
	var sdBuffer uintptr
	err := convertStringSecurityDescriptorToSecurityDescriptor(sddl, 1, &sdBuffer, nil)
	if err != nil {
		return nil, &SDDLConversionError{sddl, err}
	}
	defer localFree(sdBuffer)
	sd := make([]byte, getSecurityDescriptorLength(sdBuffer))
	copy(sd, (*[0xffff]byte)(unsafe.Pointer(sdBuffer))[:len(sd)])
	return sd, nil
}

//revive:disable-next-line:exported SecurityDescriptor is the accepted name
func SecurityDescriptorToSDDL(sd []byte) (string, error) {
	var sddl *uint16
	// The returned string length seems to include an arbitrary number of terminating NULs.
	// Don't use it.
	err := convertSecurityDescriptorToStringSecurityDescriptor(&sd[0], 1, 0xff, &sddl, nil)
	if err != nil {
		return "", err
	}
	defer localFree(uintptr(unsafe.Pointer(sddl)))
	return syscall.UTF16ToString((*[0xffff]uint16)(unsafe.Pointer(sddl))[:]), nil
}
