//go:build windows

package oc

import (
	"errors"
	"sort"
	"time"

	"go.opencensus.io/trace"

	"github.com/Microsoft/go-winio/pkg/etw"
)

// ErrNoProvider is returned when the exporter is created without a provider being configured.
var ErrNoProvider = errors.New("no ETW registered provider")

type ExporterOpt func(*exporter) error

type exporter struct {
	provider      *etw.Provider
	closeProvider bool
	// returns additional options to add to the event
	getEventsOpts func(*trace.SpanData) []etw.EventOpt
	formatTime    func(string, time.Time) etw.FieldOpt
}

var _ trace.Exporter = &exporter{}

// NewExporter returns a [trace.Exporter] that exports OpenCensus spans to ETW
// based on the the following rules:
//  * ETW entries will contain the Attributes, SpanKind, TraceID,
//   SpanID, and ParentSpanID.
//  * Annotation, MessageEvents, and Links will not be exported.
//  * The span itself will be written at [etw.LevelInfo], unless
//   `s.Status.Code != 0`, in which case it will be written at [etw.LevelError],
//   with the field `Error` set to Status.Message
func NewExporter(opts ...ExporterOpt) (trace.Exporter, error) {
	e := &exporter{}
	opts = append([]ExporterOpt{WithTimeFormat(time.RFC3339Nano)}, opts...)
	for _, o := range opts {
		if err := o(e); err != nil {
			return nil, err
		}
	}
	if e.provider == nil {
		return nil, ErrNoProvider
	}
	return e, nil
}

func (e *exporter) ExportSpan(span *trace.SpanData) {
	level := etw.LevelInfo
	hasError := span.Code != 0
	if hasError {
		level = etw.LevelError
	}
	if !e.provider.IsEnabledForLevel(level) {
		return
	}
	name := span.Name
	// extra room for two more options in addition to log level to avoid reallocating
	// if the user also provides options
	opts := make([]etw.EventOpt, 0, 3)
	opts = append(opts, etw.WithLevel(level))
	if e.getEventsOpts != nil {
		opts = append(opts, e.getEventsOpts(span)...)
	}

	// Reserve extra space for the span properties .
	fields := make([]etw.FieldOpt, 0, len(span.Attributes)+10)
	fields = append(fields,
		etw.StringField("TraceID", span.TraceID.String()),
		etw.StringField("SpanID", span.SpanID.String()),
		etw.StringField("ParentSpanID", span.ParentSpanID.String()),
		e.formatTime("StartTime", span.StartTime),
		e.formatTime("EndTime", span.EndTime),
		etw.StringField("Duration", span.EndTime.Sub(span.StartTime).String()),
		etw.Int32Field("Code", span.Code), // convert to gRPC status code string?
		etw.StringField("Error", span.Message),
		etw.StringField("SpanKind", spanKindToString(span.SpanKind)),
	)

	// Sort the fields by name so they are consistent in each instance
	// of an event. Otherwise, the fields don't line up in WPA.
	data := span.Attributes
	names := make([]string, 0, len(data))
	for k := range data {
		names = append(names, k)
	}
	sort.Strings(names)

	for _, k := range names {
		fields = append(fields, etw.SmartField(k, data[k]))
	}

	if span.DroppedAnnotationCount > 0 {
		fields = append(fields, etw.IntField("DroppedAnnotations", span.DroppedAnnotationCount))
	}

	// Firing an ETW event is essentially best effort, as the event write can
	// fail for reasons completely out of the control of the event writer (such
	// as a session listening for the event having no available space in its
	// buffers). Therefore, we don't return the error from WriteEvent, as it is
	// just noise in many cases.
	e.provider.WriteEvent(name, opts, fields) //nolint:errcheck
}
func spanKindToString(sk int) string {
	switch sk {
	case trace.SpanKindClient:
		return "client"
	case trace.SpanKindServer:
		return "server"
	default:
		return ""
	}
}
