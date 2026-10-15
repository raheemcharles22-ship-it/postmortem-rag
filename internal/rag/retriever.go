package rag

import "context"

// Retriever is an interface for retrieving scored chunks for a question.
type Retriever interface {
    Retrieve(ctx context.Context, question string, k int) ([]ScoredChunk, error)
}
