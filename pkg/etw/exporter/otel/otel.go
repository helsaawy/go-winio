//go:build windows

package otel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	// "go.opentelemetry.io/otel"
	// "go.opentelemetry.io/otel/propagation"
	// "go.opentelemetry.io/otel/sdk/resource"
	// "go.opentelemetry.io/otel/sdk/trace"
	// semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	common "github.com/Microsoft/go-winio/internal/etw/exporter"
	"github.com/Microsoft/go-winio/pkg/etw"
)

// WithExportEvents sets whether the Open Telemetry Exporter will export
func WithExportEvents(b bool) common.Opt {
	return func(e common.Exporter) error {
		switch ee := e.(type) {
		case *exporter:
			ee.exportEvents = b
		default:
			return common.ExporterTypeErr(e, &exporter{})
		}
		return nil
	}
}

type exporter struct {
	common.Common
	exportEvents bool
}

var _ trace.SpanExporter = &exporter{}

// NewExporter returns a [trace.SpanExporter] that exports Open Telemetry spans to ETW
// based on the the following rules:
//  * ETW entries will contain the Attributes, SpanKind, TraceID,
//   SpanID, and ParentSpanID.
//  * Annotation, MessageEvents, and Links will not be exported.
//  * The span itself will be written at [etw.LevelInfo], unless
//   `s.Status().Code == `, in which case it will be written at [etw.LevelError],
//   with the field `Error` set to Status.Message
func NewExporter(opts ...common.Opt) (trace.SpanExporter, error) {
	e := &exporter{}
	opts = append([]common.Opt{common.WithTimeFormat(time.RFC3339Nano)}, opts...)
	if err := common.InitExporter(e, opts...); err != nil {
		return nil, fmt.Errorf("create new OpenCensus exporter: %w", err)
	}
	return e, nil
}

func (e *exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		if err := ctx.Err(); err != nil {
			return err
		}
		if e.IsClosed() {
			return nil
		}

		level := etw.LevelInfo
		if span.Status().Code == codes.Error {
			level = etw.LevelError
		}
		if !e.Provider().IsEnabledForLevel(level) {
			continue
		}

		name := span.Name()
		sc := span.SpanContext()
		rsc := span.Resource()
		attrs := span.Attributes()
		events := span.Events()
		span.InstrumentationLibrary()
		span.Resource()
		if e.exportEvents {
			events = nil
		}

		// extra room for two more options in addition to log level to avoid reallocating
		// if the user also provides options
		opts := make([]etw.EventOpt, 0, 3)
		opts = append(opts, etw.WithLevel(level))
		opts = append(opts, e.GetEventsOpts(span)...)

		// Reserve extra space for the span properties .
		fields := make([]etw.FieldOpt, 0, 11+rsc.Len()+len(attrs))
		fields = append(fields,
			etw.StringField("trace-id", sc.TraceID().String()),
			etw.StringField("span-id", sc.SpanID().String()),
			etw.StringField("parent-span-id", span.Parent().SpanID().String()),
			e.FormatTime("start-time", span.StartTime()),
			e.FormatTime("end-time", span.EndTime()),
			etw.StringField("duration", span.EndTime().Sub(span.StartTime()).String()),
			etw.Uint32Field("span.status.code", uint32(span.Status().Code)),
			etw.StringField("span.status.error", span.Status().Description),
			etw.StringField("span.kind", span.SpanKind().String()),
		)

		fields = append(fields, resourcesToFields(rsc)...)

		fields = append(fields, attributesToFields(attrs)...)
		if n := span.DroppedAttributes(); n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_attributes", n))
		}

		if n := span.DroppedEvents(); n > 0 {
			fields = append(fields, etw.IntField("DroppedEvents", n))
		}

		// Firing an ETW event is essentially best effort, as the event write can
		// fail for reasons completely out of the control of the event writer (such
		// as a session listening for the event having no available space in its
		// buffers). Therefore, we don't return the error from WriteEvent, as it is
		// just noise in many cases.
		e.Provider().WriteEvent(name, opts, fields) //nolint:errcheck

		// events
		// for _, e := range events {
		// 	fields = append(fields, etw.SmartField(k, aMap[k].AsInterface()))
		// }
	}
	return nil
}

// var rscs = make(map[attribute.Distinct][]etw.FieldOpt, 0)

// map[attribute.Distinct][]etw.FieldOpt
var rscs = &sync.Map{}

// resource  keys are expected to conform to
// https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/
func resourcesToFields(rsc *resource.Resource) []etw.FieldOpt {
	rk := rsc.Equivalent()
	if opts, ok := rscs.Load(rk); ok {
		return opts.([]etw.FieldOpt)
	}

	i := rsc.Iter()
	fs := make([]etw.FieldOpt, i.Len())
	for i.Next() {
		kv := i.Attribute()
		if !kv.Key.Defined() {
			continue
		}

		kv.Key = attribute.Key("resource." + string(kv.Key))
		if f := keyValueToField(kv); f != nil {
			fs = append(fs, f)
		}
	}
	rscs.Store(rk, fs)

	return fs
}

func attributesToFields(attrs []attribute.KeyValue) []etw.FieldOpt {
	s := attribute.NewSet(attrs...)
	i := s.Iter()
	fs := make([]etw.FieldOpt, i.Len()+1)
	for i.Next() {
		if f := keyValueToField(i.Attribute()); f != nil {
			fs = append(fs, f)
		}
	}
	return fs
}

func keyValueToField(kv attribute.KeyValue) etw.FieldOpt {
	if !kv.Valid() {
		return nil
	}

	k := string(kv.Key)
	v := kv.Value
	switch kv.Value.Type() {
	case attribute.BOOL:
		return etw.BoolField(k, v.AsBool())
	case attribute.BOOLSLICE:
		return etw.BoolArray(k, v.AsBoolSlice())
	case attribute.INT64:
		return etw.Int64Field(k, v.AsInt64())
	case attribute.INT64SLICE:
		return etw.Int64Array(k, v.AsInt64Slice())
	case attribute.FLOAT64:
		return etw.Float64Field(k, v.AsFloat64())
	case attribute.FLOAT64SLICE:
		return etw.Float64Array(k, v.AsFloat64Slice())
	case attribute.STRING:
		return etw.StringField(k, v.AsString())
	case attribute.STRINGSLICE:
		return etw.StringArray(k, v.AsStringSlice())
	}
	return nil
}

func (e *exporter) Shutdown(ctx context.Context) error {
	return e.Close(ctx)
}
