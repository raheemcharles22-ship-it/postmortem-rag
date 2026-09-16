package rag

import (
	"context"
)

// FakeEmbedder returns deterministic fixed-size vectors for tests.
type FakeEmbedder struct {
	// Value to fill each element with; useful for distinguishing embedders.
	Value float32
}

// NewFakeEmbedder returns a FakeEmbedder filling vectors with `value`.
func NewFakeEmbedder(value float32) *FakeEmbedder {
	return &FakeEmbedder{Value: value}
}

// Embed returns N vectors of length EmbeddingDim filled with the configured value.
func (f *FakeEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	out := make([][]float32, len(texts))
	for i := range texts {
		vec := make([]float32, EmbeddingDim)
		for j := range vec {
			vec[j] = f.Value
		}
		out[i] = vec
	}
	return out, nil
}
