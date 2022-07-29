// This package includes common primitives needed for testing, especially for asynchronous
// and concurrent tests.
package testutil

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

//
// helpers
//

// U wraps a testing struct ([testing.T], [testing.B], [testing.F], etc) and offers additional
// utility functions.
type U struct {
	T       testing.TB
	mu      sync.Mutex
	u       *U            //parent
	done    chan struct{} // test is finished
	timeout chan struct{} // test timed out
	tmr     time.Timer
}

func New(t testing.TB) *U {
	u := &U{
		T: t,
	}
	u.createDone()
	return u
}

func NewContext(ctx context.Context, t testing.TB) *U {
	u := &U{
		T: t,
	}
	u.createDone()
	return u
}

// Inherit returns a child testing utility that inherits the same timeout (if set) from
// the parent, u
func (u *U) Inherit(t testing.TB) *U {
	uu := &U{
		T: t,
		u: u,
	}
	uu.createDone()
	go uu.waitBackground()
	return uu
}

func (u *U) createDone() {
	u.done = make(chan struct{})
	u.T.Cleanup(func() {
		u.mu.Lock()
		defer u.mu.Unlock()
		safeClose(u.done)
	})
}

func (u *U) Done() <-chan struct{} {
	return u.done
}

func (u *U) Timeout() <-chan struct{} {
	return u.timeout
}

func (u *U) SetTimeout() {
	// TODO
}

// checks stops execution if testing failed in another go-routine
func (u *U) Check() {
	if u.T.Failed() {
		u.T.FailNow()
	}
}

func (u *U) Assert(b bool, msgs ...string) {
	if b {
		return
	}
	u.T.Helper()
	u.T.Fatalf(msgJoin(msgs, "failed assertion"))
}

func (u *U) Is(err, target error, msgs ...string) {
	if errors.Is(err, target) {
		return
	}
	u.T.Helper()
	u.T.Fatalf(msgJoin(msgs, "got error %q; wanted %q"), err, target)
}

func (u *U) Must(err error, msgs ...string) {
	if err == nil {
		return
	}
	u.T.Helper()
	u.T.Fatalf(msgJoin(msgs, "%v"), err)
}

func (u *U) Wait(ch <-chan struct{}, d time.Duration, msgs ...string) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ch:
	case <-t.C:
		u.T.Helper()
		u.T.Fatalf(msgJoin(msgs, "timed out after %v"), d)
	}
}

func msgJoin(pre []string, s string) string {
	return strings.Join(append(pre, s), ": ")
}

func safeClose(ch chan struct{}) {
	select {
	case <-ch:
	default:
		close(ch)
	}
}

func closed(ch <-chan struct{}) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}
