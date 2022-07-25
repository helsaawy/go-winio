package deadline

import (
	"testing"
	"time"

	"github.com/Microsoft/go-winio/internal/testutil"
)

// long enough to allow things to happen
const waitTime = 5 * time.Millisecond

func TestEmpty(t *testing.T) {
	u := testutil.New(t)
	d := Empty()

	dd, ok := d.Deadline()
	u.Assert(!ok, "empty deadline should be false")
	u.Assert(dd.IsZero(), "deadline should be zero")

	select {
	case <-d.Done():
		t.Fatal("empty deadline should not expire")
	case <-time.After(waitTime):
	}

	u.Must(d.Err(), "empty deadline should be nil")

	t.Run("stop", func(t *testing.T) {
		u.Must(d.Stop(), "stop should not error")

		dd, ok = d.Deadline()
		u.Assert(ok, "stopped deadline should be true")
		u.Assert(!dd.IsZero(), "deadline should not be zero")
		// d.Deadline does not have monotonic time, so add a [ns] to break ties
		u.Assert(time.Now().Add(time.Nanosecond).After(dd), "deadline should be past")
		u.Is(d.Err(), ErrDeadlineExceeded, "stopped deadline should be cancelled")

		select {
		case <-d.Done():
		default:
			t.Fatal("empty deadline should be expired")
		}
	})

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		u.Must(d.Reset(start.Add(waitTime)), "")
		testActiveDeadline(t, d, start, waitTime)
	})
}

func TestEmptyReset(t *testing.T) {
	u := testutil.New(t)
	d := Empty()

	dd, ok := d.Deadline()
	u.Assert(!ok, "empty deadline should be false")
	u.Assert(dd.IsZero(), "deadline should be zero")

	select {
	case <-d.Done():
		t.Fatal("empty deadline should not expire")
	case <-time.After(waitTime):
	}

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		u.Must(d.Reset(start.Add(waitTime)), "")
		testActiveDeadline(t, d, start, waitTime)
	})
}

func TestNew(t *testing.T) {
	u := testutil.New(t)
	start := time.Now()
	d := New(start.Add(waitTime))
	testActiveDeadline(t, d, start, waitTime)

	t.Run("reset", func(t *testing.T) {
		start := time.Now()
		dur := time.Millisecond
		u.Must(d.Reset(start.Add(dur)), "")
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
	u := testutil.New(t)
	dd, ok := d.Deadline()
	u.Assert(ok, "deadline should be true")
	u.Assert(!dd.IsZero(), "deadline should not be zero")
	u.Assert(time.Now().Before(dd), "deadline should be in the future")
	u.Assert(dd.Equal(start.Add(dur)), "deadline should be in the future")
	u.Must(d.Err(), "deadline should not be expired")

	select {
	case <-d.Done():
		t.Fatal(" deadline should not be expired")
	default:
	}

	<-d.Done()
	end := time.Now()
	u.Assert(end.Sub(start) >= dur, "did not wait full duration")

	dd, ok = d.Deadline()
	u.Assert(ok, "deadline should be true")
	u.Assert(!dd.IsZero(), "deadline should not be zero")
	u.Assert(end.After(dd), "deadline should before end time")
	u.Is(d.Err(), ErrDeadlineExceeded, "expired deadline should Err")
}
