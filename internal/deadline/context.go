package deadline

import (
	"context"
	"time"
)

// todo: remove in go1.18
//nolint:predeclared
type any = interface{}

// fakeContext is used to convert [Deadline] into a [context.Context], to propagate
// its timeout to child contexts.
//
// See [Deadline.Context] for more information.
type fakeContext struct {
	ch doneCh
}

var _ context.Context = &fakeContext{}

func (c *fakeContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (c *fakeContext) Done() <-chan struct{} {
	return c.ch
}
func (c *fakeContext) Err() error {
	select {
	case <-c.ch:
		return context.DeadlineExceeded
	default:
	}
	return nil
}

func (*fakeContext) Value(key any) any {
	return nil
}
