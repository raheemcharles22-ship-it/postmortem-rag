package rag

import (
	"context"
	"fmt"
)

type retrieverImpl struct {
	embedder Embedder
	store    VectorStore
}

// NewRetriever constructs a Retriever backed by an Embedder and VectorStore.
func NewRetriever(e Embedder, s VectorStore) Retriever {
	return &retrieverImpl{embedder: e, store: s}
}

func (r *retrieverImpl) Retrieve(ctx context.Context, question string, k int) ([]ScoredChunk, error) {
	vecs, err := r.embedder.Embed(ctx, []string{question})
	if err != nil {
		return nil, fmt.Errorf("embedded question: %w", err)
	}
	if len(vecs) == 0 {
		return nil, fmt.Errorf("no vectors")
	}
	return r.store.Search(ctx, vecs[0], k)
}
