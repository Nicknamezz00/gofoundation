package gateway

import (
	"net/http"
	"time"

	"github.com/Nicknamezz00/gofoundation/logger"
	"go.opentelemetry.io/otel/trace"
)

// Gateway is the main HTTP gateway
type Gateway interface {
	Use(middleware ...Middleware)
	Handler(h http.Handler) http.Handler
	HandlerFunc(h http.HandlerFunc) http.HandlerFunc
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

// Middleware is a function that wraps an http.Handler
type Middleware func(http.Handler) http.Handler

// Config holds gateway configuration
type Config struct {
	ServiceName    string
	ServiceVersion string

	// Logger configuration
	Logger logger.Config

	// Tracer provider (optional, will create default if nil)
	TracerProvider trace.TracerProvider

	// Middleware configuration
	EnableCORS      bool
	CORSOrigins     []string
	CORSMethods     []string
	CORSHeaders     []string
	EnableRequestID bool

	// Custom middleware
	Middlewares []Middleware

	// Timeouts
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// DefaultConfig returns default gateway configuration
func DefaultConfig(serviceName string) Config {
	return Config{
		ServiceName:     serviceName,
		ServiceVersion:  "1.0.0",
		Logger:          logger.DefaultConfig(),
		EnableCORS:      true,
		CORSOrigins:     []string{"*"},
		CORSMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSHeaders:     []string{"Content-Type", "Authorization"},
		EnableRequestID: true,
		ReadTimeout:     15 * time.Second,
		WriteTimeout:    15 * time.Second,
	}
}
