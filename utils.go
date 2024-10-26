package gateway

import (
	"context"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

func ErrorHandler(ctx context.Context,
	sm *runtime.ServeMux,
	marshaler runtime.Marshaler,
	w http.ResponseWriter,
	r *http.Request,
	err error) {
	w.Header().Set("connection", "close")

	s := status.Convert(err)
	protoStatus := s.Proto()

	contentType := marshaler.ContentType(protoStatus)
	w.Header().Set("content-type", contentType)

	httpStatus := runtime.HTTPStatusFromCode(s.Code())

	w.WriteHeader(httpStatus)

	protoResp, err := marshaler.Marshal(protoStatus)
	if err != nil {
		log.Println("Error while marshaling error response, error	%s", err)
		return
	}
	_, err = w.Write(protoResp)

	if err != nil {
		log.Println("Error while writing error response: %s", err)
	}
}

func BlacklistHeaderMatcher(blacklist map[string]struct{}) runtime.HeaderMatcherFunc {
	return func(key string) (string, bool) {
		_, ok := blacklist[key]
		if ok {
			return "", false
		}
		return key, true
	}
}
