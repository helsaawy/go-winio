//go:build go1.18

package sockets

import (
	"fmt"
	"net"
	"sync"
	"syscall"
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/guid"
	"golang.org/x/sys/windows"
)

//go:generate go run golang.org/x/sys/windows/mkwinsyscall -output zsyscall_windows.go sockets.go

//sys getsockname(s windows.Handle, name unsafe.Pointer, namelen *int32) (err error) [failretval==socketError] = ws2_32.getsockname
//sys getpeername(s windows.Handle, name unsafe.Pointer, namelen *int32) (err error) [failretval==socketError] = ws2_32.getpeername
//sys bind(s windows.Handle, name unsafe.Pointer, namelen int32) (err error) [failretval==socketError] = ws2_32.bind

const socketError = uintptr(^uint32(0))

// CloseWriter is a connection that can disable writing to itself.
type CloseWriter interface {
	net.Conn
	CloseWrite() error
}

// CloseReader is a connection that can disable reading from itself.
type CloseReader interface {
	net.Conn
	CloseRead() error
}

// GetSockName returns the socket's local address. It will call the `rsa.FromBytes()` on the
// buffer returned by the getsockname syscall. The buffer is allocated to the size specified
// by `rsa.Sockaddr()`.
func GetSockName[T RawSockaddr](s windows.Handle, rsa *T) error {
	l := int32(unsafe.Sizeof(*rsa))
	return getsockname(s, unsafe.Pointer(rsa), &l)
}

// GetPeerName returns the remote address the socket is connected to.
//
// See GetSockName for more information.
func GetPeerName[T RawSockaddr](s windows.Handle, rsa *T) error {
	l := int32(unsafe.Sizeof(*rsa))
	return getpeername(s, unsafe.Pointer(rsa), &l)
}

func Bind[T RawSockaddr](s windows.Handle, rsa *T) (err error) {
	l := int32(unsafe.Sizeof(*rsa))
	return bind(s, unsafe.Pointer(rsa), l)
}

// "golang.org/x/sys/windows".ConnectEx and .Bind only accept internal implementations of the
// their sockaddr interface, so they cannot be used with HvsockAddr
// Replicate functionality here from
// https://cs.opensource.google/go/x/sys/+/master:windows/syscall_windows.go

// The function pointers to `AcceptEx`, `ConnectEx` and `GetAcceptExSockaddrs` must be loaded at
// runtime via a WSAIoctl call:
// https://docs.microsoft.com/en-us/windows/win32/api/Mswsock/nc-mswsock-lpfn_connectex#remarks

type runtimeFunc struct {
	id   guid.GUID
	once sync.Once
	addr uintptr
	err  error
}

func (f *runtimeFunc) Load() error {
	f.once.Do(func() {
		var s windows.Handle
		s, f.err = windows.Socket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP)
		if f.err != nil {
			return
		}
		defer windows.CloseHandle(s)

		var n uint32
		f.err = windows.WSAIoctl(s,
			windows.SIO_GET_EXTENSION_FUNCTION_POINTER,
			(*byte)(unsafe.Pointer(&f.id)),
			uint32(unsafe.Sizeof(f.id)),
			(*byte)(unsafe.Pointer(&f.addr)),
			uint32(unsafe.Sizeof(f.addr)),
			&n,
			nil, //overlapped
			0,   //completionRoutine
		)
	})
	return f.err

}

var (
	// todo: add `AcceptEx` and `GetAcceptExSockaddrs`
	WSAID_CONNECTEX = guid.GUID{
		Data1: 0x25a207b9,
		Data2: 0xddf3,
		Data3: 0x4660,
		Data4: [8]byte{0x8e, 0xe9, 0x76, 0xe5, 0x8c, 0x74, 0x06, 0x3e},
	}

	connectExFunc = runtimeFunc{id: WSAID_CONNECTEX}
)

func ConnectEx[T RawSockaddr](fd windows.Handle, rsa *T, sendBuf *byte, sendDataLen uint32, bytesSent *uint32, overlapped *windows.Overlapped) error {
	err := connectExFunc.Load()
	if err != nil {
		return fmt.Errorf("failed to load ConnectEx function pointer: %e", err)
	}
	return connectEx(fd, unsafe.Pointer(rsa), int32(unsafe.Sizeof(*rsa)), sendBuf, sendDataLen, bytesSent, overlapped)
}

// BOOL LpfnConnectex(
//   [in]           SOCKET s,
//   [in]           const sockaddr *name,
//   [in]           int namelen,
//   [in, optional] PVOID lpSendBuffer,
//   [in]           DWORD dwSendDataLength,
//   [out]          LPDWORD lpdwBytesSent,
//   [in]           LPOVERLAPPED lpOverlapped
// )
func connectEx(s windows.Handle, name unsafe.Pointer, namelen int32, sendBuf *byte, sendDataLen uint32, bytesSent *uint32, overlapped *windows.Overlapped) (err error) {
	r1, _, e1 := syscall.SyscallN(connectExFunc.addr, 7, uintptr(s), uintptr(name), uintptr(namelen), uintptr(unsafe.Pointer(sendBuf)), uintptr(sendDataLen), uintptr(unsafe.Pointer(bytesSent)), uintptr(unsafe.Pointer(overlapped)), 0, 0)
	if r1 == 0 {
		if e1 != 0 {
			err = error(e1)
		} else {
			err = syscall.EINVAL
		}
	}
	return
}
