//go:build windows

package sync

import (
	"sync"
	"sync/atomic"

	"golang.org/x/sys/windows"
)

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
