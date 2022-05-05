//go:build go1.18 && windows

// This package allows lazy loading of Windows procedures and syscalls that
// would not otherwise be possible with "golang.org/x/sys/windows".Proc or .LazyProc
package proc

import (
	"fmt"
	"os"
	"sync"
	"syscall"
	"unsafe"

	"github.com/Microsoft/go-winio/pkg/guid"
	"golang.org/x/sys/windows"
)

var ErrNotFound = fmt.Errorf("proc not found: %w", os.ErrNotExist)

// Proc replicates the `windows.Proc`, but allows for custom implementations.
type Proc interface {
	// Addr returns the in-memory address of the procedure.
	Addr() uintptr
	// Call executes procedure p with arguments a.
	Call(a ...uintptr) (uintptr, uintptr, error)
}

// LazyProc replicates the `windows.LazyProc`, but allows for custom implementations.
//
// Implementations should panic in `Call` if `Find` fails.
type LazyProc interface {
	Proc
	// Find searches for the specified procedure, and returns an error if it cannot be found.
	// Implementations should only search for the procedure once.
	Find() error
}

// helpers and default implementations

// Call passes its arguments to `syscall.SyscallN` and returns the results.
func call(p Proc, a ...uintptr) (uintptr, uintptr, error) {
	return syscall.SyscallN(p.Addr(), a...)
}

// MustFind panics if `LazyProc.Find` returns an error.
func mustFind(p LazyProc) {
	must(p.Find())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

// wsaIoctlFunc is a procedure that must be loaded at runtime via a WSAIoctl call.
type wsaIoctlFunc struct {
	name string
	id   guid.GUID

	once sync.Once
	addr uintptr
	err  error
}

func NewWSAIoctlFunc(name string, id guid.GUID) LazyProc {
	return &wsaIoctlFunc{
		name: name,
		id:   id,
	}
}

var _ LazyProc = &wsaIoctlFunc{}

func (p *wsaIoctlFunc) Name() string {
	return p.name
}

func (p *wsaIoctlFunc) Addr() uintptr {
	mustFind(p)
	return p.addr
}

func (p *wsaIoctlFunc) Call(a ...uintptr) (uintptr, uintptr, error) {
	return call(p)
}

func (p *wsaIoctlFunc) Find() error {
	p.once.Do(func() {
		var err error
		defer func() {
			if err != nil {
				p.err = fmt.Errorf("error loading function %s: %w", p.name, err)
			}
		}()

		var s windows.Handle
		s, err = windows.Socket(windows.AF_INET, windows.SOCK_STREAM, windows.IPPROTO_TCP)
		if err != nil {
			return
		}
		defer windows.CloseHandle(s) //nolint:errcheck

		var n uint32
		err = windows.WSAIoctl(s,
			windows.SIO_GET_EXTENSION_FUNCTION_POINTER,
			(*byte)(unsafe.Pointer(&p.id)),
			uint32(unsafe.Sizeof(p.id)),
			(*byte)(unsafe.Pointer(&p.addr)),
			uint32(unsafe.Sizeof(p.addr)),
			&n,
			nil, // overlapped
			0,   // completionRoutine
		)
	})

	return p.err
}
