package gateway

import (
	"context"
	"fmt"
	"github.com/adminium/gateway/middlewares"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"math"
	"net"
	"net/http"
)

type GrpcGatewayService interface {
	RegisterGrpcService(server *grpc.Server)
	RegisterGatewayHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

func NewGrpcGatewayStarter() *GrpcGatewayStarter {
	return &GrpcGatewayStarter{
		middlewares: []grpc.UnaryServerInterceptor{
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
	middlewares         []grpc.UnaryServerInterceptor
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

func (g *GrpcGatewayStarter) WithMiddleware(middleware ...grpc.UnaryServerInterceptor) *GrpcGatewayStarter {
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

	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(g.middlewares...))
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
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONBuiltin{}),
	)

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

	h := middlewares.HttpAllowCorsHandler(mux)

	log.Println("Starting Http server on:", g.httpServerAddr)

	go func() {
		log.Fatal(http.ListenAndServe(g.httpServerAddr, h))
	}()

	return
}
