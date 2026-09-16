package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"testing"
)

type fakeTransport struct {
	calls int
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	f.calls++
	// Read request body to determine number of inputs
	body, _ := io.ReadAll(req.Body)
	var reqBody struct {
		Input []string `json:"input"`
	}
	_ = json.Unmarshal(body, &reqBody)
	n := len(reqBody.Input)

	// Build response with n embeddings each of length EmbeddingDim
	resp := struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		} `json:"data"`
	}{}
	resp.Data = make([]struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	}, n)
	for i := 0; i < n; i++ {
		resp.Data[i].Embedding = make([]float64, EmbeddingDim)
		resp.Data[i].Index = i
		for j := 0; j < EmbeddingDim; j++ {
			resp.Data[i].Embedding[j] = 0.5
		}
	}
	b, _ := json.Marshal(resp)

	r := &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(bytes.NewReader(b)),
		Header:     make(http.Header),
	}
	r.Header.Set("Content-Type", "application/json")
	return r, nil
}

func TestOpenAIEmbedder_BatchingCalls(t *testing.T) {
	batchSize := 5
	total := 23
	inputs := make([]string, total)
	for i := 0; i < total; i++ {
		inputs[i] = "text"
	}

	emb := NewOpenAIEmbedder("key", batchSize).(*openAIEmbedder)
	// Inject fake transport
	ft := &fakeTransport{}
	emb.client = &http.Client{Transport: ft}

	got, err := emb.Embed(context.Background(), inputs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(got) != total {
		t.Fatalf("got %d embeddings, want %d", len(got), total)
	}
	// Expected number of calls is ceil(total/batchSize)
	wantCalls := (total + batchSize - 1) / batchSize
	if ft.calls != wantCalls {
		t.Fatalf("transport called %d times, want %d", ft.calls, wantCalls)
	}
	// dimension check
	if len(got[0]) != EmbeddingDim {
		t.Fatalf("vector dim %d, want %d", len(got[0]), EmbeddingDim)
	}
}
