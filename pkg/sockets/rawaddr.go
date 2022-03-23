//go:build go1.18

package sockets

import (
	"errors"
)

// todo: should these be custom types to store the desired/actual size and addr family?

var (
	ErrBufferSize     = errors.New("buffer size")
	ErrInvalidPointer = errors.New("invalid pointer")
	ErrAddrFamily     = errors.New("address family")
)

// RawSockaddr allows structs to be used with Bind and ConnectEx. The
// struct must meet the Wind32 sockaddr requirements specified here:
// https://docs.microsoft.com/en-us/windows/win32/winsock/sockaddr-2
type RawSockaddr interface {
	AddressFamily() uint16
}

type RawSockaddrHeader struct {
	Family uint16
}

func (a RawSockaddrHeader) AddressFamily() uint16 {
	return a.Family
}

var _ RawSockaddr = RawSockaddrHeader{}
