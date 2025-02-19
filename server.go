package gateway

import (
	"context"
	"fmt"
	"github.com/adminium/gateway/middlewares"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
	"log"
	"math"
	"net"
	"net/http"
)

const metaContextKey = "x-gateway-meta-context"

type Route struct {
	Method  string
	Path    string
	Handler runtime.HandlerFunc
}

func AddMetaValue(ctx context.Context, key, value string) context.Context {
	if ctx.Value(metaContextKey) == nil {
		ctx = context.WithValue(ctx, metaContextKey, map[string]string{})
	}
	r := ctx.Value(metaContextKey).(map[string]string)
	r[key] = value
	ctx = context.WithValue(ctx, metaContextKey, r)
	return ctx
}

type GrpcGatewayService interface {
	RegisterGrpcService(server *grpc.Server)
	RegisterGatewayHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

func NewGrpcGatewayStarter() *GrpcGatewayStarter {
	return &GrpcGatewayStarter{
		grpcMiddlewares: []grpc.UnaryServerInterceptor{
			middlewares.PanicRecoveryInspector(),
			middlewares.GrpcErrorInterceptor(),
			middlewares.APIInspector(),
		},
	}
}

type GrpcGatewayStarter struct {
	httpServerAddr      string
	grpcServerAddr      string
	grpcGatewayServices []GrpcGatewayService
	grpcMiddlewares     []grpc.UnaryServerInterceptor
	middlewares         []func(handler http.Handler) http.Handler
	cors                bool
	routes              []Route
}

func (g *GrpcGatewayStarter) WithCors() *GrpcGatewayStarter {
	g.cors = true
	return g
}

func (g *GrpcGatewayStarter) WithRoute(method, path string, handler runtime.HandlerFunc) *GrpcGatewayStarter {
	g.routes = append(g.routes, Route{method, path, handler})
	return g
}

func (g *GrpcGatewayStarter) WithHttpServerAddr(addr string) *GrpcGatewayStarter {
	g.httpServerAddr = addr
	return g
}

func (g *GrpcGatewayStarter) WithGrpcServerAddr(addr string) *GrpcGatewayStarter {
	g.grpcServerAddr = addr
	return g
}

func (g *GrpcGatewayStarter) WithGrpcGatewayService(service GrpcGatewayService) *GrpcGatewayStarter {
	g.grpcGatewayServices = append(g.grpcGatewayServices, service)
	return g
}

func (g *GrpcGatewayStarter) WithGrpcMiddleware(middleware ...grpc.UnaryServerInterceptor) *GrpcGatewayStarter {
	g.grpcMiddlewares = append(g.grpcMiddlewares, middleware...)
	return g
}

func (g *GrpcGatewayStarter) WithMiddleware(middleware ...func(handler http.Handler) http.Handler) *GrpcGatewayStarter {
	g.middlewares = append(g.middlewares, middleware...)
	return g
}

func (g *GrpcGatewayStarter) validateConfig() (err error) {
	if g.grpcServerAddr == "" {
		err = fmt.Errorf("grpcServerAddr is empty, please set it with .WithGrpcServerAddr")
		return
	}
	if g.httpServerAddr == "" {
		err = fmt.Errorf("httpServerAddr is empty, please set it with .WithHttpServerAddr")
		return
	}
	return
}

func (g *GrpcGatewayStarter) Start() (err error) {

	if err = g.validateConfig(); err != nil {
		return
	}

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(g.grpcMiddlewares...))
	for _, grpcService := range g.grpcGatewayServices {
		grpcService.RegisterGrpcService(grpcServer)
	}
	ctx := context.Background()
	var tcpListener net.Listener
	tcpListener, err = net.Listen("tcp", g.grpcServerAddr)
	if err != nil {
		return
	}
	go func() {
		log.Println("Starting Grpc server on:", g.grpcServerAddr)
		log.Fatal(grpcServer.Serve(tcpListener))
	}()

	mux := runtime.NewServeMux(
		runtime.WithIncomingHeaderMatcher(BlacklistHeaderMatcher(map[string]struct{}{
			"Authorization": {},
			"Connection":    {},
		})),
		runtime.WithOutgoingHeaderMatcher(BlacklistHeaderMatcher(map[string]struct{}{})),
		runtime.WithErrorHandler(ErrorHandler),
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.HTTPBodyMarshaler{
			Marshaler: &runtime.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					EmitUnpopulated: true,
					UseProtoNames:   true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
		runtime.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			md := make(map[string]string)
			if mv := ctx.Value(metaContextKey); mv != nil {
				mm := mv.(map[string]string)
				for k, v := range mm {
					md[k] = v
				}
			}
			return metadata.New(md)
		}),
	)

	for _, route := range g.routes {
		err = mux.HandlePath(route.Method, route.Path, route.Handler)
		if err != nil {
			err = fmt.Errorf("handle path error: %s", err)
			return
		}
	}

	conn, err := grpc.NewClient(
		g.grpcServerAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(104857600), grpc.MaxCallSendMsgSize(math.MaxInt32)), // 100MB for receive
	)
	if err != nil {
		return
	}
	for _, grpcService := range g.grpcGatewayServices {
		err = grpcService.RegisterGatewayHandler(ctx, mux, conn)
		if err != nil {
			return err
		}
	}

	var h http.Handler = mux
	for _, m := range g.middlewares {
		h = m(h)
	}

	if g.cors {
		h = httpAllowCorsHandler(h)
	}

	log.Println("Starting Http server on:", g.httpServerAddr)

	go func() {
		log.Fatal(http.ListenAndServe(g.httpServerAddr, h))
	}()

	return
}
