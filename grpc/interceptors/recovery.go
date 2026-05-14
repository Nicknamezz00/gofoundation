package interceptors

import (
	"context"
	"runtime/debug"

	"github.com/Nicknamezz00/gofoundation/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// logPanic logs a recovered panic with stack trace.
func logPanic(log logger.Logger, msg, method string, r any) {
	stack := debug.Stack()
	log.Error(msg,
		logger.String("method", method),
		logger.Any("panic", r),
		logger.String("stack", string(stack)),
	)
}

// RecoveryInterceptor recovers from panics in gRPC handlers and returns an error.
func RecoveryInterceptor(log logger.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		defer func() {
			if r := recover(); r != nil {
				logPanic(log, "panic recovered in gRPC handler", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(ctx, req)
	}
}

// StreamRecoveryInterceptor recovers from panics in streaming gRPC handlers.
func StreamRecoveryInterceptor(log logger.Logger) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) (err error) {
		defer func() {
			if r := recover(); r != nil {
				logPanic(log, "panic recovered in streaming gRPC handler", info.FullMethod, r)
				err = status.Error(codes.Internal, "internal server error")
			}
		}()
		return handler(srv, ss)
	}
}

