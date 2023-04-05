//go:build windows

package etw

import (
	"context"
	"errors"

	tracesdk "go.opentelemetry.io/otel/sdk/trace"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// TODO: create "github.com/Microsoft/go-winio/pkg/otel/exporters" module for common code

// ErrNoProvider is returned when a hook is created without a provider being configured.
var ErrNoProvider = errors.New("no ETW registered provider")

// not thread-safe

type exporter struct {
	p             *etw.Provider
	closeProvider bool
}

var _ tracesdk.SpanExporter = &exporter{}

func New(opts ...Opt) (tracesdk.SpanExporter, error) {
	e := &exporter{}

	for _, o := range opts {
		if err := o(e); err != nil {
			return nil, err
		}
	}

	if err := e.validate(); err != nil {
		return nil, err
	}
	return e, nil
}

func (e *exporter) validate() error {
	if e.p == nil {
		return ErrNoProvider
	}
	return nil
}

func (e *exporter) ExportSpans(ctx context.Context, spans []tracesdk.ReadOnlySpan) error {
	// TODO
	for _, s := range spans {
		if err := ctx.Err(); err != nil {
			return err
		}

		s.Name()

	}
	return nil
}

func (e *exporter) Shutdown(_ context.Context) error {
	if !e.closeProvider || e.p == nil {
		return nil
	}
	err := e.p.Close()
	e.p = nil
	return err
}
