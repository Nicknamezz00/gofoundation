package gateway

import (
	"fmt"
	"net/http"

	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/middleware"
	"github.com/Nicknamezz00/gofoundation/response"
	"github.com/Nicknamezz00/gofoundation/trace"
	oteltrace "go.opentelemetry.io/otel/trace"
)

type gateway struct {
	config      Config
	middlewares []Middleware
	tracer      oteltrace.Tracer
	logger      logger.Logger
}

// New creates a new Gateway
func New(config Config) (Gateway, error) {
	// Validate config
	if config.ServiceName == "" {
		return nil, fmt.Errorf("service name is required")
	}

	// Initialize tracer
	tp, err := trace.InitTracer(trace.Config{
		ServiceName:    config.ServiceName,
		ServiceVersion: config.ServiceVersion,
		TracerProvider: config.TracerProvider,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tracer: %w", err)
	}

	// Initialize logger
	log, err := logger.New(config.Logger)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize logger: %w", err)
	}

	g := &gateway{
		config:      config,
		middlewares: make([]Middleware, 0),
		tracer:      tp.Tracer(config.ServiceName),
		logger:      log,
	}

	// Add built-in middleware
	g.Use(middleware.Trace(config.ServiceName))

	if config.EnableRequestID {
		g.Use(middleware.RequestID())
	}

	if config.EnableCORS {
		g.Use(middleware.CORS(middleware.CORSConfig{
			Origins: config.CORSOrigins,
			Methods: config.CORSMethods,
			Headers: config.CORSHeaders,
		}))
	}

	g.Use(middleware.Recovery(log))
	g.Use(middleware.Logger(log))

	// Add custom middleware
	if len(config.Middlewares) > 0 {
		g.Use(config.Middlewares...)
	}

	return g, nil
}

func (g *gateway) Use(middleware ...Middleware) {
	g.middlewares = append(g.middlewares, middleware...)
}

func (g *gateway) Handler(h http.Handler) http.Handler {
	// Wrap handler with response writer
	var wrapped http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get trace info from context
		traceInfo := trace.GetTraceInfo(r.Context())

		// Wrap response writer
		rw := response.NewWriter(w, traceInfo)

		// Call handler
		h.ServeHTTP(rw, r)
	})

	// Apply middleware in reverse order
	for i := len(g.middlewares) - 1; i >= 0; i-- {
		wrapped = g.middlewares[i](wrapped)
	}

	return wrapped
}

func (g *gateway) HandlerFunc(h http.HandlerFunc) http.HandlerFunc {
	return g.Handler(h).ServeHTTP
}

func (g *gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This allows Gateway to be used as an http.Handler
	g.Handler(http.DefaultServeMux).ServeHTTP(w, r)
}
