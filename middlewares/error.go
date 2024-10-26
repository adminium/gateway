package middlewares

import (
	"context"
	"github.com/gozelle/logger"
	"google.golang.org/grpc"
)

var log = logger.NewLogger("grpc-middlewares")

func GrpcErrorInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		if err != nil {
			log.Errorf("[grpc error interceptor]: %+v", err)
		}
		return resp, err
	}
}
