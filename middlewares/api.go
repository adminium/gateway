package middlewares

import (
	"context"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"time"
)

func APIInspector() grpc.UnaryServerInterceptor {
	//l := logger.NewLogger("api")
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		traceID := uuid.New().String()
		startTime := time.Now().Local()
		l := log.With("trace_id", traceID)

		err = grpc.SetHeader(ctx, metadata.Pairs("trace-id", traceID))
		if err != nil {
			l.Errorf("Error setting trace-id header: %v", err)
			return
		}
		l.Infof("API:[%s], Request:[%s]", info.FullMethod, req)
		resp, err = handler(ctx, req)
		l.Infof("API:[%s], Response:[%s]", info.FullMethod, resp)
		l.Infof("API:[%s], Time Taken:[%s]", info.FullMethod, time.Since(startTime))
		if err != nil {
			l.Errorf("API:[%s], Error:[%s]", info.FullMethod, err)
		}
		return resp, err
	}
}
