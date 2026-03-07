package trace

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
)

// Info contains trace information for a request
type Info struct {
	TraceID   string    `json:"trace_id"`
	SpanID    string    `json:"span_id"`
	Timestamp time.Time `json:"timestamp"`
}

// contextKey is a private type for context keys
type contextKey int

const (
	traceInfoKey contextKey = iota
	spanKey
)

// WithTraceInfo adds trace info to context
func WithTraceInfo(ctx context.Context, info *Info) context.Context {
	return context.WithValue(ctx, traceInfoKey, info)
}

// GetTraceInfo retrieves trace info from context
func GetTraceInfo(ctx context.Context) *Info {
	if info, ok := ctx.Value(traceInfoKey).(*Info); ok {
		return info
	}
	return nil
}

// WithSpan adds a span to context
func WithSpan(ctx context.Context, span trace.Span) context.Context {
	return context.WithValue(ctx, spanKey, span)
}

// GetSpan retrieves span from context
func GetSpan(ctx context.Context) trace.Span {
	if span, ok := ctx.Value(spanKey).(trace.Span); ok {
		return span
	}
	return nil
}

// ExtractTraceInfo extracts trace info from span context
func ExtractTraceInfo(spanCtx trace.SpanContext) *Info {
	if !spanCtx.IsValid() {
		return &Info{
			Timestamp: time.Now(),
		}
	}

	return &Info{
		TraceID:   spanCtx.TraceID().String(),
		SpanID:    spanCtx.SpanID().String(),
		Timestamp: time.Now(),
	}
}
