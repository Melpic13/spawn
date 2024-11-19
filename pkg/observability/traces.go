package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

// Tracer wraps an OpenTelemetry tracer.
type Tracer struct {
	trace.Tracer
}

// NewTracer creates a new tracer.
func NewTracer(name string) *Tracer {
	if name == "" {
		name = "spawn"
	}
	return &Tracer{Tracer: otel.Tracer(name)}
}

// StartSpan starts a trace span.
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, trace.Span) {
	return t.Start(ctx, name)
}

// TraceID returns string trace id if present.
func TraceID(ctx context.Context) string {
	span := trace.SpanFromContext(ctx)
	sc := span.SpanContext()
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// EnsureTracing validates tracer setup.
func EnsureTracing(t *Tracer) error {
	if t == nil || t.Tracer == nil {
		return fmt.Errorf("ensure tracing: tracer is not initialized")
	}
	return nil
}
