package rag

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

type Generator interface {
	Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error)
}

type openAIGenerator struct {
	apiKey string
	model  string
	client *http.Client
}

func NewOpenAIGenerator(apiKey string) Generator {
	if apiKey == "" {
		apiKey = os.Getenv("OPENAI_API_KEY")
	}
	return &openAIGenerator{
		apiKey: apiKey,
		model:  "gpt-4o-mini", // mid-tier chat model; confirm current ID
		client: http.DefaultClient,
	}
}

// chat completions request/response shapes
type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (g *openAIGenerator) Generate(ctx context.Context, systemPrompt, userPrompt string) (string, error) {
	if g.apiKey == "" {
		return "", fmt.Errorf("openai api key not set")
	}

	reqBody := chatRequest{
		Model: g.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userPrompt},
		},
	}

	// --- perform request like the embedder's requestBatch ---
	b, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+g.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var maybe chatResponse
		if err := json.Unmarshal(bodyBytes, &maybe); err == nil && maybe.Error != nil && maybe.Error.Message != "" {
			return "", fmt.Errorf("openai returned status %d: %s", resp.StatusCode, maybe.Error.Message)
		}
		return "", fmt.Errorf("openai returned status %d", resp.StatusCode)
	}

	var r chatResponse
	if err := json.Unmarshal(bodyBytes, &r); err != nil {
		return "", err
	}
	if r.Error != nil && r.Error.Message != "" {
		return "", fmt.Errorf("openai error: %s", r.Error.Message)
	}
	if len(r.Choices) == 0 {
		return "", fmt.Errorf("openai returned no choices")
	}
	return r.Choices[0].Message.Content, nil
}
