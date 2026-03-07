package response

import (
	"bufio"
	"net"
	"net/http"

	jsoniter "github.com/json-iterator/go"
	"github.com/Nicknamezz00/gofoundation/trace"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

type writer struct {
	http.ResponseWriter
	traceInfo  *trace.Info
	statusCode int
	written    bool
}

// NewWriter creates a new response writer
func NewWriter(w http.ResponseWriter, traceInfo *trace.Info) Writer {
	return &writer{
		ResponseWriter: w,
		traceInfo:      traceInfo,
		statusCode:     http.StatusOK,
		written:        false,
	}
}

func (w *writer) WriteHeader(statusCode int) {
	if w.written {
		return
	}
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
	w.written = true
}

func (w *writer) Write(data []byte) (int, error) {
	if !w.written {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(data)
}

func (w *writer) WriteJSON(statusCode int, data interface{}) error {
	var traceID string
	if w.traceInfo != nil {
		traceID = w.traceInfo.TraceID
	}

	envelope := Envelope{
		Data:    data,
		TraceID: traceID,
	}

	w.Header().Set("Content-Type", "application/json")
	if traceID != "" {
		w.Header().Set("X-Trace-Id", traceID)
	}

	w.WriteHeader(statusCode)

	return json.NewEncoder(w.ResponseWriter).Encode(envelope)
}

func (w *writer) WriteError(statusCode int, err error) error {
	errorInfo := &ErrorInfo{
		Code:    http.StatusText(statusCode),
		Message: err.Error(),
	}

	var traceID string
	if w.traceInfo != nil {
		traceID = w.traceInfo.TraceID
	}

	envelope := Envelope{
		Error:   errorInfo,
		TraceID: traceID,
	}

	w.Header().Set("Content-Type", "application/json")
	if traceID != "" {
		w.Header().Set("X-Trace-Id", traceID)
	}

	w.WriteHeader(statusCode)

	return json.NewEncoder(w.ResponseWriter).Encode(envelope)
}

func (w *writer) TraceInfo() *trace.Info {
	return w.traceInfo
}

func (w *writer) Status() int {
	return w.statusCode
}

// Hijack implements http.Hijacker
func (w *writer) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if hj, ok := w.ResponseWriter.(http.Hijacker); ok {
		return hj.Hijack()
	}
	return nil, nil, http.ErrNotSupported
}

// Flush implements http.Flusher
func (w *writer) Flush() {
	if f, ok := w.ResponseWriter.(http.Flusher); ok {
		f.Flush()
	}
}
