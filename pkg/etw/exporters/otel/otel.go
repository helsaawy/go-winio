//go:build windows

package otel

import (
	"context"
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
	// semconv "go.opentelemetry.io/otel/semconv/v1.4.0"

	"github.com/Microsoft/go-winio/pkg/etw"
	common "github.com/Microsoft/go-winio/pkg/etw/internal/exporters"
	"github.com/Microsoft/go-winio/pkg/etw/internal/exporters/fields"
)

type input = trace.ReadOnlySpan

type Opt func(*exporter) error

type exporter struct {
	c            common.Common[trace.ReadOnlySpan]
	exportEvents bool
}

var _ trace.SpanExporter = &exporter{}

// New returns a [trace.SpanExporter] that exports Open Telemetry spans to ETW
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
func New(opts ...Opt) (trace.SpanExporter, error) {
	e := &exporter{}
	opts = append([]Opt{WithTimeFormat(time.RFC3339Nano)}, opts...)
	for _, o := range opts {
		if err := o(e); err != nil {
			return nil, err
		}
	}
	if err := e.c.Validate(); err != nil {
		return nil, err
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
		if e.c.IsClosed() {
			return nil
		}

		level := etw.LevelInfo
		st := span.Status()
		if st.Code == codes.Error {
			level = etw.LevelError
		}
		if !e.c.Provider.IsEnabledForLevel(level) {
			continue
		}

		sc := span.SpanContext()
		psc := span.Parent()
		rsc := span.Resource()
		il := span.InstrumentationLibrary()
		attrs := span.Attributes()

		// extra room for two more options in addition to log level to avoid reallocating
		// if the user also provides options
		opts := make([]etw.EventOpt, 0, 5)
		opts = append(opts, etw.WithLevel(level))
		if e.c.EnableActivityID {
			opts = append(opts, etw.WithActivityID(common.SpanIDToGUID(sc.SpanID())))
		}
		if e.c.EnableRelatedActivityID {
			opts = append(opts, etw.WithRelatedActivityID(common.SpanIDToGUID(psc.SpanID())))
		}
		opts = append(opts, e.c.GetEventsOpts(span)...)

		// Reserve extra space for the span properties .
		efs := make([]etw.FieldOpt, 0, 12+rsc.Len()+2+len(attrs)+3)
		efs = append(efs,
			etw.StringField(fields.PayloadName, span.Name()),
			etw.StringField(fields.SpanParentID, psc.SpanID().String()),
			etw.StringField(fields.SpanID, sc.SpanID().String()),
			etw.StringField(fields.TraceID, sc.TraceID().String()),
			e.c.FormatTime(fields.StartTime, span.StartTime()),
			etw.Int64Field(fields.Time, span.StartTime().UnixMilli()),
			e.c.FormatTime(fields.EndTime, span.EndTime()),
			etw.StringField(fields.Duration, span.EndTime().Sub(span.StartTime()).String()),
		)
		if sk := span.SpanKind(); sk != oteltrace.SpanKindUnspecified && sk != oteltrace.SpanKindInternal {
			efs = append(efs, etw.StringField(fields.SpanKind, sk.String()))
		}

		if st.Code != codes.Unset {
			efs = append(efs,
				etw.BoolField(fields.Success, st.Code == codes.Ok),
				etw.Uint32Field(fields.StatusCode, uint32(st.Code)),
			)
			if st.Description != "" {
				efs = append(efs, etw.StringField(fields.StatusMessage, st.Description))
			}
		}

		efs = append(efs, resourcesToFields(rsc)...)
		efs = append(efs, libToFields(il)...)

		efs = append(efs, attributesToFields(attrs)...)

		if n := span.DroppedAttributes(); n > 0 {
			efs = append(efs, etw.IntField("otel.dropped_attributes_count", n))
		}
		if n := span.DroppedEvents(); n > 0 {
			efs = append(efs, etw.IntField("otel.dropped_events_count", n))
		}
		if n := span.DroppedLinks(); n > 0 {
			efs = append(efs, etw.IntField("otel.dropped_links_count", n))
		}

		if err := e.c.Provider.WriteEvent(fields.SpanEventName, opts, efs); err != nil {
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
		if e.c.IsClosed() {
			return nil
		}

		attrs := evt.Attributes
		fields := make([]etw.FieldOpt, 0, 2+len(base)+len(attrs)+1)
		fields = append(fields,
			etw.StringField("event.name", evt.Name),
			e.c.FormatTime("event.time", evt.Time),
		)
		fields = append(fields, base...)
		fields = append(fields, attributesToFields(attrs)...)
		if n := evt.DroppedAttributeCount; n > 0 {
			fields = append(fields, etw.IntField("otel.dropped_attributes_count", n))
		}

		if err := e.c.Provider.WriteEvent(name, opts, fields); err != nil {
			return err
		}
	}
	return nil
}

func (e *exporter) Shutdown(ctx context.Context) error {
	return e.c.Close(ctx)
}
