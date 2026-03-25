package middleware

import (
	"net/http"

	"github.com/Nicknamezz00/gofoundation/trace"
	grpcmeta "github.com/Nicknamezz00/rago-api/shared/grpc/metadata"
)

// GRPCTrace injects the current request's trace ID into the outgoing gRPC metadata
// on the context. Place this after the Trace middleware so trace info is already set.
func GRPCTrace() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if ti := trace.GetTraceInfo(r.Context()); ti != nil && ti.TraceID != "" {
				r = r.WithContext(grpcmeta.InjectTraceID(r.Context(), ti.TraceID))
			}
			next.ServeHTTP(w, r)
		})
	}
}
