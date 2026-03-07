package errors

import (
	"net/http"
	"testing"
)

func TestAppError(t *testing.T) {
	err := BadRequest("invalid input")

	if err.Code != "BAD_REQUEST" {
		t.Errorf("expected code BAD_REQUEST, got %s", err.Code)
	}

	if err.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, err.StatusCode)
	}

	if err.Message != "invalid input" {
		t.Errorf("expected message 'invalid input', got %s", err.Message)
	}
}

func TestWithDetails(t *testing.T) {
	err := NotFound("user not found").WithDetails("user_id: 123")

	if err.Details != "user_id: 123" {
		t.Errorf("expected details 'user_id: 123', got %s", err.Details)
	}

	expected := "NOT_FOUND: user not found (user_id: 123)"
	if err.Error() != expected {
		t.Errorf("expected error string '%s', got '%s'", expected, err.Error())
	}
}

func TestGetStatusCode(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected int
	}{
		{"AppError", BadRequest("test"), http.StatusBadRequest},
		{"NotFound", NotFound("test"), http.StatusNotFound},
		{"InternalError", InternalError("test"), http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code := GetStatusCode(tt.err)
			if code != tt.expected {
				t.Errorf("expected status %d, got %d", tt.expected, code)
			}
		})
	}
}
