package otel

import (
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
	"github.com/Microsoft/go-winio/pkg/etw/internal/exporters"
)

// EnableEventExport sets whether the Open Telemetry Exporter will export
func EnableEventExport(e *exporter) error {
	e.exportEvents = true
	return nil
}

// WithNewETWProvider registers a new ETW provider and sets the hook to log using it.
// The provider will be closed when the hook is closed.
func WithNewETWProvider(n string) Opt {
	return func(e *exporter) error {
		return exporters.WithNewETWProvider[input](n)(&e.c)
	}
}

// WithExistingETWProvider configures the hook to use an existing ETW provider.
// The provider will not be closed when the hook is closed.
func WithExistingETWProvider(p *etw.Provider) Opt {
	return func(e *exporter) error {
		return exporters.WithExistingETWProvider[input](p)(&e.c)
	}
}

// EnableActivityID sets if the ETW Activity ID should be set to the Span ID
func EnableActivityID(e *exporter) error {
	return exporters.EnableActivityID[input](&e.c)
}

// EnableRelatedActivityID sets if the ETW Activity ID should be set to the Span ID
func EnableRelatedActivityID(e *exporter) error {
	return exporters.EnableRelatedActivityID[input](&e.c)
}

// WithGetName sets the ETW EventName of an event to the value returned by f
// If the name is empty, the default event name will be used.
func WithGetName(f func(input) string) Opt {
	return func(e *exporter) error {
		return exporters.WithGetName[input](f)(&e.c)
	}
}

// WithEventOpts allows additional ETW event properties (keywords, tags, etc.) to be specified
func WithEventOpts(f func(input) []etw.EventOpt) Opt {
	return func(e *exporter) error {
		return exporters.WithEventOpts[input](f)(&e.c)
	}
}

// WithTimeFormat sets how span start and stop time are formatted using [time.Format].
// Leave blank to encode in ETW's native format
func WithTimeFormat(f string) Opt {
	return func(e *exporter) error {
		return exporters.WithTimeFormat[input](f)(&e.c)
	}
}

// WithCustomTimeFormat allows configuring an alternative time format.
func WithCustomTimeFormat(f func(n string, t time.Time) etw.FieldOpt) Opt {
	return func(e *exporter) error {
		return exporters.WithCustomTimeFormat[input](f)(&e.c)
	}
}
