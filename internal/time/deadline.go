package time

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

type emptyCh = chan struct{}
type doneCh = <-chan struct{}

var ErrUninitialized = errors.New("cannot reset an uninitialized deadline")

type Timeout interface {
	Deadline() (time.Time, bool)
	Done() doneCh
}

var _ Timeout = context.Background()

// Deadline is a resettable timer.
// Must be created via [NewDeadline].
// To wait on a deadline, use [<-Deadline.Done()].
//
// A Deadline does not distinguish between a reaching the specified deadline and calling
// stop.
//
// Resetting an unexpired Deadline does not cancel it, but updates the deadline for all
// pending and future operations waiting on [Deadline.Done].
type Deadline struct {
	mu sync.RWMutex //locks fields below
	// will be closed when timer expires, and recreated by `Reset`
	ch emptyCh
	// unix nanosecond time stamp that timer will expire on; zero if deadline never expires.
	// See comment in Reset() for more.
	//
	// Ideally, this would be internal representation used by go (`runtime.nanotime`), but
	// we cannot access that without using `//go:linkname` and `unsafe`.
	nano int64
	t    *time.Timer
}

var _ Timeout = &Deadline{}

// NewDeadline creates a [Deadline] that expires at [time.Time] t. If t is zero,
// the [Deadline] never expires. If t is in the past, the Deadline executes immediately.
func NewDeadline(t time.Time) *Deadline {
	d := &Deadline{
		ch:   make(emptyCh),
		nano: nano(t),
	}
	// initialize regardles of t only if !t.IsZero
	if d.nano != 0 {
		d.t = time.AfterFunc(time.Until(t), d.closeCh)
	}
	return d
}

// Cancels the deadline, stopping the underlying timer and notifying waiters.
// This function updates the value returned by [Deadline.Deadline] and waits until
// [Deadline.Done] returns.
func (d *Deadline) Stop() error {
	if d == nil {
		return fmt.Errorf("cannot stop an uninitialized deadline")
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	// update nano so that closeCh can execute
	d.nano = now()
	d.t.Stop()
	<-d.ch
	return nil
}

// Reset changes when the [Deadline] will expire. If the [Deadline] is still active, then
// it will not signal pending operations and will remain active.
func (d *Deadline) Reset(t time.Time) error {
	if d == nil {
		return fmt.Errorf("cannot reset an uninitialized deadline")
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	d.nano = nano(t)
	if closed(d.ch) {
		d.ch = make(emptyCh)
	}

	if d.nano == 0 {
		// t.IsZero(); deadline does not expire
		if d.t != nil {
			d.t.Stop()
		}
		return nil
	}

	if d.t == nil {
		d.t = time.AfterFunc(time.Until(t), d.closeCh)
		return nil
	}

	// There is a race where the timer expires but the goroutine running d.closeCh is
	// scheduled after this function returns (and releases mu).
	// This would lead to a stale reset of d.ch.
	// Prevent this by comparing the current timestamp with d.nano inside d.closeCh
	//
	// A better solution would be to use `time.runtimeTimer` (or `runtime.timer` (in time.go))
	// directly and increment `timer.seq` here (similar to `runtime.pollDesc` (in netpoll.go)).
	// That way, our `closeCh` could accept the the `seq` input (which `time.goFunc` ignores)
	// and prevent stale resets without needing to check timestamps.
	d.t.Reset(time.Until(t))
	return nil
}

func (d *Deadline) closeCh() {
	d.mu.Lock()
	defer d.mu.Unlock()

	// Only close d.ch if the deadline (d.nano) is non-zero (ie, is a valid deadline) and the
	// current time is after the deadline.
	// See `Reset` for more.

	// `time.Now().UnixNano() >= d.nano` should be faster than `time.Now().After(time.Unix(0, d.nano))`
	if d.nano != 0 && now() >= d.nano && !closed(d.ch) {
		close(d.ch)
	}
}

//
// context.Context implementation
//
// since AsyncIO may be passed nil, have these functions accept nil receivers

func (d *Deadline) Done() doneCh {
	if d == nil {
		// waits on nil channel wil block forever
		return nil
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.ch
}

func (d *Deadline) Deadline() (t time.Time, ok bool) {
	if d == nil {
		return t, ok
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.nano == 0 {
		return t, ok
	}
	return time.Unix(0, d.nano), true
}

// checks if ch is closed
func closed(ch emptyCh) bool {
	select {
	case <-ch:
		return true
	default:
	}
	return false
}

// nano returns t as the unix time in nanoseconds, or zero if t.IsZero is true
func nano(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}

// now returns the current time as a unix nanosecond timestamp
func now() int64 {
	return time.Now().UnixNano()
}
