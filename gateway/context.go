package gateway

import (
	"context"

	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/response"
	"github.com/Nicknamezz00/gofoundation/trace"
)

type contextKey int

const (
	loggerKey contextKey = iota
	responseWriterKey
)

// WithLogger adds logger to context
func WithLogger(ctx context.Context, log logger.Logger) context.Context {
	return context.WithValue(ctx, loggerKey, log)
}

// GetLogger retrieves logger from context
func GetLogger(ctx context.Context) logger.Logger {
	if log, ok := ctx.Value(loggerKey).(logger.Logger); ok {
		return log
	}
	return nil
}

// WithResponseWriter adds response writer to context
func WithResponseWriter(ctx context.Context, w response.Writer) context.Context {
	return context.WithValue(ctx, responseWriterKey, w)
}

// GetResponseWriter retrieves response writer from context
func GetResponseWriter(ctx context.Context) response.Writer {
	if w, ok := ctx.Value(responseWriterKey).(response.Writer); ok {
		return w
	}
	return nil
}

// GetTraceInfo is a convenience function to get trace info from context
func GetTraceInfo(ctx context.Context) *trace.Info {
	return trace.GetTraceInfo(ctx)
}
