//go:build windows

package otel

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/sdk/instrumentation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	oteltrace "go.opentelemetry.io/otel/trace"

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

// although otel guarantees that the SpanExporter will be called synchronously,
// enforce thread safety in case multiple TraceProviders use the same exporter

// todo: find a way to allow

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
//   `s.Status().Code == codes.Error`, in which case it will be written at [etw.LevelError],
//   with the field `otel.status_description` set to Status().Description
//
// This exporter ignores the [WithGetName] option and sets the ETW task name to the span name.
// Additionally, the function specified in [WithEventOpts] is expected to work on both
// [trace.ReadOnlySpan], and if [WithExportEvents] is set to true, [trace.Event] as well.
func NewExporter(opts ...common.Opt) (trace.SpanExporter, error) {
	e := &exporter{}
	opts = append([]common.Opt{common.WithTimeFormat(time.RFC3339Nano)}, opts...)
	if err := common.InitExporter(e, opts...); err != nil {
		return nil, fmt.Errorf("create new OpenCensus exporter: %w", err)
	}
	return e, nil
}

// ETW field names conform to the OTel attribute naming specification and their recommended mappings:
// https://opentelemetry.io/docs/reference/specification/common/attribute-naming/
// https://opentelemetry.io/docs/reference/specification/common/mapping-to-non-otlp/

func (e *exporter) ExportSpans(ctx context.Context, spans []trace.ReadOnlySpan) error {
	for _, span := range spans {
		if err := ctx.Err(); err != nil {
			return err
		}
		if e.IsClosed() {
			return nil
		}

		level := etw.LevelInfo
		st := span.Status()
		if st.Code == codes.Error {
			level = etw.LevelError
		}
		if !e.Provider().IsEnabledForLevel(level) {
			continue
		}

		name := span.Name()
		sc := span.SpanContext()
		rsc := span.Resource()
		il := span.InstrumentationLibrary()
		attrs := span.Attributes()

		// extra room for two more options in addition to log level to avoid reallocating
		// if the user also provides options
		opts := make([]etw.EventOpt, 0, 3)
		opts = append(opts, etw.WithLevel(level))
		opts = append(opts, e.GetEventsOpts(span)...)

		// Reserve extra space for the span properties .
		fields := make([]etw.FieldOpt, 0, 9+rsc.Len()+2+len(attrs)+3)
		fields = append(fields,
			etw.StringField("span.trace_id", sc.TraceID().String()),
			etw.StringField("span.parent_id", span.Parent().SpanID().String()),
			etw.StringField("span.id", sc.SpanID().String()),
			e.FormatTime("span.start_time", span.StartTime()),
			e.FormatTime("span.end_time", span.EndTime()),
			etw.StringField("span.duration", span.EndTime().Sub(span.StartTime()).String()),
		)

		if st.Code != codes.Unset {
			fields = append(fields, etw.StringField("otel.status_code", span.Status().Code.String()))
			if st.Description != "" {
				fields = append(fields, etw.StringField("span.status_description", span.Status().Description))
			}
		}

		if sk := span.SpanKind(); sk != oteltrace.SpanKindUnspecified && sk != oteltrace.SpanKindInternal {
			fields = append(fields, etw.StringField("span.kind", sk.String()))
		}

		fields = append(fields, resourcesToFields(rsc)...)
		fields = append(fields, libToFields(il)...)

		fields = append(fields, attributesToFields(attrs)...)

		if n := span.DroppedAttributes(); n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_attributes_count", n))
		}
		if n := span.DroppedEvents(); n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_events_count", n))
		}
		if n := span.DroppedLinks(); n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_links_count", n))
		}

		if err := e.Provider().WriteEvent(name, opts, fields); err != nil {
			return err
		}

		return e.exportSpanEvents(ctx, span)
	}
	return nil
}

// var rscs = make(map[attribute.Distinct][]etw.FieldOpt, 0)

// map[attribute.Distinct][]etw.FieldOpt
var rscs = &sync.Map{}

// resource keys are expected to conform to
// https://opentelemetry.io/docs/reference/specification/resource/semantic_conventions/
func resourcesToFields(rsc *resource.Resource) []etw.FieldOpt {
	rk := rsc.Equivalent()
	if opts, ok := rscs.Load(rk); ok {
		return opts.([]etw.FieldOpt)
	}

	i := rsc.Iter()
	fs := make([]etw.FieldOpt, 0, i.Len())
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

// map[instrumentation.Library][]etw.FieldOpt
var ils = &sync.Map{}

func libToFields(il instrumentation.Library) []etw.FieldOpt {
	if opts, ok := ils.Load(il); ok {
		return opts.([]etw.FieldOpt)
	}
	fs := []etw.FieldOpt{
		etw.StringField("otel.scope.name", il.Name),
		etw.StringField("otel.scope.version", il.Version),
	}
	ils.Store(il, fs)
	return fs
}

func attributesToFields(attrs []attribute.KeyValue) []etw.FieldOpt {
	s := attribute.NewSet(attrs...)
	i := s.Iter()
	fs := make([]etw.FieldOpt, 0, i.Len())
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

func (e *exporter) exportSpanEvents(ctx context.Context, span trace.ReadOnlySpan) error {
	if !e.exportEvents {
		return nil
	}
	name := span.Name()
	events := span.Events()
	sc := span.SpanContext()
	base := []etw.FieldOpt{
		etw.StringField("span.trace_id", sc.TraceID().String()),
		etw.StringField("span.id", sc.SpanID().String()),
	}
	opts := []etw.EventOpt{etw.WithLevel(etw.LevelInfo)}

	for _, evt := range events {
		if err := ctx.Err(); err != nil {
			return err
		}
		if e.IsClosed() {
			return nil
		}

		attrs := evt.Attributes
		fields := make([]etw.FieldOpt, 0, 2+len(base)+len(attrs)+1)
		fields = append(fields,
			etw.StringField("event.name", evt.Name),
			e.FormatTime("event.time", evt.Time),
		)
		fields = append(fields, base...)
		fields = append(fields, attributesToFields(attrs)...)
		if n := evt.DroppedAttributeCount; n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_attributes_count", n))
		}

		if err := e.Provider().WriteEvent(name, opts, fields); err != nil {
			return err
		}
	}
	return nil
}

func (e *exporter) Shutdown(ctx context.Context) error {
	return e.Close(ctx)
}
