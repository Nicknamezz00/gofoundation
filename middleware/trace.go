package middleware

import (
	"net/http"

	"github.com/Nicknamezz00/gofoundation/trace"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	oteltrace "go.opentelemetry.io/otel/trace"
)

// Trace creates a middleware that extracts or creates trace context
func Trace(serviceName string) func(http.Handler) http.Handler {
	tracer := otel.Tracer(serviceName)

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Extract trace context from headers
			ctx := otel.GetTextMapPropagator().Extract(r.Context(), &headerCarrier{r.Header})

			// Start a new span
			ctx, span := tracer.Start(ctx, r.Method+" "+r.URL.Path,
				oteltrace.WithSpanKind(oteltrace.SpanKindServer),
			)
			defer span.End()

			// Extract trace info
			traceInfo := trace.ExtractTraceInfo(span.SpanContext())

			// Add trace info to context
			ctx = trace.WithTraceInfo(ctx, traceInfo)
			ctx = trace.WithSpan(ctx, span)

			// Set span attributes
			span.SetAttributes(
				attribute.String("http.method", r.Method),
				attribute.String("http.url", r.URL.String()),
				attribute.String("http.host", r.Host),
			)

			// Continue with updated context
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// headerCarrier implements TextMapCarrier for HTTP headers
type headerCarrier struct {
	header http.Header
}

func (c *headerCarrier) Get(key string) string {
	return c.header.Get(key)
}

func (c *headerCarrier) Set(key, value string) {
	c.header.Set(key, value)
}

func (c *headerCarrier) Keys() []string {
	keys := make([]string, 0, len(c.header))
	for k := range c.header {
		keys = append(keys, k)
	}
	return keys
}
