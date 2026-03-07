package response

import (
	"net/http"

	"github.com/Nicknamezz00/gofoundation/trace"
)

// Writer wraps http.ResponseWriter with additional methods
type Writer interface {
	http.ResponseWriter
	WriteJSON(statusCode int, data interface{}) error
	WriteError(statusCode int, err error) error
	TraceInfo() *trace.Info
	Status() int
}

// Envelope is the standard response format
type Envelope struct {
	Data    interface{} `json:"data,omitempty"`
	TraceID string      `json:"trace_id,omitempty"`
	Error   *ErrorInfo  `json:"error,omitempty"`
}

// ErrorInfo contains error details
type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}
