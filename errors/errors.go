package errors

import (
	"fmt"
	"net/http"
)

// AppError represents an application error
type AppError struct {
	Code       string
	Message    string
	Details    string
	StatusCode int
}

func (e *AppError) Error() string {
	if e.Details != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.Details)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// New creates a new AppError
func New(statusCode int, code, message string) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
	}
}

// WithDetails adds details to an error
func (e *AppError) WithDetails(details string) *AppError {
	e.Details = details
	return e
}

// Common error constructors
func BadRequest(message string) *AppError {
	return New(http.StatusBadRequest, "BAD_REQUEST", message)
}

func Unauthorized(message string) *AppError {
	return New(http.StatusUnauthorized, "UNAUTHORIZED", message)
}

func Forbidden(message string) *AppError {
	return New(http.StatusForbidden, "FORBIDDEN", message)
}

func NotFound(message string) *AppError {
	return New(http.StatusNotFound, "NOT_FOUND", message)
}

func Conflict(message string) *AppError {
	return New(http.StatusConflict, "CONFLICT", message)
}

func InternalError(message string) *AppError {
	return New(http.StatusInternalServerError, "INTERNAL_ERROR", message)
}

func ServiceUnavailable(message string) *AppError {
	return New(http.StatusServiceUnavailable, "SERVICE_UNAVAILABLE", message)
}

// GetStatusCode returns the HTTP status code for an error
func GetStatusCode(err error) int {
	if appErr, ok := err.(*AppError); ok {
		return appErr.StatusCode
	}
	return http.StatusInternalServerError
}
