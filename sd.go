//go:build windows

package winio

import (
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/security"
)

// Deprecated: use github.com/Microsoft/go-winio/pkg/security.AccountLookupError instead.
type AccountLookupError = security.AccountLookupError

// dont forward SddlConversionError, since security SDDLConversionError has field SDDL, not Sddl

// Deprecated: use github.com/Microsoft/go-winio/pkg/security.SDDLConversionError instead.
//
//revive:disable-next-line:var-naming SDDL, not Sddl
type SddlConversionError struct {
	//revive:disable-next-line:var-naming SDDL, not Sddl
	Sddl string
	Err  error
}

func (e *SddlConversionError) Error() string {
	return "convert " + e.Sddl + ": " + e.Err.Error()
}

func (e *SddlConversionError) Unwrap() error { return e.Err }

// LookupSidByName looks up the SID of an account by name
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/security.LookupSIDByName instead.
//
//revive:disable-next-line:var-naming SID, not Sid
func LookupSidByName(name string) (sid string, err error) {
	return security.LookupSIDByName(name)
}

// LookupNameBySid looks up the name of an account by SID
//
// Deprecated: use github.com/Microsoft/go-winio/pkg/security.LookupNameBySID instead.
//
//revive:disable-next-line:var-naming SID, not Sid
func LookupNameBySid(sid string) (name string, err error) {
	return security.LookupNameBySID(sid)
}

// Deprecated: use github.com/Microsoft/go-winio/pkg/security.SDDLToSecurityDescriptor instead.
//
//revive:disable-next-line:var-naming SDDL, not Sddl
func SddlToSecurityDescriptor(sddl string) ([]byte, error) {
	var sdBuffer uintptr
	err := convertStringSecurityDescriptorToSecurityDescriptor(sddl, 1, &sdBuffer, nil)
	if err != nil {
		return nil, &SddlConversionError{sddl, err}
	}
	defer localFree(sdBuffer)
	sd := make([]byte, getSecurityDescriptorLength(sdBuffer))
	copy(sd, (*[0xffff]byte)(unsafe.Pointer(sdBuffer))[:len(sd)])
	return sd, nil
}

// Deprecated: use github.com/Microsoft/go-winio/pkg/security.SecurityDescriptorToSDDL instead.
//
//revive:disable-next-line:var-naming SDDL, not Sddl
func SecurityDescriptorToSddl(sd []byte) (string, error) {
	return security.SecurityDescriptorToSDDL(sd)
}
