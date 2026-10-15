package rag

import (
	"context"
	"fmt"
	"strings"
)

type Answerer struct {
	retriever Retriever
	generator Generator
	threshold float32
}

type Answer struct {
	Text      string
	Citations []string
	Refused   bool
}

func NewAnswerer(r Retriever, g Generator, threshold float32) *Answerer {
	return &Answerer{retriever: r, generator: g, threshold: threshold}
}

func (a *Answerer) Answer(ctx context.Context, question string, k int) (Answer, error) {
	results, err := a.retriever.Retrieve(ctx, question, k)
	if err != nil {
		return Answer{}, err
	}
	if len(results) == 0 {
		return Answer{Refused: true, Text: "The postmortem found no responses"}, nil
	}
	if results[0].Score > a.threshold {
		return Answer{Refused: true, Citations: nil, Text: "The provided postmortems don't cover this question."}, nil
	}
	// 3. build prompt: assemble context from top-k retrieved chunks
	var ctxBuilder strings.Builder
	for i, sc := range results {
		// prefix each chunk with its source so the generator can cite it
		ctxBuilder.WriteString(fmt.Sprintf("[Source: %s]\n", sc.Source))
		ctxBuilder.WriteString(sc.Text)
		if i < len(results)-1 {
			ctxBuilder.WriteString("\n\n---\n\n")
		}
	}

	systemPrompt := "You are a concise assistant that answers questions using only the provided excerpts. If the excerpts do not contain the information necessary to answer, say you don't know and do not hallucinate. Reject the question's assumptions when the excerpts don't support them."
	userPrompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s\n\nRespond with a short answer and include explicit citations in the form [Source: <source>] where relevant.", ctxBuilder.String(), question)

	// 4. generate
	generated, err := a.generator.Generate(ctx, systemPrompt, userPrompt)
	if err != nil {
		return Answer{}, err
	}

	// 5. assemble Answer with citations: collect unique sources from results
	seen := make(map[string]struct{})
	citations := make([]string, 0, len(results))
	for _, sc := range results {
		if _, ok := seen[sc.Source]; ok {
			continue
		}
		seen[sc.Source] = struct{}{}
		citations = append(citations, sc.Source)
	}

	// Trust the threshold as the structural refusal mechanism and the
	// prompt to instruct the generator. Return the generated text as-is.
	return Answer{Refused: false, Citations: citations, Text: strings.TrimSpace(generated)}, nil
}
