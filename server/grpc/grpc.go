package server

import (
	"context"
	"log"
	"todo/utils"

	pb "github.com/ganeshdipdumbare/protos/proto/todo"

	"google.golang.org/grpc"
)

type server struct{}

func (s *server) Login(context.Context, *pb.LoginRequest) (*pb.LoginResponse, error) {
	return &pb.LoginResponse{
		Response: &pb.LoginResponse_Token{
			Token: "Success token",
		},
	}, nil
}

func ServerInitialization() {
	// if we crash the go code, we get the file name and line number
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	builder := utils.GrpcServerBuilder{}
	addInterceptors(&builder)
	builder.EnableReflection(true)

	s := builder.Build()
	s.RegisterService(serviceRegister)

	err := s.Start("0.0.0.0", 9090)
	if err != nil {
		log.Fatalf("%v", err)
	}

	s.AwaitTermination(func() {
		log.Print("Shutting down the server")
	})
}

func serviceRegister(sv *grpc.Server) {
	pb.RegisterTodoServiceServer(sv, &server{})
}

func addInterceptors(s *utils.GrpcServerBuilder) {
	s.SetUnaryInterceptors(utils.GetDefaultUnaryServerInterceptors())
	s.SetStreamInterceptors(utils.GetDefaultStreamServerInterceptors())
}
