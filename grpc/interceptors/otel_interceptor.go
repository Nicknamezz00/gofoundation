package interceptors

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

// OtelUnaryClientInterceptor injects the current OpenTelemetry context into
// outgoing gRPC metadata, allowing trace propagation across services.
func OtelUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	propagator := getPropagator()

	return func(
		ctx context.Context,
		method string,
		req, reply any,
		cc *grpc.ClientConn,
		invoker grpc.UnaryInvoker,
		opts ...grpc.CallOption,
	) error {
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			md = md.Copy()
		} else {
			md = metadata.New(nil)
		}

		propagator.Inject(ctx, metadataCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, md)

		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// OtelStreamClientInterceptor injects the current OpenTelemetry context into
// outgoing gRPC metadata for streaming calls.
func OtelStreamClientInterceptor() grpc.StreamClientInterceptor {
	propagator := getPropagator()

	return func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		cc *grpc.ClientConn,
		method string,
		streamer grpc.Streamer,
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		md, ok := metadata.FromOutgoingContext(ctx)
		if ok {
			md = md.Copy()
		} else {
			md = metadata.New(nil)
		}

		propagator.Inject(ctx, metadataCarrier(md))
		ctx = metadata.NewOutgoingContext(ctx, md)

		return streamer(ctx, desc, cc, method, opts...)
	}
}
