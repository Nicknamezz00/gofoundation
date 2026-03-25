package metadata

import (
	"context"
	"fmt"

	"google.golang.org/grpc/metadata"
)

const (
	UserIDKey  = "x-user-id"
	TraceIDKey = "x-trace-id"
)

// UserIDFromContext extracts the user ID from gRPC metadata in the context.
// Returns an error if the user ID is missing or invalid.
func UserIDFromContext(ctx context.Context) (string, error) {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return "", fmt.Errorf("no metadata in context")
	}

	values := md.Get(UserIDKey)
	if len(values) == 0 {
		return "", fmt.Errorf("user ID not found in metadata")
	}
	return values[0], nil
}

// TraceIDFromContext extracts the trace ID from incoming gRPC metadata.
func TraceIDFromContext(ctx context.Context) string {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return ""
	}
	values := md.Get(TraceIDKey)
	if len(values) == 0 {
		return ""
	}
	return values[0]
}

// InjectUserID appends a user ID to the outgoing gRPC metadata.
// Used by clients (e.g., gateway) to propagate user context to backend services.
func InjectUserID(ctx context.Context, userID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, UserIDKey, userID)
}

// InjectTraceID appends a trace ID to the outgoing gRPC metadata.
func InjectTraceID(ctx context.Context, traceID string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, TraceIDKey, traceID)
}

// MustUserIDFromContext extracts the user ID from context and panics if not found.
// Use this only in handlers where user ID is guaranteed to be present (after auth interceptor).
func MustUserIDFromContext(ctx context.Context) string {
	userID, err := UserIDFromContext(ctx)
	if err != nil {
		panic(fmt.Sprintf("user ID not found in context: %v", err))
	}
	return userID
}
