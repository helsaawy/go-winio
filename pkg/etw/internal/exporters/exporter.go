//go:build windows

package exporters

import (
	"context"
	"errors"
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// ErrNoProvider is returned when the exporter is created without a provider being configured.
var ErrNoProvider = errors.New("no ETW registered provider")

type Input any

// Opt is a configuration option for Common
type Opt[I Input] func(*Common[I]) error

// Common houses common fields for different exporters.
//
// It is not thread safe.
type Common[I Input] struct {
	Provider *etw.Provider
	// CloseProvider specifies if the ETW provider should be closed
	CloseProvider bool

	// set ETW ActivityID to SpanID
	EnableActivityID bool
	// set ETW RelatedActivityID to parent SpanID
	EnableRelatedActivityID bool

	// GetName returns the TaskName for the event
	GetName func(I) string
	// GetEventsOpts returns additional options to add to the event
	GetEventsOpts func(I) []etw.EventOpt
	// FormatTime sets the string as the field key and formats the time parameter
	FormatTime func(string, time.Time) etw.FieldOpt
}

// Validate errors if not ETW provider is configured, and sets non-configured (nil) options to a
// sensible default value
func (e *Common[I]) Validate() error {
	if e.Provider == nil {
		return ErrNoProvider
	}

	if e.GetName == nil {
		e.GetName = func(I) string {
			return ""
		}
	}
	if e.GetEventsOpts == nil {
		e.GetEventsOpts = func(I) []etw.EventOpt {
			return nil
		}
	}
	if e.FormatTime == nil {
		e.FormatTime = func(n string, t time.Time) etw.FieldOpt {
			return etw.Time(n, t)
		}
	}
	return nil
}

// Close cleans closes the [etw.Provider] returned by [Provider()] if [CloseProvider]
// is true.
func (e *Common[I]) Close(ctx context.Context) (err error) {
	closed := e.Provider == nil
	cp := e.CloseProvider

	if !closed {
		defer func() {
			e.Provider = nil
		}()
		if cp {
			if err := e.Provider.Close(); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

func (e *Common[I]) IsClosed() bool {
	return e.Provider == nil
}
