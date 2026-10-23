# Postmortem RAG

A Go-based retrieval-augmented generation app that ingests postmortem markdown, embeds it with OpenAI, stores it in PostgreSQL with pgvector, and answers natural-language questions grounded in the source material.

I built this to make incident writeups searchable and usable in real time. When a new outage or failure pattern appears, the system helps surface the relevant historical context instead of forcing someone to dig through scattered documents by hand.

Stack: Go, PostgreSQL + pgvector, OpenAI embeddings + chat completions, GraphQL, Docker, Kubernetes, Terraform

## Architecture

The app turns markdown postmortems into searchable chunks, embeds those chunks, and ranks them by semantic similarity before generating a final answer from the most relevant excerpts.

```text
Markdown postmortems
        |
        v
   cmd/ingest
        |
        v
  Chunk + embed
        |
        v
 PostgreSQL + pgvector
        |
        v
  Retriever / Answerer
        |
   +---------------------+
   | CLI or GraphQL API  |
   +---------------------+
```

At runtime, the flow is:

1. Postmortem files are read from the corpus.
2. Text is chunked with overlap and stripped of frontmatter.
3. Chunks are embedded with OpenAI.
4. Similar chunks are retrieved from pgvector.
5. The answer layer passes the top results to the generator and returns a citation-backed answer.

## Running it

This repo has been validated with the following commands.

### 1) Start PostgreSQL

```bash
docker compose -f deploy/docker-compose.yml up -d
```

### 2) Set environment variables

The repo expects `OPENAI_API_KEY` and `DATABASE_URL` to be present in the environment. A local `.env` is used in development.

```bash
set -a
source .env
set +a
```

### 3) Run database migrations

```bash
go run github.com/pressly/goose/v3/cmd/goose@latest -dir migrations postgres "$DATABASE_URL" up
```

### 4) Ingest the corpus

```bash
go run ./cmd/ingest
```

### 5) Ask a question from the CLI

```bash
go run ./cmd/ask "What caused the incident?"
```

### 6) Inspect the best matching chunks directly

```bash
go run ./cmd/query "What caused the incident?"
```

### 7) Start the GraphQL server

```bash
go run ./cmd/server
```

The server listens on port `8080` by default and exposes the GraphQL playground at `/` and the API at `/query`.

## Deployment

The app is containerized with a multi-stage Docker build that compiles a static Go binary for the runtime image. It is designed to run in a Kubernetes cluster, including local kind-based testing, and the cluster resources are provisioned through Terraform rather than ad hoc YAML alone. This keeps the deployment path consistent with the local development flow while making production-grade orchestration straightforward.

## Useful checks

```bash
go build ./...
go test ./...
```

These verify the project compiles cleanly and the retrieval logic still passes its tests.
