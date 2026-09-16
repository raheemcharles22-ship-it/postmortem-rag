package rag

import "context"

// Embedder produces vector embeddings for input texts.
type Embedder interface {
	Embed(ctx context.Context, texts []string) ([][]float32, error)
}

// EmbeddingDim is the dimension for text-embedding-3-small.
// If OpenAI changes the model, update this constant accordingly.
const EmbeddingDim = 1536
