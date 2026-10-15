package rag

import "context"

// FakeGenerator is a simple Generator used in tests or local runs.
type FakeGenerator struct{}

func NewFakeGenerator() *FakeGenerator { return &FakeGenerator{} }

func (f *FakeGenerator) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	return "(fake generated answer)", nil
}

