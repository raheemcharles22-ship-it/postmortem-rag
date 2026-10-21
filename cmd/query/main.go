package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/raheemcharles/postmortem-rag/internal/rag"
)

type retriever struct {
	embedder rag.Embedder
	store    rag.VectorStore
}

func NewRetriever(e rag.Embedder, s rag.VectorStore) rag.Retriever {
	return &retriever{embedder: e, store: s}
}

func (r *retriever) Retrieve(ctx context.Context, question string, k int) ([]rag.ScoredChunk, error) {
	vecs, err := r.embedder.Embed(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("embedded question: %w", err)
	}

	if len(vecs) == 0 {
		return nil, fmt.Errorf("no vectors")
	}

	return r.store.Search(ctx, vecs[0], k)
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/query \"your question\"")
	}
	question := os.Args[1]

	ctx := context.Background()
	embedder := rag.NewOpenAIEmbedder(os.Getenv("OPENAI_API_KEY"), 16)
	store, err := rag.NewPgxStore(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("store: %v", err)
	}
	r := NewRetriever(embedder, store)

	results, err := r.Retrieve(ctx, question, 5)
	if err != nil {
		log.Fatalf("retrieve: %v", err)
	}

	for i, sc := range results {
		preview := sc.Text
		if len(preview) > 100 {
			preview = preview[:100]
		}
		fmt.Printf("%d. [%.4f] %s\n   %s\n", i+1, sc.Score, sc.Source, preview)
	}
}
