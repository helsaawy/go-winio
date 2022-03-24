//go:build windows

package etwlogrus

import (
	"context"
	"strings"

	"github.com/sirupsen/logrus"
	octrace "go.opencensus.io/trace"
	ottrace "go.opentelemetry.io/otel/trace"

	"github.com/Microsoft/go-winio/pkg/etw"
	"github.com/Microsoft/go-winio/pkg/guid"
)

// etw provider

// WithNewETWProvider registers a new ETW provider and sets the hook to log using it.
// The provider will be closed when the hook is closed.
func WithNewETWProvider(n string) HookOpt {
	return func(h *Hook) error {
		provider, err := etw.NewProvider(n, nil)
		if err != nil {
			return err
		}

		h.provider = provider
		h.closeProvider = true
		return nil
	}
}

// WithExistingETWProvider configures the hook to use an existing ETW provider.
// The provider will not be closed when the hook is closed.
func WithExistingETWProvider(p *etw.Provider) HookOpt {
	return func(h *Hook) error {
		h.provider = p
		h.closeProvider = false
		return nil
	}
}

// event name

// WithDefaultEventName sets the ETW ActivityID of an event to the value returned by f
func WithDefaultEventName() HookOpt {
	return WithGetName(defaultEventName)
}

// WithOCSpanName tries to extract the OpenCensus Span name to use as ETW EventName.
// Otherwise, it uses the default event name (`etwlogrus.DefaultEventName`)
func WithOCSpanName() HookOpt {
	return WithGetName(tryNameFromSpanFunc(nameFromOCSpan))
}

// WithOTSpanName tries to extract the OpenTelemetry Span name to use as ETW EventName.
// Otherwise, it uses the default event name (`etwlogrus.DefaultEventName`)
func WithOTSpanName() HookOpt {
	return WithGetName(tryNameFromSpanFunc(nameFromOTSpan))
}

// WithGetName sets the ETW EventName of an event to the value returned by f
func WithGetName(f func(*logrus.Entry) string) HookOpt {
	return func(h *Hook) error {
		h.getName = f
		return nil
	}
}

// activity id

// WithOCSpanID tries to extract the OpenCensus Trace ID to use as the ETW ActivityID.
func WithOCSpanID() HookOpt {
	return WithGetID(tryIDFromSpanFunc(idFromOCSpan))
}

// WithOTSpanID tries to extract the OpenCensus Trace ID to use as the ETW ActivityID.
func WithOTSpanID() HookOpt {
	return WithGetID(tryIDFromSpanFunc(idFromOTSpan))
}

// WithGetID sets the ETW ActivityID of an event to the value returned by f
func WithGetID(f func(*logrus.Entry) guid.GUID) HookOpt {
	return func(h *Hook) error {
		h.getID = f
		return nil
	}
}

// WithAdditionalEventOpts allows additional ETW event properties (keywords, tags, etc.) to be specified
func WithAdditionalEventOpts(f func(*logrus.Entry) []etw.EventOpt) HookOpt {
	return func(h *Hook) error {
		h.getExtraEventsOpts = f
		return nil
	}
}

//
// option implementations
//

func defaultEventName(_ *logrus.Entry) string {
	return DefaultEventName
}

// tryNameFromSpanFunc returns a func that tries to extract the name from a span, or uses the default
func tryNameFromSpanFunc(f func(context.Context) (string, bool)) func(*logrus.Entry) string {
	return func(e *logrus.Entry) string {
		if ctx := e.Context; ctx != nil {
			if n, ok := f(ctx); ok {
				return n
			}
		}

		// entry could still be a span converted to a logrus entry
		if n, ok := getFirst(e.Data, "Name", "name"); ok && strings.ToLower(e.Message) == "span" {
			if nn, ok := n.(string); ok {
				return nn
			}
		}

		return DefaultEventName
	}
}

// nameFromOCSpan returns the name extracted from the OpenCensus Span, and a bool, if one was found
func nameFromOCSpan(ctx context.Context) (string, bool) {
	if span := octrace.FromContext(ctx); span.IsRecordingEvents() {
		// the OC trace API does not export a way to access a Span's name outside of `.String(`
		// current implementation returns either:
		//    fmt.Sprintf("span %s", s.spanContext.SpanID)
		//    fmt.Sprintf("span %s %q", s.spanContext.SpanID, s.data.Name)
		s := strings.SplitN(span.String(), " ", 3)
		if len(s) == 3 {
			n := s[2]
			// strip surrounding quotes
			if nn := len(n); nn >= 2 {
				if n[nn-1] == '"' {
					n = n[:nn-1]
				}
				if n[0] == '"' {
					n = n[1:]
				}
			}
			return n, true
		}
	}
	return "", false
}

// nameFromOTSpan returns the name extracted from the OpenCensus Span, and a bool, if one was found
func nameFromOTSpan(ctx context.Context) (string, bool) {
	if span := ottrace.SpanFromContext(ctx); span.IsRecording() {
		// The otel span interface doenst explicitly have a `Name()` function, but some
		// implementations (ie, in the /sdk/trace) do
		if ros, ok := span.(interface{ Name() string }); ok {
			return ros.Name(), true
		}
	}
	return "", false
}

// tryNameFromSpanFunc returns a func that tries to extract the span's trace ID, or returns an empty one
func tryIDFromSpanFunc(f func(context.Context) (guid.GUID, bool)) func(*logrus.Entry) guid.GUID {
	return func(e *logrus.Entry) guid.GUID {
		if ctx := e.Context; ctx != nil {
			if g, ok := f(ctx); ok {
				return g
			}
		}

		// check if the entry has a trace ID
		if n, ok := getFirst(e.Data, "traceID", "TraceID", "traceid", "trace-id"); ok {
			if nn, ok := n.(string); ok {
				if g, err := guid.FromString(nn); err == nil {
					return g
				}
			}
		}

		return guid.GUID{}
	}
}

func idFromOCSpan(ctx context.Context) (guid.GUID, bool) {
	if span := octrace.FromContext(ctx); span != nil {
		return guid.FromArray(span.SpanContext().TraceID), true
	}
	return guid.GUID{}, false
}

func idFromOTSpan(ctx context.Context) (guid.GUID, bool) {
	if sc := ottrace.SpanContextFromContext(ctx); sc.HasTraceID() {
		return guid.FromArray(sc.TraceID()), true
	}
	return guid.GUID{}, false
}

// getFirst returns the first value that exists in s, or returns with (nil, false)
func getFirst(d logrus.Fields, ks ...string) (v interface{}, ok bool) {
	for _, k := range ks {
		v, ok = d[k]
		if ok {
			break
		}
	}
	return v, ok
}
