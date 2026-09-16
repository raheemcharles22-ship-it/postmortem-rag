package rag

import (
	"context"
	"testing"
)

func TestFakeEmbedder_Dimensions(t *testing.T) {
	f := NewFakeEmbedder(0.123)
	inputs := []string{"a", "b", "c"}
	got, err := f.Embed(context.Background(), inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != len(inputs) {
		t.Fatalf("got %d vectors, want %d", len(got), len(inputs))
	}
	for i := range got {
		if len(got[i]) != EmbeddingDim {
			t.Fatalf("vector %d len %d, want %d", i, len(got[i]), EmbeddingDim)
		}
		// spot check value
		if got[i][0] != 0.123 {
			t.Fatalf("unexpected value: %v", got[i][0])
		}
	}
}
