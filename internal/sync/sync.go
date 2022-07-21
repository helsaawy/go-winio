// This package contains custom synchronization primitives
package sync // "github.com/Microsoft/go-winio/internal/sync"

import (
	"sync/atomic"
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
