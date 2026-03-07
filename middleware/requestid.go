package middleware

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

const RequestIDHeader = "X-Request-Id"

// RequestID creates a middleware that generates or extracts request IDs
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get(RequestIDHeader)
			if requestID == "" {
				requestID = uuid.New().String()
			}

			// Set request ID in response header
			w.Header().Set(RequestIDHeader, requestID)

			// Add to context
			ctx := r.Context()
			ctx = contextWithRequestID(ctx, requestID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

type requestIDKey struct{}

func contextWithRequestID(ctx context.Context, requestID string) context.Context {
	return context.WithValue(ctx, requestIDKey{}, requestID)
}

func GetRequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey{}).(string); ok {
		return requestID
	}
	return ""
}
