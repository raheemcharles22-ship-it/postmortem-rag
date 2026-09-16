package rag

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pgvector/pgvector-go"
)

type pgxStore struct {
	pool *pgxpool.Pool
}

type ScoredChunk struct {
	Chunk
	Score float32
}

type VectorStore interface {
	Upsert(ctx context.Context, chunks []Chunk, vecs [][]float32) error
	Search(ctx context.Context, query []float32, k int) ([]ScoredChunk, error)
}

func NewPgxStore(ctx context.Context, url string) (VectorStore, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, fmt.Errorf("creating pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging postgres: %w", err)
	}
	return &pgxStore{pool: pool}, nil
}

func (s *pgxStore) Upsert(ctx context.Context, chunks []Chunk, vecs [][]float32) error {

	if len(chunks) != len(vecs) {
		return fmt.Errorf("mismatch lengths")
	}

	sql := `
    INSERT INTO chunks (source, chunk_index, chunk_text, embedding)
    VALUES ($1, $2, $3, $4)
    ON CONFLICT (source, chunk_index) DO UPDATE
    SET chunk_text = EXCLUDED.chunk_text, embedding = EXCLUDED.embedding`

	for i, chunk := range chunks {
		_, err := s.pool.Exec(ctx, sql,
			chunk.Source,
			chunk.Index,
			chunk.Text,
			pgvector.NewVector(vecs[i]),
		)
		if err != nil {
			return fmt.Errorf("upsert chunk %d: %w", i, err)
		}
	}
	return nil
}

func (s *pgxStore) Search(ctx context.Context, query []float32, k int) ([]ScoredChunk, error) {
	sql := `
        SELECT chunk_text, source, chunk_index, embedding <=> $1 AS score
        FROM chunks
        ORDER BY embedding <=> $1
        LIMIT $2`

	rows, err := s.pool.Query(ctx, sql, pgvector.NewVector(query), k)
	if err != nil {
		return nil, fmt.Errorf("search query: %w", err)
	}
	defer rows.Close()

	var out []ScoredChunk
	for rows.Next() {
		var sc ScoredChunk
		if err := rows.Scan(&sc.Text, &sc.Source, &sc.Index, &sc.Score); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		out = append(out, sc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration: %w", err)
	}
	return out, nil
}
