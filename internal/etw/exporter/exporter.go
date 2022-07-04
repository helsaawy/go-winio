//go:build windows

package exporter

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// ErrNoProvider is returned when the exporter is created without a provider being configured.
var ErrNoProvider = errors.New("no ETW registered provider")

// todo(helsaawy): with go1.18, remove need for Input and swap to:
//	type Common[I any] struct {
//		GetName      func(I) string
//		GetEventOpts func(I) []etw.EventOpts
//	}

// Input is the data the Exporter will be exporting to ETW
type Input interface{}

// Opt is a configuration option for an Exporter
type Opt func(Exporter) error

// Exporter is a dummy interface to specify that a type is an ETW Exporter, so that
// [Opt] is a is not specified as:
//
//	Opt func(interface{}) error
//
// The intended use is to embed [Common] into a struct
//
// This is a workaround until we switch to go 1.18 and can write:
//
//	New[Opt (func(*Common) error) | (func(*MyExporter) error)](opts ...Opt) {
//		e := MyExporter{}
//		for _, o := range opts {
//			switch oo := any(o).(type) {
//			case func(*Common) error:
//				oo(&e.Common)
//			case func(*MyExporter) error:
//				oo(&e)
//			}
//		}
//	}
type Exporter interface {
	// Validate ensures the exporter is properly configured
	Validate() error
	Provider() *etw.Provider
	// GetName returns the TaskName for the event
	GetName(Input) string
	// GetEventsOpts returns additional options to add to the event
	GetEventsOpts(Input) []etw.EventOpt
	// FormatTime sets the string as the field key and formats the time parameter
	FormatTime(string, time.Time) etw.FieldOpt

	getCommon() *Common
}

// InitExporter initializes the [Exporter] with the options and then calls validate on
// the result.
func InitExporter(e Exporter, opts ...Opt) error {
	for _, o := range opts {
		if err := o(e); err != nil {
			return err
		}
	}
	if err := e.Validate(); err != nil {
		return err
	}
	return nil
}

// Common houses common fields for different exporters.
//
// It is not thread safe.
type Common struct {
	mu       sync.RWMutex // locks provider and closeProvider
	provider *etw.Provider
	// CloseProvider specifies if the ETW provider should be closed
	closeProvider bool

	// GetName returns the TaskName for the event
	getName func(Input) string
	// GetEventsOpts returns additional options to add to the event
	getEventsOpts func(Input) []etw.EventOpt
	// FormatTime sets the string as the field key and formats the time parameter
	formatTime func(string, time.Time) etw.FieldOpt
}

var _ Exporter = &Common{}

// Validate errors if not ETW provider is configured, and sets non-configured (nil) options to a
// sensible default value
func (e *Common) Validate() error {
	if e.provider == nil {
		return ErrNoProvider
	}

	if e.getName == nil {
		e.getName = func(Input) string {
			return ""
		}
	}
	if e.getEventsOpts == nil {
		e.getEventsOpts = func(Input) []etw.EventOpt {
			return nil
		}
	}
	if e.formatTime == nil {
		e.formatTime = func(n string, t time.Time) etw.FieldOpt {
			return etw.Time(n, t)
		}
	}
	return nil
}

func (e *Common) Provider() *etw.Provider {
	e.mu.RLock()
	defer e.mu.RUnlock()
	return e.provider
}

func (e *Common) GetName(i Input) string {
	return e.getName(i)
}

func (e *Common) GetEventsOpts(i Input) []etw.EventOpt {
	return e.getEventsOpts(i)
}

func (e *Common) FormatTime(n string, t time.Time) etw.FieldOpt {
	return e.formatTime(n, t)
}

// implementation of [Exporter], so any type that embeds [Common] will implement this
func (e *Common) getCommon() *Common {
	return e
}

// Close cleans closes the [etw.Provider] returned by [Provider()] if [CloseProvider]
// is true.
func (e *Common) Close(ctx context.Context) (err error) {
	e.mu.RLock()
	closed := e.provider == nil
	cp := e.closeProvider
	e.mu.RUnlock()

	if !closed {
		e.mu.Lock()
		defer e.mu.Unlock()
		defer func() {
			e.provider = nil
		}()
		if cp {
			if err := e.provider.Close(); err != nil {
				return err
			}
		}
	}
	return ctx.Err()
}

func (e *Common) IsClosed() bool {
	return e.Provider() == nil
}
