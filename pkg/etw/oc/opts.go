package oc

import (
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
	"go.opencensus.io/trace"
)

// WithNewETWProvider registers a new ETW provider and sets the hook to log using it.
// The provider will be closed when the hook is closed.
func WithNewETWProvider(n string) ExporterOpt {
	return func(e *exporter) (err error) {
		if e.provider, err = etw.NewProvider(n, nil); err != nil {
			return err
		}
		e.closeProvider = true
		return nil
	}
}

// WithExistingETWProvider configures the hook to use an existing ETW provider.
// The provider will not be closed when the hook is closed.
func WithExistingETWProvider(p *etw.Provider) ExporterOpt {
	return func(e *exporter) error {
		e.provider = p
		e.closeProvider = false
		return nil
	}
}

// WithEventOpts allows additional ETW event properties (keywords, tags, etc.) to be specified
func WithEventOpts(f func(*trace.SpanData) []etw.EventOpt) ExporterOpt {
	return func(e *exporter) error {
		e.getEventsOpts = f
		return nil
	}
}

// WithTimeFormat sets how span start and stop time are formatted using [time.Format].
// Leave blank to encode in ETW's native format
func WithTimeFormat(f string) ExporterOpt {
	return func(e *exporter) error {
		if f == "" {
			e.formatTime = func(n string, t time.Time) etw.FieldOpt {
				return etw.Time(n, t)
			}
		} else {
			e.formatTime = func(n string, t time.Time) etw.FieldOpt {
				return etw.StringField(n, t.Format(f))
			}
		}
		return nil
	}
}
