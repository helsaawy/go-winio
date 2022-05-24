//go:build windows

// Code generated by 'go generate' using "github.com/Microsoft/go-winio/tools/mkwinsyscall"; DO NOT EDIT.

package winio

import (
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
)

var _ unsafe.Pointer

// Do the interface allocations only once for common
// Errno values.
const (
	errnoERROR_IO_PENDING = 997
)

var (
	errERROR_IO_PENDING error = syscall.Errno(errnoERROR_IO_PENDING)
	errERROR_EINVAL     error = syscall.EINVAL
)

// errnoErr returns common boxed Errno values, to prevent
// allocations at runtime.
func errnoErr(e syscall.Errno) error {
	switch e {
	case 0:
		return errERROR_EINVAL
	case errnoERROR_IO_PENDING:
		return errERROR_IO_PENDING
	}
	// TODO: add more here, after collecting data on the common
	// error values see on Windows. (perhaps when running
	// all.bat?)
	return e
}

var (
	modadvapi32 = windows.NewLazySystemDLL("advapi32.dll")
	modkernel32 = windows.NewLazySystemDLL("kernel32.dll")
	modntdll    = windows.NewLazySystemDLL("ntdll.dll")
	modws2_32   = windows.NewLazySystemDLL("ws2_32.dll")

	procAdjustTokenPrivileges                                = modadvapi32.NewProc("AdjustTokenPrivileges")
	procConvertSecurityDescriptorToStringSecurityDescriptorW = modadvapi32.NewProc("ConvertSecurityDescriptorToStringSecurityDescriptorW")
	procConvertSidToStringSidW                               = modadvapi32.NewProc("ConvertSidToStringSidW")
	procConvertStringSecurityDescriptorToSecurityDescriptorW = modadvapi32.NewProc("ConvertStringSecurityDescriptorToSecurityDescriptorW")
	procGetSecurityDescriptorLength                          = modadvapi32.NewProc("GetSecurityDescriptorLength")
	procImpersonateSelf                                      = modadvapi32.NewProc("ImpersonateSelf")
	procLookupAccountNameW                                   = modadvapi32.NewProc("LookupAccountNameW")
	procLookupPrivilegeDisplayNameW                          = modadvapi32.NewProc("LookupPrivilegeDisplayNameW")
	procLookupPrivilegeNameW                                 = modadvapi32.NewProc("LookupPrivilegeNameW")
	procLookupPrivilegeValueW                                = modadvapi32.NewProc("LookupPrivilegeValueW")
	procOpenThreadToken                                      = modadvapi32.NewProc("OpenThreadToken")
	procRevertToSelf                                         = modadvapi32.NewProc("RevertToSelf")
	procBackupRead                                           = modkernel32.NewProc("BackupRead")
	procBackupWrite                                          = modkernel32.NewProc("BackupWrite")
	procCancelIoEx                                           = modkernel32.NewProc("CancelIoEx")
	procConnectNamedPipe                                     = modkernel32.NewProc("ConnectNamedPipe")
	procCreateFileW                                          = modkernel32.NewProc("CreateFileW")
	procCreateIoCompletionPort                               = modkernel32.NewProc("CreateIoCompletionPort")
	procCreateNamedPipeW                                     = modkernel32.NewProc("CreateNamedPipeW")
	procGetCurrentThread                                     = modkernel32.NewProc("GetCurrentThread")
	procGetNamedPipeHandleStateW                             = modkernel32.NewProc("GetNamedPipeHandleStateW")
	procGetNamedPipeInfo                                     = modkernel32.NewProc("GetNamedPipeInfo")
	procGetQueuedCompletionStatus                            = modkernel32.NewProc("GetQueuedCompletionStatus")
	procSetFileCompletionNotificationModes                   = modkernel32.NewProc("SetFileCompletionNotificationModes")
	procNtCreateNamedPipeFile                                = modntdll.NewProc("NtCreateNamedPipeFile")
	procRtlDefaultNpAcl                                      = modntdll.NewProc("RtlDefaultNpAcl")
	procRtlDosPathNameToNtPathName_U                         = modntdll.NewProc("RtlDosPathNameToNtPathName_U")
	procRtlNtStatusToDosErrorNoTeb                           = modntdll.NewProc("RtlNtStatusToDosErrorNoTeb")
	procWSAGetOverlappedResult                               = modws2_32.NewProc("WSAGetOverlappedResult")
	procbind                                                 = modws2_32.NewProc("bind")
)

func adjustTokenPrivileges(token windows.Token, releaseAll bool, input *byte, outputSize uint32, output *byte, requiredSize *uint32) (success bool, err error) {
	var _p0 uint32
	if releaseAll {
		_p0 = 1
	}
	r0, _, e1 := syscall.Syscall6(procAdjustTokenPrivileges.Addr(), 6, uintptr(token), uintptr(_p0), uintptr(unsafe.Pointer(input)), uintptr(outputSize), uintptr(unsafe.Pointer(output)), uintptr(unsafe.Pointer(requiredSize)))
	success = r0 != 0
	if true {
		err = errnoErr(e1)
	}
	return
}

func convertSecurityDescriptorToStringSecurityDescriptor(sd *byte, revision uint32, secInfo uint32, sddl **uint16, sddlSize *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procConvertSecurityDescriptorToStringSecurityDescriptorW.Addr(), 5, uintptr(unsafe.Pointer(sd)), uintptr(revision), uintptr(secInfo), uintptr(unsafe.Pointer(sddl)), uintptr(unsafe.Pointer(sddlSize)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func convertSidToStringSid(sid *byte, str **uint16) (err error) {
	r1, _, e1 := syscall.Syscall(procConvertSidToStringSidW.Addr(), 2, uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(str)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func convertStringSecurityDescriptorToSecurityDescriptor(str string, revision uint32, sd *uintptr, size *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(str)
	if err != nil {
		return
	}
	return _convertStringSecurityDescriptorToSecurityDescriptor(_p0, revision, sd, size)
}

func _convertStringSecurityDescriptorToSecurityDescriptor(str *uint16, revision uint32, sd *uintptr, size *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procConvertStringSecurityDescriptorToSecurityDescriptorW.Addr(), 4, uintptr(unsafe.Pointer(str)), uintptr(revision), uintptr(unsafe.Pointer(sd)), uintptr(unsafe.Pointer(size)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func getSecurityDescriptorLength(sd uintptr) (len uint32) {
	r0, _, _ := syscall.Syscall(procGetSecurityDescriptorLength.Addr(), 1, uintptr(sd), 0, 0)
	len = uint32(r0)
	return
}

func impersonateSelf(level uint32) (err error) {
	r1, _, e1 := syscall.Syscall(procImpersonateSelf.Addr(), 1, uintptr(level), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func lookupAccountName(systemName *uint16, accountName string, sid *byte, sidSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(accountName)
	if err != nil {
		return
	}
	return _lookupAccountName(systemName, _p0, sid, sidSize, refDomain, refDomainSize, sidNameUse)
}

func _lookupAccountName(systemName *uint16, accountName *uint16, sid *byte, sidSize *uint32, refDomain *uint16, refDomainSize *uint32, sidNameUse *uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procLookupAccountNameW.Addr(), 7, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(accountName)), uintptr(unsafe.Pointer(sid)), uintptr(unsafe.Pointer(sidSize)), uintptr(unsafe.Pointer(refDomain)), uintptr(unsafe.Pointer(refDomainSize)), uintptr(unsafe.Pointer(sidNameUse)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func lookupPrivilegeDisplayName(systemName string, name *uint16, buffer *uint16, size *uint32, languageId *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(systemName)
	if err != nil {
		return
	}
	return _lookupPrivilegeDisplayName(_p0, name, buffer, size, languageId)
}

func _lookupPrivilegeDisplayName(systemName *uint16, name *uint16, buffer *uint16, size *uint32, languageId *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procLookupPrivilegeDisplayNameW.Addr(), 5, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(size)), uintptr(unsafe.Pointer(languageId)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func lookupPrivilegeName(systemName string, luid *uint64, buffer *uint16, size *uint32) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(systemName)
	if err != nil {
		return
	}
	return _lookupPrivilegeName(_p0, luid, buffer, size)
}

func _lookupPrivilegeName(systemName *uint16, luid *uint64, buffer *uint16, size *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procLookupPrivilegeNameW.Addr(), 4, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(luid)), uintptr(unsafe.Pointer(buffer)), uintptr(unsafe.Pointer(size)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func lookupPrivilegeValue(systemName string, name string, luid *uint64) (err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(systemName)
	if err != nil {
		return
	}
	var _p1 *uint16
	_p1, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _lookupPrivilegeValue(_p0, _p1, luid)
}

func _lookupPrivilegeValue(systemName *uint16, name *uint16, luid *uint64) (err error) {
	r1, _, e1 := syscall.Syscall(procLookupPrivilegeValueW.Addr(), 3, uintptr(unsafe.Pointer(systemName)), uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(luid)))
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func openThreadToken(thread syscall.Handle, accessMask uint32, openAsSelf bool, token *windows.Token) (err error) {
	var _p0 uint32
	if openAsSelf {
		_p0 = 1
	}
	r1, _, e1 := syscall.Syscall6(procOpenThreadToken.Addr(), 4, uintptr(thread), uintptr(accessMask), uintptr(_p0), uintptr(unsafe.Pointer(token)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func revertToSelf() (err error) {
	r1, _, e1 := syscall.Syscall(procRevertToSelf.Addr(), 0, 0, 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func backupRead(h syscall.Handle, b []byte, bytesRead *uint32, abort bool, processSecurity bool, context *uintptr) (err error) {
	var _p0 *byte
	if len(b) > 0 {
		_p0 = &b[0]
	}
	var _p1 uint32
	if abort {
		_p1 = 1
	}
	var _p2 uint32
	if processSecurity {
		_p2 = 1
	}
	r1, _, e1 := syscall.Syscall9(procBackupRead.Addr(), 7, uintptr(h), uintptr(unsafe.Pointer(_p0)), uintptr(len(b)), uintptr(unsafe.Pointer(bytesRead)), uintptr(_p1), uintptr(_p2), uintptr(unsafe.Pointer(context)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func backupWrite(h syscall.Handle, b []byte, bytesWritten *uint32, abort bool, processSecurity bool, context *uintptr) (err error) {
	var _p0 *byte
	if len(b) > 0 {
		_p0 = &b[0]
	}
	var _p1 uint32
	if abort {
		_p1 = 1
	}
	var _p2 uint32
	if processSecurity {
		_p2 = 1
	}
	r1, _, e1 := syscall.Syscall9(procBackupWrite.Addr(), 7, uintptr(h), uintptr(unsafe.Pointer(_p0)), uintptr(len(b)), uintptr(unsafe.Pointer(bytesWritten)), uintptr(_p1), uintptr(_p2), uintptr(unsafe.Pointer(context)), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func cancelIoEx(file syscall.Handle, o *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall(procCancelIoEx.Addr(), 2, uintptr(file), uintptr(unsafe.Pointer(o)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func connectNamedPipe(pipe syscall.Handle, o *syscall.Overlapped) (err error) {
	r1, _, e1 := syscall.Syscall(procConnectNamedPipe.Addr(), 2, uintptr(pipe), uintptr(unsafe.Pointer(o)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func createFile(name string, access uint32, mode uint32, sa *syscall.SecurityAttributes, createmode uint32, attrs uint32, templatefile syscall.Handle) (handle syscall.Handle, err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _createFile(_p0, access, mode, sa, createmode, attrs, templatefile)
}

func _createFile(name *uint16, access uint32, mode uint32, sa *syscall.SecurityAttributes, createmode uint32, attrs uint32, templatefile syscall.Handle) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall9(procCreateFileW.Addr(), 7, uintptr(unsafe.Pointer(name)), uintptr(access), uintptr(mode), uintptr(unsafe.Pointer(sa)), uintptr(createmode), uintptr(attrs), uintptr(templatefile), 0, 0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		err = errnoErr(e1)
	}
	return
}

func createIoCompletionPort(file syscall.Handle, port syscall.Handle, key uintptr, threadCount uint32) (newport syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall6(procCreateIoCompletionPort.Addr(), 4, uintptr(file), uintptr(port), uintptr(key), uintptr(threadCount), 0, 0)
	newport = syscall.Handle(r0)
	if newport == 0 {
		err = errnoErr(e1)
	}
	return
}

func createNamedPipe(name string, flags uint32, pipeMode uint32, maxInstances uint32, outSize uint32, inSize uint32, defaultTimeout uint32, sa *syscall.SecurityAttributes) (handle syscall.Handle, err error) {
	var _p0 *uint16
	_p0, err = syscall.UTF16PtrFromString(name)
	if err != nil {
		return
	}
	return _createNamedPipe(_p0, flags, pipeMode, maxInstances, outSize, inSize, defaultTimeout, sa)
}

func _createNamedPipe(name *uint16, flags uint32, pipeMode uint32, maxInstances uint32, outSize uint32, inSize uint32, defaultTimeout uint32, sa *syscall.SecurityAttributes) (handle syscall.Handle, err error) {
	r0, _, e1 := syscall.Syscall9(procCreateNamedPipeW.Addr(), 8, uintptr(unsafe.Pointer(name)), uintptr(flags), uintptr(pipeMode), uintptr(maxInstances), uintptr(outSize), uintptr(inSize), uintptr(defaultTimeout), uintptr(unsafe.Pointer(sa)), 0)
	handle = syscall.Handle(r0)
	if handle == syscall.InvalidHandle {
		err = errnoErr(e1)
	}
	return
}

func getCurrentThread() (h syscall.Handle) {
	r0, _, _ := syscall.Syscall(procGetCurrentThread.Addr(), 0, 0, 0, 0)
	h = syscall.Handle(r0)
	return
}

func getNamedPipeHandleState(pipe syscall.Handle, state *uint32, curInstances *uint32, maxCollectionCount *uint32, collectDataTimeout *uint32, userName *uint16, maxUserNameSize uint32) (err error) {
	r1, _, e1 := syscall.Syscall9(procGetNamedPipeHandleStateW.Addr(), 7, uintptr(pipe), uintptr(unsafe.Pointer(state)), uintptr(unsafe.Pointer(curInstances)), uintptr(unsafe.Pointer(maxCollectionCount)), uintptr(unsafe.Pointer(collectDataTimeout)), uintptr(unsafe.Pointer(userName)), uintptr(maxUserNameSize), 0, 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func getNamedPipeInfo(pipe syscall.Handle, flags *uint32, outSize *uint32, inSize *uint32, maxInstances *uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetNamedPipeInfo.Addr(), 5, uintptr(pipe), uintptr(unsafe.Pointer(flags)), uintptr(unsafe.Pointer(outSize)), uintptr(unsafe.Pointer(inSize)), uintptr(unsafe.Pointer(maxInstances)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func getQueuedCompletionStatus(port syscall.Handle, bytes *uint32, key *uintptr, o **ioOperation, timeout uint32) (err error) {
	r1, _, e1 := syscall.Syscall6(procGetQueuedCompletionStatus.Addr(), 5, uintptr(port), uintptr(unsafe.Pointer(bytes)), uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(o)), uintptr(timeout), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func setFileCompletionNotificationModes(h syscall.Handle, flags uint8) (err error) {
	r1, _, e1 := syscall.Syscall(procSetFileCompletionNotificationModes.Addr(), 2, uintptr(h), uintptr(flags), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func ntCreateNamedPipeFile(pipe *syscall.Handle, access uint32, oa *objectAttributes, iosb *ioStatusBlock, share uint32, disposition uint32, options uint32, typ uint32, readMode uint32, completionMode uint32, maxInstances uint32, inboundQuota uint32, outputQuota uint32, timeout *int64) (status ntstatus) {
	r0, _, _ := syscall.Syscall15(procNtCreateNamedPipeFile.Addr(), 14, uintptr(unsafe.Pointer(pipe)), uintptr(access), uintptr(unsafe.Pointer(oa)), uintptr(unsafe.Pointer(iosb)), uintptr(share), uintptr(disposition), uintptr(options), uintptr(typ), uintptr(readMode), uintptr(completionMode), uintptr(maxInstances), uintptr(inboundQuota), uintptr(outputQuota), uintptr(unsafe.Pointer(timeout)), 0)
	status = ntstatus(r0)
	return
}

func rtlDefaultNpAcl(dacl *uintptr) (status ntstatus) {
	r0, _, _ := syscall.Syscall(procRtlDefaultNpAcl.Addr(), 1, uintptr(unsafe.Pointer(dacl)), 0, 0)
	status = ntstatus(r0)
	return
}

func rtlDosPathNameToNtPathName(name *uint16, ntName *unicodeString, filePart uintptr, reserved uintptr) (status ntstatus) {
	r0, _, _ := syscall.Syscall6(procRtlDosPathNameToNtPathName_U.Addr(), 4, uintptr(unsafe.Pointer(name)), uintptr(unsafe.Pointer(ntName)), uintptr(filePart), uintptr(reserved), 0, 0)
	status = ntstatus(r0)
	return
}

func rtlNtStatusToDosError(status ntstatus) (winerr error) {
	r0, _, _ := syscall.Syscall(procRtlNtStatusToDosErrorNoTeb.Addr(), 1, uintptr(status), 0, 0)
	if r0 != 0 {
		winerr = syscall.Errno(r0)
	}
	return
}

func wsaGetOverlappedResult(h syscall.Handle, o *syscall.Overlapped, bytes *uint32, wait bool, flags *uint32) (err error) {
	var _p0 uint32
	if wait {
		_p0 = 1
	}
	r1, _, e1 := syscall.Syscall6(procWSAGetOverlappedResult.Addr(), 5, uintptr(h), uintptr(unsafe.Pointer(o)), uintptr(unsafe.Pointer(bytes)), uintptr(_p0), uintptr(unsafe.Pointer(flags)), 0)
	if r1 == 0 {
		err = errnoErr(e1)
	}
	return
}

func bind(s syscall.Handle, name unsafe.Pointer, namelen int32) (err error) {
	r1, _, e1 := syscall.Syscall(procbind.Addr(), 3, uintptr(s), uintptr(name), uintptr(namelen))
	if r1 == socketError {
		err = errnoErr(e1)
	}
	return
}
