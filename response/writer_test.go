package response

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Nicknamezz00/gofoundation/trace"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	traceInfo := &trace.Info{
		TraceID:   "test-trace-id",
		SpanID:    "test-span-id",
		Timestamp: time.Now(),
	}

	rw := NewWriter(w, traceInfo)

	data := map[string]string{"message": "hello"}
	err := rw.WriteJSON(http.StatusOK, data)
	if err != nil {
		t.Fatalf("WriteJSON failed: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	if tid := w.Header().Get("X-Trace-Id"); tid != "test-trace-id" {
		t.Errorf("expected X-Trace-Id test-trace-id, got %s", tid)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	traceInfo := &trace.Info{
		TraceID:   "test-trace-id",
		SpanID:    "test-span-id",
		Timestamp: time.Now(),
	}

	rw := NewWriter(w, traceInfo)

	err := rw.WriteError(http.StatusNotFound, http.ErrNotSupported)
	if err != nil {
		t.Fatalf("WriteError failed: %v", err)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestStatus(t *testing.T) {
	w := httptest.NewRecorder()
	rw := NewWriter(w, nil)

	rw.WriteHeader(http.StatusCreated)

	if rw.Status() != http.StatusCreated {
		t.Errorf("expected status %d, got %d", http.StatusCreated, rw.Status())
	}
}
