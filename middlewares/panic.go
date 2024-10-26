package middlewares

import (
	"context"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"runtime/debug"
)

func panicHandler(ctx context.Context, p any) error {
	log.Errorf("panic recoverd: [%+v], debug stack: %s", p, debug.Stack())
	return status.Errorf(codes.Internal, "panic recovered: %v", p)
}

func PanicRecoveryInspector() grpc.UnaryServerInterceptor {
	return grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandlerContext(panicHandler))
}
