package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/playground"

	"github.com/raheemcharles/postmortem-rag/graph" // <- your module path
	"github.com/raheemcharles/postmortem-rag/internal/rag"
)

const defaultPort = "8080"

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = defaultPort
	}

	ctx := context.Background()

	// same wiring as cmd/ask
	embedder := rag.NewOpenAIEmbedder(os.Getenv("OPENAI_API_KEY"), 16)
	store, err := rag.NewPgxStore(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	generator := rag.NewOpenAIGenerator(os.Getenv("OPENAI_API_KEY"))
	retriever := rag.NewRetriever(embedder, store)
	answerer := rag.NewAnswerer(retriever, generator, 0.6)

	// hand the answerer to the resolver
	resolver := &graph.Resolver{Answerer: answerer}

	srv := handler.NewDefaultServer(graph.NewExecutableSchema(graph.Config{Resolvers: resolver}))

	http.Handle("/", playground.Handler("GraphQL playground", "/query"))
	http.Handle("/query", srv)

	log.Printf("listening on http://localhost:%s/", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
