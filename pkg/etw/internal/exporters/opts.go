//go:build windows

package exporters

import (
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// WithNewETWProvider registers a new ETW provider and sets the hook to log using it.
// The provider will be closed when the hook is closed.
func WithNewETWProvider[I Input](n string) Opt[I] {
	return func(c *Common[I]) (err error) {
		if c.Provider, err = etw.NewProvider(n, nil); err != nil {
			return err
		}
		c.CloseProvider = true
		return nil
	}
}

// WithExistingETWProvider configures the hook to use an existing ETW provider.
// The provider will not be closed when the hook is closed.
func WithExistingETWProvider[I Input](p *etw.Provider) Opt[I] {
	return func(e *Common[I]) error {
		e.Provider = p
		e.CloseProvider = false
		return nil
	}
}

// EnableActivityID sets if the ETW Activity ID should be set to the Span ID
func EnableActivityID[I Input](e *Common[I]) error {
	e.EnableActivityID = true
	return nil
}

// EnableRelatedActivityID sets if the ETW Activity ID should be set to the Span ID
func EnableRelatedActivityID[I Input](e *Common[I]) error {
	e.EnableRelatedActivityID = true
	return nil
}

// WithGetName sets the ETW EventName of an event to the value returned by f
// If the name is empty, the default event name will be used.
func WithGetName[I Input](f func(I) string) Opt[I] {
	return func(h *Common[I]) error {
		h.GetName = f
		return nil
	}
}


// WithEventOpts allows additional ETW event properties (keywords, tags, etc.) to be specified
func WithEventOpts[I Input](f func(I) []etw.EventOpt) Opt[I] {
	return func(e *Common[I]) error {
		e.GetEventsOpts = f
		return nil
	}
}

// WithTimeFormat sets how span start and stop time are formatted using [time.Format].
// Leave blank to encode in ETW's native format
func WithTimeFormat[I Input](f string) Opt[I] {
	return func(e *Common[I]) error {
		if f == "" {
			e.FormatTime = func(n string, t time.Time) etw.FieldOpt {
				return etw.Time(n, t)
			}
		} else {
			e.FormatTime = func(n string, t time.Time) etw.FieldOpt {
				return etw.StringField(n, t.Format(f))
			}
		}
		return nil
	}
}

// WithCustomTimeFormat allows configuring an alternative time format.
func WithCustomTimeFormat[I Input](f func(n string, t time.Time) etw.FieldOpt) Opt[I] {
	return func(e *Common[I]) error {
		e.FormatTime = f
		return nil
	}
}
