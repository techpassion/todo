package graph

// This file will be automatically regenerated based on the schema, any resolver implementations
// will be copied through when generating and any unknown code will be moved to the end.

import (
	"context"
	"log"
	"time"
	"todo/server/graphql/graph/generated"
	"todo/server/graphql/graph/model"

	pb "github.com/ganeshdipdumbare/protos/proto/todo"
	"google.golang.org/grpc"
)

const (
	address     = "localhost:9090"
	defaultName = "world"
)

func (r *queryResolver) Login(ctx context.Context, input model.LoginRequest) (*model.LoginResponse, error) {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewTodoServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.Login(ctx, &pb.LoginRequest{Username: input.Username, Password: input.Password})
	if err != nil {
		log.Fatalf("could not greet: %v", err)
	}
	log.Printf("Greeting: %s", resp.GetToken())
	return &model.LoginResponse{
		Token: resp.GetToken(),
	}, nil
}

// Query returns generated.QueryResolver implementation.
func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }

type queryResolver struct{ *Resolver }
