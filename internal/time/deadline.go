package time

import (
	"errors"
	"sync"
	"time"
)

type emptyCh = chan struct{}
type doneCh = <-chan struct{}

var ErrUninitialized = errors.New("cannot reset an uninitialized deadline")

// a resettable timer that can be viewed as a stripped down [context.Context]
type Deadline struct {
	mu sync.RWMutex //locks fields below
	// will be closed when timer expires
	ch emptyCh
	// time (unixnano) that timer will expire on
	// don't rely on t.when, since (1) that uses time.Time internals; and (2) it is manipulated during runtime
	when int64
	// allows only latest timer instance to reset ch
	// see comment in Reset() for more
	seq uintptr
	t   timer
}

// NewDeadline creates a [Deadline] that expires at [time.Time] t. If t is in the past or zero,
// the deadline never expires.
func NewDeadline(t time.Time) *Deadline {
	d := &Deadline{
		ch:   make(emptyCh),
		when: nano(t),
		t: timer{
			when: when(time.Until(t)),
			f:    closeCh,
		},
	}
	d.t.arg = d
	startTimer(&d.t)
	return d
}

// Reset changes when the [Deadline] will expire. If the [Deadline] is still active, then
// it will not signal pending operations and will remain active.
func (d *Deadline) Reset(t time.Time) error {
	if d == nil {
		return ErrUninitialized
	}

	d.mu.Lock()
	defer d.mu.Unlock()
	if d.t.f == nil {
		return ErrUninitialized
	}

	// there is a race here if the timer expires but the goroutine running closeCh is queued and runs
	// after this function releases mu.
	// however, resetting d.t would schedule closeCh to run again.
	// increment d.seq to prevent old instances from closing d.ch
	d.seq++
	d.when = nano(t)

	if !d.valid() {
		d.ch = make(emptyCh)
	}
	resetTimer(&d.t, when(time.Until(t)))

	return nil
}

func (d *Deadline) Done() doneCh {
	if d == nil {
		return make(emptyCh)
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.ch
}

func (d *Deadline) Deadline() (time.Time, bool) {
	if d == nil {
		return time.Time{}, false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()
	if d.when == 0 {
		return time.Time{}, false
	}
	return time.Unix(0, d.when), true
}

// checks if d.ch is not closed. caller must own d.mu
func (d *Deadline) valid() bool {
	select {
	case <-d.ch:
		return false
	default:
	}
	return true
}

// nano returns t as the unix time in nanoseconds, or zero if t.IsZero is true
func nano(t time.Time) int64 {
	if t.IsZero() {
		return 0
	}
	return t.UnixNano()
}

func closeCh(t any, seq uintptr) {
	d := t.(*Deadline)
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.seq == d.t.seq && d.valid() {
		close(d.ch)
	}
}
