package socket

import (
	"errors"
	"fmt"
	"reflect"
	"unsafe"
)

// todo: should these be custom types to store the desired/actual size and addr family?

var (
	ErrBufferSize     = errors.New("buffer size")
	ErrInvalidPointer = errors.New("invalid pointer")
	ErrAddrFamily     = errors.New("address family")
)

const AddressFamilySize = unsafe.Sizeof(uint16(0)) // ushort

// todo: helsaawy - replace this with generics, along with GetSockName and co.

// RawSockaddr allows structs to be used with Bind, ConnectEx and other socket functions.
// The struct must meet the Win32 sockaddr requirements specified here:
// https://docs.microsoft.com/en-us/windows/win32/winsock/sockaddr-2
//
// This a dummy interface to prevent socket functions from accepting arbitrary interface{} parameters.
type RawSockaddr interface {
	IsRawSockaddr()
	Validate() error
}

func rawSockAsBuffer(r RawSockaddr) (unsafe.Pointer, int32, error) {
	v := reflect.ValueOf(r)
	if v.Type().Kind() != reflect.Ptr {
		return nil, 0, fmt.Errorf("receiver is not a pointer: %w", ErrBufferSize)
	}

	v = v.Elem()
	ptr := v.UnsafePointer()
	n := int32(v.Type().Size())
	if n < int32(AddressFamilySize) {
		return nil, 0, fmt.Errorf("RawSockaddr struct size is %d, should be larger than %d: %w", n, AddressFamilySize, ErrBufferSize)
	}

	return ptr, n, nil
}
