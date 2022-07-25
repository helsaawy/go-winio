package testutil // "github.com/Microsoft/go-winio/internal/testutil"

import (
	"errors"
	"strings"
	"testing"
	"time"
)

// currently TestUtil is just an interface, which is just one word more than a
// pointer. So rather a value receiver for now

type TestUtil struct {
	T testing.TB
}

func New(t testing.TB) TestUtil {
	return TestUtil{
		T: t,
	}
}

// checks stops execution if testing failed in another go-routine
func (u TestUtil) Check() {
	if u.T.Failed() {
		u.T.FailNow()
	}
}

func (u TestUtil) Assert(b bool, msgs ...string) {
	if !b {
		u.T.Helper()
		u.T.Fatalf(_msgJoin(msgs, "failed assertion"))
	}
}

func (u TestUtil) Is(err, target error, msgs ...string) {
	if !errors.Is(err, target) {
		u.T.Helper()
		u.T.Fatalf(_msgJoin(msgs, "got error %q; wanted %q"), err, target)
	}
}

func (u TestUtil) Must(err error, msgs ...string) {
	if err != nil {
		u.T.Helper()
		u.T.Fatalf(_msgJoin(msgs, "%v"), err)
	}
}

func (u TestUtil) Wait(ch <-chan struct{}, d time.Duration, msgs ...string) {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ch:
	case <-t.C:
		u.T.Helper()
		u.T.Fatalf(_msgJoin(msgs, "timed out after %v"), d)
	}
}

func _msgJoin(pre []string, s string) string {
	return strings.Join(append(pre, s), ": ")
}
