package interceptors

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// TracingInterceptor creates OpenTelemetry spans for gRPC requests.
func TracingInterceptor(serviceName string) grpc.UnaryServerInterceptor {
	tracer := otel.Tracer(serviceName)
	propagator := getPropagator()

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Extract remote context (traceparent, baggage, etc.) from gRPC metadata
		// so server spans become children of client spans.
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			ctx = propagator.Extract(ctx, metadataCarrier(md))
		}

		// Start a new span
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Add user ID as span attribute if present
		if userID, err := UserIDFromContext(ctx); err == nil {
			span.SetAttributes(attribute.String("user.id", userID))
		}

		// Call the handler
		resp, err := handler(ctx, req)

		// Record error if present
		if err != nil {
			st, _ := status.FromError(err)
			span.SetStatus(codes.Error, st.Message())
			span.SetAttributes(
				attribute.String("grpc.status_code", st.Code().String()),
				attribute.String("error.message", st.Message()),
			)
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

type metadataCarrier metadata.MD

func (c metadataCarrier) Get(key string) string {
	values := metadata.MD(c).Get(key)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

func (c metadataCarrier) Set(key, value string) {
	metadata.MD(c).Set(key, value)
}

func (c metadataCarrier) Keys() []string {
	md := metadata.MD(c)
	out := make([]string, 0, len(md))
	for k := range md {
		out = append(out, k)
	}
	return out
}

func getPropagator() propagation.TextMapPropagator {
	if p := otel.GetTextMapPropagator(); p != nil {
		return p
	}
	return propagation.NewCompositeTextMapPropagator()
}
