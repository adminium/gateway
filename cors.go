package gateway

import (
	"net/http"
	"strings"
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

var ExposedHeaders = []string{
	"Content-Length",
	"Access-Control-Allow-Origin",
	"Access-Control-Allow-Headers",
	"Cache-Control",
	"Content-Language",
	"Content-Type",
}

var ResponseHeaders = map[string]string{
	"Access-Control-Allow-Methods":     strings.Join(AllowedMethods, ","),
	"Access-Control-Allow-Headers":     strings.Join(AllowedHeaders, ","),
	"Access-Control-Expose-Headers":    strings.Join(ExposedHeaders, ","),
	"Access-Control-Allow-Credentials": "true",
}

func httpAllowCorsHandler(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Method
		origin := r.Header.Get("Origin")
		if origin != "" {
			if ResponseHeaders["Access-Control-Allow-Origin"] != "*" {
				ResponseHeaders["Access-Control-Allow-Origin"] = origin
			}
			for k, v := range ResponseHeaders {
				w.Header().Set(k, v)
			}
		}
		if method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		h.ServeHTTP(w, r)
	})
}
