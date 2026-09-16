package rag

import (
	"context"
	"os"
	"testing"
)

func testStore(t *testing.T) VectorStore {
	t.Helper()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	// construct your store from url here.
	// fail the test (not skip) if construction errors.
	panic("wire your store constructor")
}

func TestVectorStore_SearchReturnsNearestFirst(t *testing.T) {
	store := testStore(t)
	ctx := context.Background()

	// Three chunks with hand-built vectors so you KNOW the ranking.
	// Vectors are EmbeddingDim long. Make the first dimension carry the signal.
	near := makeVec(1.0) // closest to query
	mid := makeVec(0.5)
	far := makeVec(0.0) // farthest

	chunks := []Chunk{
		{Text: "near", Index: 0, Source: "doc-1"},
		{Text: "mid", Index: 1, Source: "doc-1"},
		{Text: "far", Index: 2, Source: "doc-1"},
	}
	vecs := [][]float32{near, mid, far}

	if err := store.Upsert(ctx, chunks, vecs); err != nil {
		t.Fatalf("upsert: %v", err)
	}

	query := makeVec(1.0) // identical to `near`
	got, err := store.Search(ctx, query, 2)
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d results, want 2 (k=2)", len(got))
	}
	if got[0].Text != "near" {
		t.Fatalf("rank 1 is %q, want \"near\"", got[0].Text)
	}
	if got[0].Source != "doc-1" {
		t.Errorf("lost source on retrieved chunk")
	}
}

// makeVec returns an EmbeddingDim vector with the first element set to v.
func makeVec(v float32) []float32 {
	out := make([]float32, EmbeddingDim)
	out[0] = v
	return out
}
