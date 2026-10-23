# Design Decisions Inventory — Postmortem-RAG

Every real decision made in this project, with alternatives and the reasoning behind it. The anchor is shorthand for the "why" already worked through; each item below expands that rationale into a clear design decision.

## Chunking

- Unit: runes, not bytes or tokens. Byte slicing would corrupt multi-byte characters, and token-aware chunking is a valid upgrade but adds a tokenizer dependency that was deliberately deferred.
- Overlap at roughly 15% of chunk size, with stride equal to size minus overlap. This keeps a thought intact across chunk boundaries without creating too much duplication.
- Keep short tails instead of discarding them. The final section of a postmortem often contains the resolution, so merging small trailing fragments avoids losing data.
- Reject invalid chunking config such as overlap >= size or size <= 0. Explicit validation is better than silent clamping or a panic, because bad configuration is a runtime contract violation that should be handled by the caller.

## Embedding

- Provider: OpenAI text-embedding-3-small. It was the simplest option with strong Go support, good API ergonomics, and enough free credit to cover the build and validation loop.
- Batch per file rather than one giant corpus-wide embedding call. This isolates failures to a specific file and respects request-size constraints more naturally.
- The project defines an `Embedder` interface plus a fake implementation so the rest of the pipeline can be tested offline without incurring API cost or network dependency.

## Vector storage

- PostgreSQL with pgvector, rather than a hosted vector database. The project intentionally prefers a stateful database it already has to provision, which keeps storage costs under control and avoids vendor lock-in.
- Cosine distance is the chosen operator. OpenAI embeddings are unit-normalized, so cosine, L2, and inner product rank identically; I chose cosine because it stays magnitude-invariant if that ever drifts.
- No vector index at MVP scale. A few hundred chunks are cheap to scan sequentially, and an HNSW or IVFFlat index would be premature optimization before the dataset grows enough to justify it.
- Composite primary key on `(source, chunk_index)` with `ON CONFLICT DO UPDATE`. This makes re-ingestion idempotent and prevents duplicates when the same corpus file is processed again.
- `pgxpool` with a connect ping. This gives the project an actual connection pool for later concurrency and ensures any dead database is caught at startup rather than failing mid-request.

## Migrations

- Goose is used in a dedicated `migrations/` directory. This makes schema changes versioned, reproducible, and easy to apply from scratch without embedding SQL directly into Go code.

## Retrieval

- The pipeline returns a `ScoredChunk` object that includes the similarity score. That score is required later for threshold-based refusal logic, so surfacing it early prevented a later refactor.
- Top-k retrieval is kept in the 3-5 range. More results create noise and extra token cost, while fewer results risk missing the answer.
- Queries are embedded with the same model used for the corpus so they are compared in the same vector space.

## Generation

- Refusal threshold is approximately 0.6, chosen from actual measurements rather than a guessed constant. The relevant query scores were separated from off-topic ones by a meaningful margin, and this threshold captured that gap.
- Prompting is strict: the answer must be grounded only in the provided context. For a postmortem tool, making up a failure mode is more dangerous than admitting uncertainty.
- Citations are generated programmatically from retrieved chunk sources rather than trusting the model to provide them. This ensures the cited sources are real and deduplicated.
- A second refusal mechanism is included in the prompt so questions that are on-topic but not actually supported by the corpus still get a clean “not covered” answer.

## Corpus and ingest

- The source corpus is the `icco/postmortems` set, roughly thirty markdown files. The project intentionally picked a corpus where each file is a self-contained postmortem and therefore straightforward to ingest.
- Parsing is intentionally minimal: strip frontmatter and ingest plain text, without full markdown parsing. That keeps the ingestion path fast and avoids spending time on a rich parser that adds little operational value for RAG.

## GraphQL API

- gqlgen is used for the live GraphQL server because it serves runtime queries at execution time, which is distinct from the genql code-generation work in DPI. That contrast was the important phase decision: gqlgen is the runtime serving layer, while genql is a different codegen workflow entirely.
- The schema exposes a typed boolean `refused`, a list of citation strings, and an optional `k` input argument. This makes client code simpler and avoids parsing refusal semantics out of plain text.

## Docker

- The app uses a multi-stage Docker build that compiles a static Go binary with `CGO_ENABLED=0`. This produces a smaller, predictable runtime image and avoids glibc or musl mismatch problems.
- The runtime base is Alpine during development, with awareness that distroless would be the more production-focused final answer. Alpine is useful for debugging and local cluster use.
- `.dockerignore` excludes `.env` so secret values such as API keys are not baked into image layers.

## Kubernetes

- The app reaches Postgres via the Kubernetes Service name rather than `localhost`. This is one of the key cluster-specific facts: inside a pod, `localhost` refers to the pod itself, not the database service.
- Secrets are injected via environment variables and Kubernetes Secret resources rather than being embedded in the image.
- A readiness probe is used so Kubernetes does not route traffic before the app is actually serving requests, avoiding the common start-order race.
- Local development uses `kind load` with `imagePullPolicy: IfNotPresent` so the locally built image is available in the cluster without requiring a registry.

## Terraform

- The project chooses a local-kind-first Terraform setup instead of going directly to a cloud stack. This teaches the real infrastructure model without the cost and complexity of cloud-specific setup, while still matching production-style provisioning patterns.
- The Terraform layout is split into provider, variables, main, and outputs files to keep each concern separated and easier to scale.
- Secret variables are marked `sensitive = true` to keep API keys and passwords out of logs and plan output.
- Resource references are used for implicit dependencies rather than literal string references so Terraform can order creation correctly without manual sequencing.
- The PVC uses `wait_until_bound = false` because binding can deadlock unless the pod that needs the volume is created in the right order; Terraform's default wait is not always compatible with that flow.

## Meta / process decisions

- The project was built with reviewer-seat AI collaboration: write the core design decisions yourself, delegate boilerplate, and review everything before accepting it. The judgment and tradeoff analysis are the key skills, not the mechanical code generation.
- The migration job was cut when it hit the “image has no goose binary” wall. That was a deliberate decision to stop polishing an unimportant path instead of continuing to invest in a detour.
- Fail-first testing was used deliberately. A test that never fails is a test that may not actually prove the behavior.

## How to use this for communication

For the README, the most useful decisions to pull are the strongest storage, generation, and GraphQL points: the vector-store choice, the refusal threshold, and the typed API contract.
For interviews, the two highest-value points to memorize cold are the cosine-distance rationale and the Kubernetes Service-name behavior; both are easy to explain and demonstrate real understanding of the system.
