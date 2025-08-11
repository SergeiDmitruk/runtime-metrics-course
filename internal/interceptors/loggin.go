package interceptors

import (
	"context"
	"time"

	"github.com/runtime-metrics-course/internal/logger"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/status"
)

func GRPCLoggerInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (resp interface{}, err error) {
		start := time.Now()

		var clientIP string
		if p, ok := peer.FromContext(ctx); ok {
			clientIP = p.Addr.String()
		}

		resp, err = handler(ctx, req)

		st, _ := status.FromError(err)

		logger.Log.Sugar().Infow("gRPC call",
			"method", info.FullMethod,
			"client_ip", clientIP,
			"status_code", st.Code(),
			"status_message", st.Message(),
			"duration", time.Since(start),
		)

		return resp, err
	}
}
