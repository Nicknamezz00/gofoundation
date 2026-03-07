package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Nicknamezz00/gofoundation/gateway"
	"github.com/Nicknamezz00/gofoundation/logger"
	"github.com/Nicknamezz00/gofoundation/response"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// Initialize Jaeger exporter
	ctx := context.Background()

	// Create OTLP HTTP exporter
	exporter, err := otlptrace.New(ctx, otlptracehttp.NewClient(
		otlptracehttp.WithEndpoint("localhost:4318"),
		otlptracehttp.WithInsecure(),
	))
	if err != nil {
		log.Printf("Warning: failed to create exporter: %v", err)
		log.Println("Continuing without Jaeger integration...")
	}

	// Create resource
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName("full-featured-example"),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		log.Fatal(err)
	}

	// Create tracer provider
	var tp *sdktrace.TracerProvider
	if exporter != nil {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithBatcher(exporter),
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
	} else {
		tp = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
	}

	// Create gateway with full configuration
	config := gateway.Config{
		ServiceName:    "full-featured-example",
		ServiceVersion: "1.0.0",

		// Logger config
		Logger: logger.Config{
			Level:   logger.InfoLevel,
			DevMode: true, // Output to stdout for demo
		},

		// Tracer provider
		TracerProvider: tp,

		// CORS config
		EnableCORS:  true,
		CORSOrigins: []string{"*"},
		CORSMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		CORSHeaders: []string{"Content-Type", "Authorization"},

		// Request ID
		EnableRequestID: true,
	}

	gw, err := gateway.New(config)
	if err != nil {
		log.Fatal(err)
	}

	// Define handlers
	http.HandleFunc("/api/users", gw.HandlerFunc(handleUsers))
	http.HandleFunc("/api/orders", gw.HandlerFunc(handleOrders))

	// Start server
	log.Println("Server starting on :8080")
	log.Println("Jaeger UI: http://localhost:16686")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}

func handleUsers(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)

	// Simulate database query
	ctx := r.Context()
	tracer := otel.Tracer("full-featured-example")
	ctx, span := tracer.Start(ctx, "query-users")
	defer span.End()

	time.Sleep(50 * time.Millisecond)

	users := []map[string]interface{}{
		{"id": 1, "name": "Alice", "email": "alice@example.com"},
		{"id": 2, "name": "Bob", "email": "bob@example.com"},
	}

	rw.WriteJSON(http.StatusOK, users)
}

func handleOrders(w http.ResponseWriter, r *http.Request) {
	rw := w.(response.Writer)

	// Simulate multiple operations
	ctx := r.Context()
	tracer := otel.Tracer("full-featured-example")

	// Query orders
	ctx, span1 := tracer.Start(ctx, "query-orders")
	time.Sleep(30 * time.Millisecond)
	span1.End()

	// Process orders
	ctx, span2 := tracer.Start(ctx, "process-orders")
	time.Sleep(20 * time.Millisecond)
	span2.End()

	orders := []map[string]interface{}{
		{"id": 1, "user_id": 1, "total": 99.99},
		{"id": 2, "user_id": 2, "total": 149.99},
	}

	rw.WriteJSON(http.StatusOK, orders)
}
