package middlewares

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"net/http"
	"strings"

	"google.golang.org/grpc"
)

var AllowedMethods = []string{
	"GET",
	"POST",
	"PUT",
	"DELETE",
	"OPTIONS",
	"UPDATE",
	"PATCH",
}

var AllowedHeaders = []string{
	"Origin",
	"Authorization",
	"Content-Type",
	"Content-Length",
	"Accept-Encoding",
	"X-CSRF-Token",
}

func AllowedCorsInterceptor() grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req any,
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (any, error) {
		err := grpc.SetHeader(ctx, metadata.Pairs(
			"Access-Control-Allow-Origin", "*",
			"Access-Control-Allow-Methods", strings.Join(AllowedMethods, ","),
			"Access-Control-Allow-Headers", strings.Join(AllowedHeaders, ","),
			"Vary", "Origin",
			"Vary", "Access-Control-Request-Headers",
			"vary", "Access-Control-Request-Method",
			"Access-Control-Expose-Headers", "Content-Disposition",
		))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to set header: %+v", err)
		}
		return handler(ctx, req)
	}
}

func HttpAllowCorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}
