//go:build windows

package exporter

import (
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// WithNewETWProvider registers a new ETW provider and sets the hook to log using it.
// The provider will be closed when the hook is closed.
func WithNewETWProvider(n string) Opt {
	return newOpt(func(c *Common) (err error) {
		if c.Provider, err = etw.NewProvider(n, nil); err != nil {
			return err
		}
		c.CloseProvider = true
		return nil
	})
}

// WithExistingETWProvider configures the hook to use an existing ETW provider.
// The provider will not be closed when the hook is closed.
func WithExistingETWProvider(p *etw.Provider) Opt {
	return newOpt(func(e *Common) error {
		e.Provider = p
		e.CloseProvider = false
		return nil
	})
}

// WithGetName sets the ETW EventName of an event to the value returned by f
// If the name is empty, the default event name will be used.
func WithGetName(f func(Input) string) Opt {
	return newOpt(func(h *Common) error {
		h.GetName = f
		return nil
	})
}

// WithEventOpts allows additional ETW event properties (keywords, tags, etc.) to be specified
func WithEventOpts(f func(Input) []etw.EventOpt) Opt {
	return newOpt(func(e *Common) error {
		e.GetEventsOpts = f
		return nil
	})
}

// WithTimeFormat sets how span start and stop time are formatted using [time.Format].
// Leave blank to encode in ETW's native format
func WithTimeFormat(f string) Opt {
	return newOpt(func(e *Common) error {
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
	})
}

func newOpt(f func(*Common) error) Opt {
	return func(e Exporter) error {
		return f(e.getCommon())
	}
}
