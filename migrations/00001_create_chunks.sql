-- +goose Up
-- no vector index at MVP scale; seq scan is exact and fast. add HNSW past ~10k rows.
CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE chunks (
    chunk_text text,
    source    text,
    chunk_index integer,
    embedding vector(1536),
    PRIMARY KEY (source, chunk_index)
);
-- +goose Down
DROP TABLE chunks;