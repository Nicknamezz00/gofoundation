package middleware

import (
	"net/http"
)

const TraceIDHeader = "X-Trace-Id"

// ExtraResponseHeader echoes the current trace id (from context) as an HTTP response header.
// This is meant to be used after the Trace middleware has populated context.
func ExtraResponseHeader() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ti := GetTraceInfo(r.Context()); ti != nil && ti.TraceID != "" {
				// Must be set before the handler writes the response.
				w.Header().Set(TraceIDHeader, ti.TraceID)
			}
			next.ServeHTTP(w, r)
		})
	}
}
