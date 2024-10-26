package middlewares

import (
	"context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func CloseConnectionInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		err := grpc.SetHeader(ctx, metadata.Pairs("Connection", "close"))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to close connection: %v", err)
		}
		return handler(ctx, req)
	}
}
