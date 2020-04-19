package utils

import (
	"errors"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/apssouza22/grpc-server-go/clientinterceptor"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/opentracing/opentracing-go"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

//Fiji GRPC server interface
type GrpcServer interface {
	Start(address string, port uint) error
	AwaitTermination(shutdownHook func())
	RegisterService(reg func(*grpc.Server))
}

//GRPC server builder
type GrpcServerBuilder struct {
	options            []grpc.ServerOption
	enabledReflection  bool
	shutdownHook       func()
	enabledHealthCheck bool
}

type grpcServer struct {
	server   *grpc.Server
	listener net.Listener
}

//DialOption configures how we set up the connection.
func (sb *GrpcServerBuilder) AddOption(o grpc.ServerOption) {
	sb.options = append(sb.options, o)
}

// EnableReflection enables the reflection
// gRPC Server Reflection provides information about publicly-accessible gRPC services on a server,
// and assists clients at runtime to construct RPC requests and responses without precompiled service information.
// It is used by gRPC CLI, which can be used to introspect server protos and send/receive test RPCs.
//Warning! We should not have this enabled in production
func (sb *GrpcServerBuilder) EnableReflection(e bool) {
	sb.enabledReflection = e
}

// ServerParameters is used to set keepalive and max-age parameters on the server-side.
func (sb *GrpcServerBuilder) SetServerParameters(serverParams keepalive.ServerParameters) {
	keepAlive := grpc.KeepaliveParams(serverParams)
	sb.AddOption(keepAlive)
}

// SetStreamInterceptors set a list of interceptors to the Grpc server for stream connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (sb *GrpcServerBuilder) SetStreamInterceptors(interceptors []grpc.StreamServerInterceptor) {
	chain := grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(interceptors...))
	sb.AddOption(chain)
}

// SetUnaryInterceptors set a list of interceptors to the Grpc server for unary connection
// By default, gRPC doesn't allow one to have more than one interceptor either on the client nor on the server side.
// By using `grpc_middleware` we are able to provides convenient method to add a list of interceptors
func (sb *GrpcServerBuilder) SetUnaryInterceptors(interceptors []grpc.UnaryServerInterceptor) {
	chain := grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...))
	sb.AddOption(chain)
}

//Build is responsible for building a Fiji GRPC server
func (sb *GrpcServerBuilder) Build() GrpcServer {
	srv := grpc.NewServer(sb.options...)
	grpc_health_v1.RegisterHealthServer(srv, health.NewServer())

	if sb.enabledReflection {
		reflection.Register(srv)
	}
	return &grpcServer{srv, nil}
}

// RegisterService register the services to the server
func (s grpcServer) RegisterService(reg func(*grpc.Server)) {
	reg(s.server)
}

// Start the GRPC server
func (s *grpcServer) Start(address string, port uint) error {
	var err error
	add := fmt.Sprintf("%s:%d", address, port)
	s.listener, err = net.Listen("tcp", add)

	if err != nil {
		msg := fmt.Sprintf("Failed to listen: %v", err)
		return errors.New(msg)
	}

	go s.serv()

	log.Infof("grpcServer started on port: %d ", port)
	return nil
}

// AwaitTermination makes the program wait for the signal termination
// Valid signal termination (SIGINT, SIGTERM)
func (s *grpcServer) AwaitTermination(shutdownHook func()) {
	interruptSignal := make(chan os.Signal, 1)
	signal.Notify(interruptSignal, syscall.SIGINT, syscall.SIGTERM)
	<-interruptSignal
	s.cleanup()
	if shutdownHook != nil {
		shutdownHook()
	}
}

func (s *grpcServer) cleanup() {
	log.Info("Stopping the server")
	s.server.GracefulStop()
	log.Info("Closing the listener")
	s.listener.Close()
	log.Info("End of Program")
}

func (s *grpcServer) serv() {
	if err := s.server.Serve(s.listener); err != nil {
		log.Errorf("failed to serve: %v", err)
	}
}

func requestErrorHandler(p interface{}) (err error) {
	logrus.Error(p)
	return status.Errorf(codes.Internal, "Something went wrong :( ")
}

// GetDefaultUnaryServerInterceptors returns the default interceptors server unary connections
func GetDefaultUnaryServerInterceptors() []grpc.UnaryServerInterceptor {
	return []grpc.UnaryServerInterceptor{
		//Recovery handlers should typically be last in the chain so that other middleware
		// (e.g. logging) can operate on the recovered state instead of being directly affected by any panic
		grpc_recovery.UnaryServerInterceptor(grpc_recovery.WithRecoveryHandler(requestErrorHandler)),
	}
}

// GetDefaultStreamServerInterceptors returns the default interceptors for server streams connections
func GetDefaultStreamServerInterceptors() []grpc.StreamServerInterceptor {
	return []grpc.StreamServerInterceptor{
		grpc_recovery.StreamServerInterceptor(grpc_recovery.WithRecoveryHandler(requestErrorHandler)),
	}
}

//GetDefaultUnaryClientInterceptors returns the default interceptors for client unary connections
func GetDefaultUnaryClientInterceptors() []grpc.UnaryClientInterceptor {
	tracing := grpc_opentracing.UnaryClientInterceptor(
		grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
	)
	interceptors := []grpc.UnaryClientInterceptor{
		clientinterceptor.UnaryTimeoutInterceptor(),
		tracing,
	}
	return interceptors
}

//GetDefaultStreamClientInterceptors returns the default interceptors for client stream connections
func GetDefaultStreamClientInterceptors() []grpc.StreamClientInterceptor {
	tracing := grpc_opentracing.StreamClientInterceptor(
		grpc_opentracing.WithTracer(opentracing.GlobalTracer()),
	)
	interceptors := []grpc.StreamClientInterceptor{
		clientinterceptor.StreamTimeoutInterceptor(),
		tracing,
	}
	return interceptors
}
