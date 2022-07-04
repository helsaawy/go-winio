//go:build windows

package exporter

import (
	"time"

	"github.com/Microsoft/go-winio/pkg/etw"
)

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
	getCommon() *Common
}

// Common houses common fields for different exporters.
//
// It is not thread safe.
type Common struct {
	Provider *etw.Provider
	// CloseProvider specifies if the ETW provider should be closed
	CloseProvider bool
	// GetName returns the TaskName for the event
	GetName func(Input) string
	// GetEventsOpts returns additional options to add to the event
	GetEventsOpts func(Input) []etw.EventOpt
	// FormatTime sets the string as the field key and formats the time parameter
	FormatTime func(string, time.Time) etw.FieldOpt
}

// implementation of [Exporter], so any type that embeds [Common] will implement this
func (c *Common) getCommon() *Common {
	return c
}

// Close cleans closes the [etw.Provider] returned by [Provider()] if [CloseProvider]
// is true.
func (c *Common) Close() error {
	defer func() {
		c.Provider = nil
	}()
	if c.CloseProvider {
		return c.Provider.Close()
	}
	return nil
}
