// cmd/server/main.go
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/raheemcharles/postmortem-rag/internal/rag"
)

type config struct {
	databaseURL string
	apiKey      string
	corpusDir   string
	chunkSize   int
	overlap     int
}

func loadConfig() config {
	return config{
		databaseURL: mustEnv("DATABASE_URL"),
		apiKey:      mustEnv("OPENAI_API_KEY"),
		corpusDir:   envOr("CORPUS_DIR", "corpus"),
		chunkSize:   envIntOr("CHUNK_SIZE", 512),
		overlap:     envIntOr("CHUNK_OVERLAP", 77), // ~15%
	}
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required env var %s not set", key)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envIntOr(key string, def int) int {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		log.Fatalf("env var %s must be an integer, got %q", key, v)
	}
	return n
}

func run(ctx context.Context, cfg config, embedder rag.Embedder, store rag.VectorStore) error {
	files, err := filepath.Glob(filepath.Join(cfg.corpusDir, "*.md"))
	if err != nil {
		return fmt.Errorf("globbing corpus: %w", err)
	}
	if len(files) == 0 {
		return fmt.Errorf("no .md files in %s", cfg.corpusDir)
	}

	totalChunks := 0
	for _, path := range files {
		n, err := ingestFile(ctx, path, embedder, store, cfg.chunkSize, cfg.overlap)
		if err != nil {
			return fmt.Errorf("ingesting %s: %w", path, err)
		}
		totalChunks += n
		log.Printf("ingested %s: %d chunks", filepath.Base(path), n)
	}
	log.Printf("done: %d files, %d chunks", len(files), totalChunks)
	return nil
}

func ingestFile(ctx context.Context, path string, embedder rag.Embedder, store rag.VectorStore, size, overlap int) (int, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return 0, fmt.Errorf("reading file: %w", err)
	}
	source := filepath.Base(path)
	text := rag.StripFrontmatter(string(data))
	chunks, err := rag.ChunkText(text, source, size, overlap)
	if err != nil {
		return 0, fmt.Errorf("chunking: %w", err)
	}
	if len(chunks) == 0 {
		return 0, nil
	}

	texts := make([]string, len(chunks))
	for i := range chunks {
		texts[i] = chunks[i].Text
	}

	vecs, err := embedder.Embed(ctx, texts)
	if err != nil {
		return 0, fmt.Errorf("embedding: %w", err)
	}

	if err := store.Upsert(ctx, chunks, vecs); err != nil {
		return 0, fmt.Errorf("upserting: %w", err)
	}

	return len(chunks), nil
}

func main() {
	cfg := loadConfig()
	ctx := context.Background()

	embedder := rag.NewOpenAIEmbedder(cfg.apiKey, 16)
	store, err := rag.NewPgxStore(ctx, cfg.databaseURL)
	if err != nil {
		log.Fatalf("store: %v", err)
	}

	if err := run(ctx, cfg, embedder, store); err != nil {
		log.Fatalf("ingest: %v", err)
	}
}
