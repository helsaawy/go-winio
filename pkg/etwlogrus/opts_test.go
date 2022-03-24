//go:build windows

package etwlogrus

import (
	"context"
	"testing"

	"github.com/Microsoft/go-winio/pkg/guid"
	octrace "go.opencensus.io/trace"
	ottracesdk "go.opentelemetry.io/otel/sdk/trace"
)

func TestSpanOC(t *testing.T) {
	type test struct {
		n  string
		en bool
	}

	tests := []test{
		{
			n:  "NoSampling",
			en: false,
		},
		{
			n:  "Sampling",
			en: true,
		},
	}

	f := func(t *testing.T, tt test, s octrace.Sampler) {
		nWant := t.Name()
		ctx, span := octrace.StartSpan(context.Background(), nWant, octrace.WithSampler(s))
		t.Cleanup(span.End)

		t.Run("Name", func(t *testing.T) {
			n, ok := nameFromOCSpan(ctx)
			if ok != tt.en {
				t.Fatalf("name extraction success %t, wanted %t", ok, tt.en)
			}
			if tt.en && n != nWant {
				t.Fatalf("name extracted got %s, wanted %s", n, nWant)
			}
		})

		t.Run("ActivityID", func(t *testing.T) {
			g, ok := idFromOCSpan(ctx)
			if !ok {
				t.Fatal("id extration failed")
			}
			gWant := guid.FromArray(span.SpanContext().TraceID)
			if g != gWant {
				t.Fatalf("id extracted got %v, wanted %v", g, gWant)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			s := octrace.NeverSample()
			if tt.en {
				s = octrace.AlwaysSample()
			}

			f(t, tt, s)

			// test nested spans
			t.Run("Nested", func(t *testing.T) {
				f(t, tt, s)
			})
		})

	}
}

func TestSpanOTName(t *testing.T) {
	type test struct {
		n  string
		en bool
	}

	ctx := context.Background()
	tests := []test{
		{
			n:  "NoSampling",
			en: false,
		},
		{
			n:  "Sampling",
			en: true,
		},
	}

	f := func(t *testing.T, tt test, tvp *ottracesdk.TracerProvider) {
		nWant := t.Name()
		ctx, span := tvp.Tracer("").Start(ctx, nWant)
		t.Cleanup(func() { span.End() })

		t.Run("Name", func(t *testing.T) {
			n, ok := nameFromOTSpan(ctx)
			if ok != tt.en {
				t.Fatalf("name extraction success %t, wanted %t", ok, tt.en)
			}
			if tt.en && n != nWant {
				t.Fatalf("name extracted got %s, wanted %s", n, nWant)
			}
		})

		t.Run("ActivityID", func(t *testing.T) {
			g, ok := idFromOTSpan(ctx)
			if !ok {
				t.Fatal("id extration failed")
			}
			gWant := guid.FromArray(span.SpanContext().TraceID())
			if g != gWant {
				t.Fatalf("id extracted got %v, wanted %v", g, gWant)
			}
		})
	}

	for _, tt := range tests {
		t.Run(tt.n, func(t *testing.T) {
			s := ottracesdk.NeverSample()
			if tt.en {
				s = ottracesdk.AlwaysSample()
			}

			tvp := ottracesdk.NewTracerProvider(ottracesdk.WithSampler(s))
			t.Cleanup(func() { tvp.Shutdown(ctx) })
			f(t, tt, tvp)

			// test nested spans
			t.Run("Nested", func(t *testing.T) {
				f(t, tt, tvp)
			})
		})

	}
}
