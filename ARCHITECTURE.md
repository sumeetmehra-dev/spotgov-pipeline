# Architecture

SpotGov Pipeline is a Go service that ingests public procurement tenders from European data sources, normalizes them into a unified schema, enables full-text and semantic search, and matches tenders against company profiles using a hybrid scoring algorithm.

## System Overview

```
                    ┌─────────────────────────────────────────────────┐
                    │                  Ingestion Layer                 │
                    │                                                  │
  [TED API] ──────▶│  ted.Client ──▶ ted.Normalize ──┐               │
                    │                                  ├──▶ Orchestrator ──▶ Upsert DB
  [dados.gov.pt] ─▶│  dados.Client ──▶ dados.Normalize┘               │
                    │                                                  │
                    └──────────────────────────┬──────────────────────┘
                                               │
                                               ▼
                    ┌──────────────────────────────────────────────────┐
                    │                  Storage Layer                    │
                    │                                                   │
                    │  PostgreSQL (pgvector)     Elasticsearch 8.x     │
                    │  ├── tenders              ├── tenders index      │
                    │  ├── companies            │   (PT + EN analyzer) │
                    │  ├── matches              │                      │
                    │  └── vector embeddings    └── BM25 scoring       │
                    │                                                   │
                    └──────────────────────────┬───────────────────────┘
                                               │
                                               ▼
                    ┌──────────────────────────────────────────────────┐
                    │                Matching Engine                    │
                    │                                                   │
                    │  1. Generate company embedding from profile       │
                    │  2. pgvector cosine similarity → top-K tenders   │
                    │  3. CPV code Jaccard overlap                     │
                    │  4. Value range fit (exponential decay)          │
                    │  5. Composite score:                             │
                    │     0.4×vector + 0.3×BM25 + 0.2×CPV + 0.1×value│
                    │                                                   │
                    └──────────────────────────┬───────────────────────┘
                                               │
                                               ▼
                    ┌──────────────────────────────────────────────────┐
                    │              HTTP API (Chi Router)                │
                    │                                                   │
                    │  /api/v1/tenders     — CRUD, search, ingestion   │
                    │  /api/v1/companies   — Company profile CRUD      │
                    │  /api/v1/match       — Trigger matching          │
                    │  /api/v1/health      — Liveness + readiness      │
                    │                                                   │
                    └──────────────────────────┬───────────────────────┘
                                               │
                                               ▼
                    ┌──────────────────────────────────────────────────┐
                    │           Frontend (Next.js + Tailwind)          │
                    │                                                   │
                    │  /              — Tender dashboard + search       │
                    │  /tenders/:id  — Tender detail view              │
                    │  /profile      — Company profile + match results │
                    │                                                   │
                    └──────────────────────────────────────────────────┘
```

## Data Flow

1. **Ingestion**: The orchestrator launches goroutines (via `errgroup`) for each data source. TED and dados.gov.pt clients fetch concurrently. Each source implements the `Source` interface (`Name()` + `Fetch()`), so new sources (e.g., BASE, OpenTender) can be added without modifying the orchestrator.

2. **Normalization**: Source-specific normalizers map API responses to the unified `model.Tender` struct. Original API responses are preserved in the `raw_data` JSONB column for debugging and reprocessing.

3. **Storage**: Tenders are bulk-upserted using `ON CONFLICT (source_id)` to handle deduplication. The same tenders are indexed into Elasticsearch with Portuguese and English text analyzers for multi-language search.

4. **Embedding**: Mistral AI `mistral-embed` generates 1024-dimensional vectors from tender title + description + buyer + CPV codes. Vectors are stored directly on the tender row using pgvector. Mistral's free tier requires no credit card, keeping the entire stack zero-cost to run.

5. **Matching**: When triggered for a company, the matcher generates an embedding from the company profile, queries pgvector for the top-100 similar tenders by cosine distance, then computes a composite score incorporating CPV overlap (Jaccard index) and value fit (exponential decay outside the company's preferred range).

## Data Model

```
┌────────────────┐       ┌────────────────┐       ┌────────────────┐
│    tenders     │       │    matches     │       │   companies    │
├────────────────┤       ├────────────────┤       ├────────────────┤
│ id (UUID PK)   │◄──────│ tender_id (FK) │       │ id (UUID PK)   │
│ source         │       │ company_id (FK)│──────▶│ name           │
│ source_id (UQ) │       │ score          │       │ description    │
│ title          │       │ vector_score   │       │ cpv_codes[]    │
│ description    │       │ bm25_score     │       │ keywords[]     │
│ cpv_codes[]    │       │ cpv_overlap    │       │ countries[]    │
│ buyer_name     │       │ matched_at     │       │ min_value      │
│ buyer_country  │       └────────────────┘       │ max_value      │
│ estimated_value│                                 │ embedding      │
│ deadline       │                                 └────────────────┘
│ raw_data       │
│ embedding      │
└────────────────┘
```

## Design Decisions

### GORM over raw SQL
Auto-migration lets the schema evolve without maintaining separate migration files during rapid prototyping. The custom Zap logger integration surfaces slow queries (>1s threshold) without per-query noise, and the struct-tag-based model definitions keep the schema colocated with the Go types that use it. Considered raw `database/sql` with pgx for performance, but the query complexity here doesn't justify the boilerplate — there are no complex joins or CTEs where an ORM gets in the way, and GORM's `Clauses(OnConflict{})` made the bulk upsert logic clean.

### Chi over Gin/Echo
Handlers are plain `http.HandlerFunc` — no framework-specific context wrappers, no hidden magic. Middleware is composable via `r.Use()` and `r.With()`, so the request pipeline is explicit. Gin was considered but its custom `gin.Context` creates lock-in: every handler signature is framework-specific, and moving away from it later means rewriting every endpoint. Echo has the same problem. Chi's middleware ecosystem (request ID, real IP, logging) is sufficient, and the zero-dependency approach means the binary stays lean.

### pgvector over Qdrant/Pinecone
Keeping vectors in PostgreSQL avoids adding another stateful service to the stack. At the current scale (tens of thousands of tenders, not millions), pgvector's IVFFlat index performs well — the approximate nearest neighbor overhead is negligible compared to the network round-trip a dedicated vector DB would add. Qdrant was considered for its filtering capabilities, but pgvector's ability to combine vector search with SQL `WHERE` clauses in a single query (e.g., filter by country + CPV before ranking) eliminates the two-phase filter-then-rank pattern. If the dataset grows past ~1M vectors, migrating to HNSW indexing or an external vector store becomes worthwhile.

### pgvector on tender rows (not a separate table)
There is a strict 1:1 relationship between tenders and their embeddings. A separate table would add JOIN overhead on every similarity query with no schema benefit. Considered a separate `tender_embeddings` table for cleaner separation, but profiling showed the JOIN adds ~2ms per query at scale with no upside — the embedding is never accessed without its tender.

### Source interface for extensibility
```go
type Source interface {
    Name() string
    Fetch(ctx context.Context, since time.Time) ([]model.Tender, error)
}
```
Adding BASE, OpenTender, or any OCDS-compliant source is a single struct implementing two methods. The orchestrator discovers nothing about source internals — it just calls `Fetch()` and upserts the results.

### Elasticsearch over Typesense/Meilisearch
TED publishes ~700K+ notices per year across all EU languages. Portuguese procurement text requires proper stemming and stop-word handling — Elasticsearch's built-in `portuguese` analyzer is production-grade and battle-tested, while Typesense's language support is limited to basic tokenization. Meilisearch was considered for its simpler setup, but it lacks multi-field analyzers (we need separate Portuguese and English analysis on the same field) and its sharding story is immature for datasets that grow year-over-year. Elasticsearch's mature bulk API and BM25 scoring also make it the natural choice for the text-relevance component of hybrid search.

### Composite scoring with weighted components
```
score = 0.4 × vector_similarity + 0.3 × bm25_score + 0.2 × cpv_overlap + 0.1 × value_fit
```
Vector similarity captures semantic relevance. CPV overlap provides exact classification matching. Value fit ensures tenders are within the company's operational scale. Weights are configurable and can be tuned per customer. Considered a learned-to-rank model, but the training data doesn't exist yet — a weighted linear combination is interpretable, debuggable, and good enough to validate the product hypothesis before investing in ML infrastructure.

### Reciprocal Rank Fusion for hybrid search
Combines BM25 (keyword relevance) and vector (semantic relevance) rankings without requiring normalized scores:
```
score = 1/(k + rank_bm25) + 1/(k + rank_vector)    where k = 60
```
k=60 follows the standard from Cormack, Clarke, and Buettcher's original reciprocal rank fusion paper, which showed this value performs robustly across diverse ranking pair combinations without task-specific tuning. The alternative was score-level fusion (normalize BM25 and cosine scores to the same scale, then combine), but BM25 scores are unbounded and dataset-dependent — normalization requires knowing the score distribution, which changes with every ingestion. RRF sidesteps this entirely by operating on ranks, not scores.

### Structured logging with Zap
Environment-aware: development mode gives colored console output with caller info; production mode emits JSON for log aggregation. Zero-allocation in the hot path. Considered `slog` (stdlib, Go 1.21+) for fewer dependencies, but Zap's `Named()` sub-loggers and `DPanic` level make it easier to trace logs across packages in a multi-source ingestion pipeline.

## Concurrency Model

The ingestion orchestrator uses `golang.org/x/sync/errgroup` for structured concurrency:
- Each data source runs in its own goroutine — in practice this means 2 goroutines today (TED, dados), scaling linearly as sources are added. The goroutine count is bounded by the number of registered sources, not by data volume.
- Results are collected into a shared `[]model.Tender` slice protected by a `sync.Mutex`. The mutex is only held during the append (microseconds), not during the HTTP fetch, so sources never block each other.
- Individual source failures are logged as errors but return `nil` from the goroutine — `errgroup.Wait()` only propagates errors that would indicate a systemic problem, not a single flaky API.
- The TED client enforces a 200ms sleep between paginated requests to avoid hammering the API. This is a cooperative rate limit, not a token bucket — appropriate for a single-instance service hitting a public API, where the goal is politeness rather than throughput maximization.
- The shared `errgroup` context enables cancellation propagation: if the parent context is cancelled (e.g., server shutdown), in-flight HTTP requests are abandoned via `http.NewRequestWithContext`.

## Error Handling and Resilience

The pipeline is designed to degrade gracefully rather than fail atomically:

- **Source failures are isolated.** The orchestrator wraps each source in its own goroutine. If the TED API is down or times out, dados.gov.pt ingestion still completes, and vice versa. Failed sources are logged as errors but don't propagate — the pipeline upserts whatever it successfully fetched. This means a partial ingestion (e.g., 200 TED notices but 0 from dados) is always better than no ingestion.

- **Database connection retries with exponential backoff.** On startup, PostgreSQL connection attempts retry 3 times with 2s/4s/8s backoff. This handles the common Docker Compose race where the API container starts before Postgres finishes initialization. After startup, GORM's connection pool handles transient disconnects transparently.

- **Embedding failures don't block ingestion.** The OpenAI API key is optional — if missing, the server starts with matching features disabled and logs a warning. During batch embedding generation, individual failures are logged and skipped; the remaining tenders still get their embeddings. Tenders without embeddings are tracked via `WHERE embedding IS NULL` and can be retried on the next call.

- **Elasticsearch unavailability doesn't break reads.** The tender list and detail endpoints query PostgreSQL directly. Elasticsearch is only used for the `/tenders/search` endpoint. If ES is temporarily unreachable, search returns an error but CRUD operations continue normally.

- **Upsert-based deduplication is idempotent.** Re-running ingestion against the same date range is safe — `ON CONFLICT (source_id)` updates existing rows instead of creating duplicates. This means recovery from a partial failure is just "run it again."

- **Observability.** Every request carries a unique ID via Chi's `middleware.RequestID`, propagated through the context. Zap's `Named()` sub-loggers tag each log line with the originating package (`ted`, `dados`, `orchestrator`, `matcher`), so a single ingestion run can be traced end-to-end by filtering on the request ID. The GORM logger flags slow queries (>1s) as warnings, giving immediate visibility into database bottlenecks without enabling full query logging.

## Future Considerations

- **Scheduled ingestion.** Currently triggered manually via API. A cron-based background goroutine (or external scheduler) would run ingestion every 6 hours, ensuring new TED notices are captured within the same day they're published. This is the single highest-impact improvement for making the system useful without human intervention.

- **Event-driven indexing via message queue.** Right now, DB upsert → ES indexing → embedding generation happens synchronously in the ingestion path. Decoupling these with a message queue (Kafka, NATS) would let ingestion complete faster and let indexing/embedding retry independently on failure — important when OpenAI rate limits hit during a large batch.

- **Webhook notifications.** Companies that configure a profile want to know when a high-scoring tender appears, not poll the dashboard. A notification system that fires when a new match exceeds a score threshold would close the loop between ingestion and user action.
