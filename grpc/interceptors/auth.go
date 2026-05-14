package interceptors

import (
	"context"
	"fmt"

	grpcmeta "github.com/Nicknamezz00/gofoundation/grpc/metadata"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// AuthInterceptor extracts the user ID from gRPC metadata and validates it.
// This interceptor should be used by all backend services to ensure user context is present.
// Public methods (like Register, Login) should use AuthInterceptorWithSkip to bypass auth.
func AuthInterceptor() grpc.UnaryServerInterceptor {
	return AuthInterceptorWithSkip()
}

// AuthInterceptorWithSkip creates an auth interceptor that skips authentication for specified methods.
// skipMethods should be full method names like "/identity.v1.IdentityService/Register"
func AuthInterceptorWithSkip(skipMethods ...string) grpc.UnaryServerInterceptor {
	skipMap := make(map[string]bool)
	for _, method := range skipMethods {
		skipMap[method] = true
	}

	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Skip auth for public methods
		if skipMap[info.FullMethod] {
			return handler(ctx, req)
		}

		// Extract metadata from context
		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		// Get user ID from metadata
		values := md.Get(grpcmeta.UserIDKey)
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "user ID not found in metadata")
		}

		// Validate user ID format
		userID := values[0]

		// Store user ID in context for handlers to access
		ctx = context.WithValue(ctx, userIDContextKey{}, userID)

		// Call the handler with the enriched context
		return handler(ctx, req)
	}
}

// userIDContextKey is a private type for context keys to avoid collisions
type userIDContextKey struct{}

// UserIDFromContext extracts the user ID from the context.
// This should be called by service handlers after the AuthInterceptor has run.
func UserIDFromContext(ctx context.Context) (string, error) {
	userID, ok := ctx.Value(userIDContextKey{}).(string)
	if !ok {
		return "", fmt.Errorf("user ID not found in context")
	}
	return userID, nil
}

// AuthStreamInterceptor is the streaming equivalent of AuthInterceptorWithSkip.
// It extracts and validates the user ID from incoming gRPC metadata for streaming RPCs.
func AuthStreamInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := ss.Context()

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get(grpcmeta.UserIDKey)
		if len(values) == 0 {
			return status.Error(codes.Unauthenticated, "user ID not found in metadata")
		}

		// Inject user ID into stream context.
		enriched := context.WithValue(ctx, userIDContextKey{}, values[0])
		return handler(srv, &wrappedServerStream{ServerStream: ss, ctx: enriched})
	}
}

// wrappedServerStream replaces the context on a grpc.ServerStream.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context { return w.ctx }

// MustUserIDFromContext extracts the user ID and panics if not found.
// Use this in handlers where user ID is guaranteed to be present.
func MustUserIDFromContext(ctx context.Context) string {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("user ID not found in context: %v", err))
	}
	return userID
}

// ForwardUserIDInterceptor is a client-side unary interceptor that reads the user ID
// stored in the context by the server-side AuthInterceptor and forwards it as outgoing
// gRPC metadata, so downstream services receive it without manual WithUserContext calls.
func ForwardUserIDInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		if userID, err := UserIDFromContext(ctx); err == nil && userID != "" {
			ctx = grpcmeta.InjectUserID(ctx, userID)
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ForwardUserIDStreamInterceptor is the streaming equivalent of ForwardUserIDInterceptor.
func ForwardUserIDStreamInterceptor() grpc.StreamClientInterceptor {
	return func(ctx context.Context, desc *grpc.StreamDesc, cc *grpc.ClientConn, method string, streamer grpc.Streamer, opts ...grpc.CallOption) (grpc.ClientStream, error) {
		if userID, err := UserIDFromContext(ctx); err == nil && userID != "" {
			ctx = grpcmeta.InjectUserID(ctx, userID)
		}
		return streamer(ctx, desc, cc, method, opts...)
	}
}
