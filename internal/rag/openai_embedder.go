package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
)

// openAIEmbedder implements Embedder using OpenAI's embeddings endpoint.
type openAIEmbedder struct {
	apiKey    string
	batchSize int
	client    *http.Client
}

// NewOpenAIEmbedder creates a new OpenAI-backed embedder.
// If apiKey is empty, it will read from the OPENAI_API_KEY environment variable.
// batchSize controls how many inputs are sent per request (must be >=1).
func NewOpenAIEmbedder(apiKey string, batchSize int) Embedder {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	if batchSize <= 0 {
		batchSize = 16
	}
	return &openAIEmbedder{apiKey: apiKey, batchSize: batchSize, client: http.DefaultClient}
}

type openAIRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

type openAIResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (o *openAIEmbedder) Embed(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return [][]float32{}, nil
	}
	if o.apiKey == "" {
		return nil, errors.New("openai api key not set")
	}

	var out [][]float32
	for i := 0; i < len(texts); i += o.batchSize {
		end := i + o.batchSize
		if end > len(texts) {
			end = len(texts)
		}
		batch := texts[i:end]
		batchOut, err := o.requestBatch(ctx, batch)
		if err != nil {
			return nil, err
		}
		out = append(out, batchOut...)
	}
	return out, nil
}

func (o *openAIEmbedder) requestBatch(ctx context.Context, inputs []string) ([][]float32, error) {
	reqBody := openAIRequest{Model: "text-embedding-3-small", Input: inputs}
	b, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+o.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := o.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var r openAIResponse
	// Read the body first so we can inspect status and any error payload.
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	// If non-2xx, try to extract an error message from the body.
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var maybe openAIResponse
		if err := json.Unmarshal(bodyBytes, &maybe); err == nil && maybe.Error != nil && maybe.Error.Message != "" {
			return nil, fmt.Errorf("openai returned status %d: %s", resp.StatusCode, maybe.Error.Message)
		}
		return nil, fmt.Errorf("openai returned status %d", resp.StatusCode)
	}
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		return nil, err
	}
	if r.Error != nil && r.Error.Message != "" {
		return nil, fmt.Errorf("openai error: %s", r.Error.Message)
	}
	if len(r.Data) != len(inputs) {
		// OpenAI usually returns same length; guard nonetheless.
		return nil, fmt.Errorf("openai returned %d embeddings for %d inputs", len(r.Data), len(inputs))
	}
	out := make([][]float32, len(r.Data))
	for i, d := range r.Data {
		vec := make([]float32, len(d.Embedding))
		for j, v := range d.Embedding {
			vec[j] = float32(v)
		}
		out[i] = vec
	}
	return out, nil
}
