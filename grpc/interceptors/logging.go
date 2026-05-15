package interceptors

import (
	"context"
	"time"

	"github.com/Nicknamezz00/gofoundation/logger"
	grpcmeta "github.com/Nicknamezz00/gofoundation/grpc/metadata"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

// LoggingInterceptor logs gRPC requests and responses with timing information.
func LoggingInterceptor(l logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		start := time.Now()
		userID, _ := UserIDFromContext(ctx)
		traceID := grpcmeta.TraceIDFromContext(ctx)

		resp, err := handler(ctx, req)
		duration := time.Since(start)

		var fields []logger.Field
		if traceID != "" {
			fields = append(fields, logger.String("trace_id", traceID))
		}
		fields = append(fields,
			logger.String("method", info.FullMethod),
			logger.String("user_id", userID),
			logger.Int64("duration_ms", duration.Milliseconds()),
		)

		if err != nil {
			st, _ := status.FromError(err)
			fields = append(fields,
				logger.String("error", st.Message()),
				logger.String("code", st.Code().String()),
				logger.Any("detail", st.Details()),
			)
			l.WithContext(ctx).Error("gRPC request completed with error", fields...)
		} else {
			l.WithContext(ctx).Info("gRPC request completed", fields...)
		}

		return resp, err
	}
}
