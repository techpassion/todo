package main

import (
	"sync"
	graphql "todo/server/graphql"
	grpc "todo/server/grpc"
)

func main() {
	var wg sync.WaitGroup

	wg.Add(2)
	go func() {
		grpc.ServerInitialization()
		wg.Done()
	}()

	go func() {
		graphql.StartGQLServer()
		wg.Done()
	}()

	wg.Wait()
}
