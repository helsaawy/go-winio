// This package contains custom synchronization primitives
package sync // "github.com/Microsoft/go-winio/internal/sync"

import (
	"sync"
	"sync/atomic"

	"golang.org/x/sys/windows"
)

// todo: replace with "sync/atomic".Bool in go1.19
// this implementation is based off of sync atomic's new bool
// https://cs.opensource.google/go/go/+/refs/tags/go1.19rc2:src/sync/atomic/type.go;l=11
type AtomicBool struct {
	_ NoCopy

	v uint32
}

func (x *AtomicBool) IsSet() bool      { return atomic.LoadUint32(&x.v) != 0 }
func (x *AtomicBool) SetFalse()        { x.Store(false) }
func (x *AtomicBool) SetTrue()         { x.Store(true) }
func (x *AtomicBool) Store(v bool)     { atomic.StoreUint32(&x.v, bu32(v)) }
func (x *AtomicBool) Swap(v bool) bool { return atomic.SwapUint32(&x.v, bu32(v)) == 1 }
func (x *AtomicBool) CompareAndSwap(old, new bool) (swapped bool) { //nolint:predeclared
	return atomic.CompareAndSwapUint32(&x.v, bu32(old), bu32(new))
}

// directly copied from
// https://cs.opensource.google/go/go/+/refs/tags/go1.19rc2:src/sync/atomic/type.go;l=31-36
func bu32(b bool) (u uint32) {
	if b {
		u = 1
	}
	return u
}

// AtomicHandle is a wrapper around [atomic/sync] Load/Store/SwapUintptr functions for
// [golang.org/x/sys/windows.Handle], with additional symantics:
//  - the value is valid if it does not equal [golang.org/x/sys/windows.InvalidHandle]
//  - the value is set if it does not equal windows.Handle(0)
type AtomicHandle struct {
	_ NoCopy

	v uintptr
}

func (x *AtomicHandle) Load() windows.Handle   { return windows.Handle(atomic.LoadUintptr(&x.v)) }
func (x *AtomicHandle) Store(v windows.Handle) { atomic.StoreUintptr(&x.v, uintptr(v)) }
func (x *AtomicHandle) Swap(v windows.Handle) windows.Handle {
	return windows.Handle(atomic.SwapUintptr(&x.v, uintptr(v)))
}
func (x *AtomicHandle) CompareAndSwap(old, new windows.Handle) (swapped bool) { //nolint:predeclared
	return atomic.CompareAndSwapUintptr(&x.v, uintptr(old), uintptr(new))
}

// LazyHandle is a Handle that is initialized on the first call, using the same
// semantics as [AtomicHandle]. It is based off of [golang.org/x/sys/windows.LazyProc].
type LazyHandle struct {
	_ NoCopy

	f    func()
	once sync.Once
	h    windows.Handle
	err  error
}

// NewLazyHandle allows delaying calling function f to load a Handle until the first use.
// The function, f, will be called only once if it returns a non-zero Handle even on failure.
//
// Otherwise, if f returns windows.Handle(0), then subsequent calls to [LazyHandle.Load] or
// [LazyHandle.Handle] will attempt to re-load and initialize the Handle value.
func NewLazyHandle(f func() (windows.Handle, error)) *LazyHandle {
	x := &LazyHandle{}
	x.f = func() {
		x.h, x.err = f()
	}
	return x
}

func (x *LazyHandle) Load() error {
	x.once.Do(x.f)
	return x.err
}

func (x *LazyHandle) mustLoad() {
	if err := x.Load(); err != nil {
		panic(err)
	}
}

func (x *LazyHandle) Handle() windows.Handle {
	x.mustLoad()
	return x.h
}

// NoCopy can be added which must not be copied after the first use.
//
// Requires checking via `go vet -copylocks`
//
// copied from sync/atomic/type.go
// https://cs.opensource.google/go/go/+/refs/tags/go1.19rc2:src/sync/atomic/type.go;l=182
type NoCopy struct{}

// Lock is a no-op used by -copylocks checker from `go vet`.
func (*NoCopy) Lock()   {}
func (*NoCopy) Unlock() {}
