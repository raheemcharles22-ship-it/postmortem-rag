package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/raheemcharles/postmortem-rag/internal/rag"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/ask \"your question\"")
	}
	question := os.Args[1]

	ctx := context.Background()

	embedder := rag.NewOpenAIEmbedder(os.Getenv("OPENAI_API_KEY"), 16)
	store, err := rag.NewPgxStore(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	generator := rag.NewOpenAIGenerator(os.Getenv("OPENAI_API_KEY")) // <- match your real constructor

	retriever := rag.NewRetriever(embedder, store)
	answerer := rag.NewAnswerer(retriever, generator, 0.6)

	ans, err := answerer.Answer(ctx, question, 5)
	if err != nil {
		log.Fatalf("answer: %v", err)
	}

	if ans.Refused {
		fmt.Println("REFUSED:", ans.Text)
		return
	}

	fmt.Println("ANSWER:")
	fmt.Println(ans.Text)
	fmt.Println("\nCITATIONS:")
	for _, c := range ans.Citations {
		fmt.Println("  -", c)
	}
}
