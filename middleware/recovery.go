package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/response"
)

// Recovery creates a middleware that recovers from panics
func Recovery(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Log the panic with stack trace
					log.WithContext(r.Context()).Error(
						"panic recovered",
						logger.String("panic", fmt.Sprintf("%v", err)),
						logger.String("stack", string(debug.Stack())),
					)

					// Try to write error response
					if rw, ok := w.(response.Writer); ok {
						rw.WriteError(http.StatusInternalServerError, fmt.Errorf("internal server error"))
					} else {
						http.Error(w, "Internal Server Error", http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		})
	}
}
