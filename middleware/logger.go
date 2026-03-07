package middleware

import (
	"net/http"
	"time"

	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/response"
)

// Logger creates a middleware that logs requests
func Logger(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			// Wrap response writer to capture status
			var rw response.Writer
			if existingRW, ok := w.(response.Writer); ok {
				rw = existingRW
			} else {
				rw = response.NewWriter(w, nil)
			}

			// Process request
			next.ServeHTTP(rw, r)

			// Log request
			duration := time.Since(start)
			log.WithContext(r.Context()).Info(
				"request completed",
				logger.String("method", r.Method),
				logger.String("path", r.URL.Path),
				logger.Int("status", rw.Status()),
				logger.Int64("duration_ms", duration.Milliseconds()),
				logger.String("remote_addr", r.RemoteAddr),
			)
		})
	}
}
