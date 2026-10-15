package rag

import (
	"context"
	"testing"
)

// in answer_test.go
type fakeRetriever struct {
	results []ScoredChunk
}

func (f *fakeRetriever) Retrieve(ctx context.Context, question string, k int) ([]ScoredChunk, error) {
	return f.results, nil
}

func TestAnswerer_RefusesWhenBestScoreTooFar(t *testing.T) {
	ret := &fakeRetriever{
		results: []ScoredChunk{
			{Chunk: Chunk{Text: "irrelevant", Source: "x.md"}, Score: 0.9},
		},
	}
	a := NewAnswerer(ret, NewFakeGenerator(), 0.6)

	ans, err := a.Answer(context.Background(), "banana bread", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ans.Refused {
		t.Fatalf("score 0.9 with threshold 0.6: want refused, got answered")
	}
}

func TestAnswerer_AnswersWhenScoreClose(t *testing.T) {
	ret := &fakeRetriever{
		results: []ScoredChunk{
			{Chunk: Chunk{Text: "relevant", Source: "db.md"}, Score: 0.4},
		},
	}
	a := NewAnswerer(ret, NewFakeGenerator(), 0.6)

	ans, err := a.Answer(context.Background(), "database outage", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ans.Refused {
		t.Fatalf("score 0.4 with threshold 0.6: want answered, got refused")
	}
}

func TestAnswerer_RefusesWhenEmpty(t *testing.T) {
	ret := &fakeRetriever{results: nil}
	a := NewAnswerer(ret, NewFakeGenerator(), 0.6)

	ans, err := a.Answer(context.Background(), "anything", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ans.Refused {
		t.Fatalf("empty retrieve: want refused, got answered")
	}
}

func TestAnswerer_AnswersWithGenerator(t *testing.T) {
	ret := &fakeRetriever{
		results: []ScoredChunk{
			{Chunk: Chunk{Text: "the db failed over", Source: "db.md"}, Score: 0.4},
			{Chunk: Chunk{Text: "retry storm followed", Source: "db.md"}, Score: 0.45},
		},
	}
	a := NewAnswerer(ret, NewFakeGenerator(), 0.6)

	ans, err := a.Answer(context.Background(), "what happened", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ans.Refused {
		t.Fatal("close score should not refuse")
	}
	if ans.Text == "" {
		t.Fatal("expected generated text, got empty")
	}
	if len(ans.Citations) == 0 {
		t.Fatal("expected citations, got none")
	}
}
