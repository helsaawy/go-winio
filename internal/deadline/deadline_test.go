package deadline

import (
	"errors"
	"testing"
	"time"
)

// long enough to allow things to happen
const waitTime = 5 * time.Millisecond

func TestEmpty(t *testing.T) {
	d := Empty()

	dd, ok := d.Deadline()
	assert(t, !ok, "empty deadline should be false")
	assert(t, dd.IsZero(), "deadline should be zero")

	select {
	case <-d.Done():
		t.Fatal("empty deadline should not expire")
	case <-time.After(waitTime):
	}

	must(t, d.Err(), "empty deadline should be nil")

	t.Run("stop", func(t *testing.T) {
		must(t, d.Stop(), "stop should not error")

		dd, ok = d.Deadline()
		assert(t, ok, "stopped deadline should be true")
		assert(t, !dd.IsZero(), "deadline should not be zero")
		// d.Deadline does not have monotonic time, so add a [ns] to break ties
		assert(t, time.Now().Add(time.Nanosecond).After(dd), "deadline should be past")
		is(t, d.Err(), ErrDeadlineExceeded, "stopped deadline should be cancelled")

		select {
		case <-d.Done():
		default:
			t.Fatal("empty deadline should be expired")
		}
	})

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		must(t, d.Reset(start.Add(waitTime)), "")
		testActiveDeadline(t, d, start, waitTime)
	})
}

func TestEmptyReset(t *testing.T) {
	d := Empty()

	dd, ok := d.Deadline()
	assert(t, !ok, "empty deadline should be false")
	assert(t, dd.IsZero(), "deadline should be zero")

	select {
	case <-d.Done():
		t.Fatal("empty deadline should not expire")
	case <-time.After(waitTime):
	}

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		must(t, d.Reset(start.Add(waitTime)), "")
		testActiveDeadline(t, d, start, waitTime)
	})
}

func TestNew(t *testing.T) {
	start := time.Now()
	d := New(start.Add(waitTime))
	testActiveDeadline(t, d, start, waitTime)

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		dur := time.Millisecond
		must(t, d.Reset(start.Add(dur)), "")
		testActiveDeadline(t, d, start, dur)
	})

	t.Run("reset+stop", func(t *testing.T) {
	})
}

// reset active deadline to zero
func TestResetZero(t *testing.T) {
}

// test extending a current deadline to the future
func TestResetActive(t *testing.T) {
}

// test deadline set to the past
func TestNewPast(t *testing.T) {
}

func TestContext(t *testing.T) {
}

func testActiveDeadline(t testing.TB, d *Deadline, start time.Time, dur time.Duration) {
	dd, ok := d.Deadline()
	assert(t, ok, "deadline should be true")
	assert(t, !dd.IsZero(), "deadline should not be zero")
	assert(t, time.Now().Before(dd), "deadline should be in the future")
	assert(t, dd.Equal(start.Add(dur)), "deadline should be in the future")
	must(t, d.Err(), "deadline should not be expired")

	select {
	case <-d.Done():
		t.Fatal(" deadline should not be expired")
	default:
	}

	<-d.Done()
	end := time.Now()
	assert(t, end.Sub(start) >= dur, "did not wait full duration")

	dd, ok = d.Deadline()
	assert(t, ok, "deadline should be true")
	assert(t, !dd.IsZero(), "deadline should not be zero")
	assert(t, end.After(dd), "deadline should before end time")
	is(t, d.Err(), ErrDeadlineExceeded, "expired deadline should Err")
}

func assert(t testing.TB, b bool, msg string) {
	if msg != "" {
		msg = ": " + msg
	}
	if !b {
		t.Helper()
		t.Fatalf("assertion failed" + msg)
	}
}

func is(t testing.TB, err, target error, msg string) {
	if errors.Is(err, target) {
		return
	}
	t.Helper()
	if msg != "" {
		msg += ": "
	}
	t.Fatalf(msg+"got %q, wanted %q", err, target)
}

func must(t testing.TB, err error, msg string) {
	if err == nil {
		return
	}
	t.Helper()
	if msg != "" {
		msg += ": "
	}
	t.Fatalf(msg+"%v", err)
}
