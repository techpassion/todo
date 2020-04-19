package server

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
	"todo/server/graphql/graph"
	"todo/server/graphql/graph/generated"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"
)

const defaultPort = "8080"

func StartGQLServer() {
	quit := make(chan os.Signal)
	signal.Notify(quit, os.Interrupt)

	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	srv := handler.NewDefaultServer(generated.NewExecutableSchema(generated.Config{Resolvers: &graph.Resolver{}}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("connect to http://localhost:%s/ for GraphQL playground", port)
	server := &http.Server{Addr: ":" + port, Handler: nil}
	go func() {
		log.Println("server starting")
		if err := server.ListenAndServe(); err != nil {
			log.Fatalf("listenAndServe failed: %v", err)
		}

	}()

	<-quit
	// gracefully stop server
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
	log.Println("server stopped")
}
