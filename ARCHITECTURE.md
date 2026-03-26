# Architecture

A Go service that pulls public procurement tenders from EU data sources, normalizes them, indexes them for full-text and semantic search, and ranks them against company profiles. The short version: tenders go in, scored matches come out.

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

**Ingestion.** The orchestrator spins up a goroutine per data source via `errgroup`. TED and dados.gov.pt fetch concurrently. Both implement the `Source` interface (`Name()` + `Fetch()`), so adding a new source (BASE, OpenTender, whatever) doesn't touch the orchestrator at all.

**Normalization.** Each source has its own normalizer that maps raw API responses into the shared `model.Tender` struct. The original JSON is kept in a `raw_data` JSONB column — useful for debugging when the mapping is wrong but the original data is fine.

The TED normalizer handles i18n text fields (preferring English, then Portuguese, then whatever's available) and has some fiddly logic around descriptions. TED has both `description-proc` (the overall procurement description) and `description-lot` (per-lot labels), and a lot of the lot descriptions are just "Lote 1" or similar noise. The normalizer prefers proc descriptions, filters out bare lot labels with regex, deduplicates against the title, and combines both fields when the lot description actually adds something the proc description doesn't say.

**Storage.** Tenders get bulk-upserted with `ON CONFLICT (source_id)` for deduplication. Same batch gets indexed into Elasticsearch with both Portuguese and English text analyzers, since procurement notices on TED can be in either language (or both).

**Embedding.** Mistral AI `mistral-embed` generates 1024-dimensional vectors from a concatenation of the tender's title, description, buyer name, and CPV codes. Stored directly on the tender row via pgvector. Mistral's free tier means the whole stack runs at zero cost without a credit card, which matters for a demo project that someone needs to actually clone and run.

**Matching.** When you trigger matching for a company, the system generates an embedding from the company profile text, pulls the top-100 most similar tenders by cosine distance from pgvector, then layers on CPV overlap (Jaccard) and value fit (exponential decay outside the company's preferred range) to produce a final composite score.

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

Three tables. `matches` is the join between tenders and companies, with the individual score components broken out so you can debug why a particular tender ranked where it did.

## Design Decisions

### GORM over raw SQL

I went with GORM mainly for auto-migration. During early development the schema changed constantly, and maintaining separate migration files for a project that doesn't have a stable schema yet is busywork. The struct-tag-based model definitions mean the Go type *is* the schema, which keeps things in one place.

I did consider raw `database/sql` with pgx. It would be faster, sure. But the queries here are simple: there are no complex joins, no CTEs, nothing where GORM's abstraction gets in the way. And `Clauses(OnConflict{})` gave me clean bulk upsert logic without hand-rolling the SQL. The custom Zap logger integration catches slow queries over 1s, which is enough observability without logging every `SELECT`. GORM is confined to the repository layer — handlers and business logic work with plain Go structs — so if raw pgx ever becomes necessary for performance, the swap happens in one package without touching a single handler.

### Chi over Gin/Echo

Handlers are plain `http.HandlerFunc`. No custom context wrappers, nothing framework-specific leaking into business logic. Middleware composes with `r.Use()` and `r.With()`, so you can read the router setup and know exactly what runs on each route.

I looked at Gin first. The problem is `gin.Context` — once you're writing handlers that take `*gin.Context`, every handler in the codebase is coupled to Gin. If you ever want to move off it, you're rewriting every endpoint. Echo has the same issue. Chi avoids this entirely because it's just `net/http` underneath, and its middleware (request ID, real IP, logging) covers what I actually needed.

### pgvector over Qdrant/Pinecone

Another service in the stack means another thing to monitor, configure, and debug. At the scale this project operates at (tens of thousands of tenders, not millions), pgvector's IVFFlat index handles similarity search without breaking a sweat. The latency overhead of ANN is negligible compared to the network round-trip you'd pay talking to a separate vector DB.

Qdrant has better filtering, arguably. But pgvector lets you combine vector search with SQL `WHERE` clauses in one query. Filter by country and CPV code, *then* rank by similarity, all in a single round-trip. With a dedicated vector DB you'd need to filter first, fetch IDs, then join back to Postgres. Two hops instead of one.

If the dataset grows past about a million vectors, switching to HNSW indexing or an external store would make sense. Not a concern right now.

### Embeddings on the tender row, not a separate table

Tenders and embeddings are 1:1. Always. The embedding is never loaded without the tender it belongs to, and there's no use case where you'd want the embedding schema to diverge. I initially considered a `tender_embeddings` table for tidiness, but that JOIN adds about 2ms to every similarity query and buys you nothing. Keeping it on the same row means one index scan, one row fetch.

### Source interface

```go
type Source interface {
    Name() string
    Fetch(ctx context.Context, since time.Time) ([]model.Tender, error)
}
```

Two methods. That's the entire contract for adding a new data source. The orchestrator doesn't know or care how a source gets its data — HTTP, scraping, CSV, whatever. It calls `Fetch()`, gets back tenders, upserts them. If you wanted to add BASE or OpenTender or any OCDS-compliant source, you'd write a struct with these two methods and register it. Nothing else changes.

### Elasticsearch over Typesense/Meilisearch

This one came down to language support. Portuguese procurement text needs real stemming and stop-word handling, not just tokenization. Elasticsearch's built-in `portuguese` analyzer handles this well; Typesense's language support is noticeably thinner.

Meilisearch was tempting for its simpler setup. But it can't run separate analyzers on the same field (I need Portuguese *and* English analysis, since TED notices come in both), and its sharding is still immature for a dataset that grows every year. TED alone publishes over 700K notices annually. Elasticsearch's bulk API and BM25 scoring also fit naturally into the hybrid search design.

### Composite scoring

```
score = 0.4 × vector_similarity + 0.3 × bm25_score + 0.2 × cpv_overlap + 0.1 × value_fit
```

Four signals, weighted by how much I trust each one. Vector similarity gets the most weight because semantic matching catches tenders that keyword search would miss entirely (a "building maintenance" company matching against a tender for "facility upkeep services"). CPV overlap is a hard classification signal. Value fit penalizes tenders that are orders of magnitude outside a company's operational range.

The weights are configurable. A learned-to-rank model would be the right long-term answer, but there's no training data yet. This gets the product to a testable state where you can actually see if the matching makes sense, then collect feedback to train something better.

**Worked example.** A construction company with CPV codes `[45000000, 45210000]` and a preferred contract range of €100K–€500K. A tender titled *"Serviços de manutenção de edifícios — Lisboa"* (CPV `50700000`, value €280K) scores:

| Signal | Value | Score |
|--------|-------|-------|
| Vector similarity | 0.91 cosine | 0.91 |
| BM25 | index-relative | 0.74 |
| CPV Jaccard | shares division, not group | 0.50 |
| Value fit | €280K within range | 0.88 |
| **Composite** | | **0.79** |

CPV scores lower because `50700000` (maintenance) and `45000000` (construction) share a division but not a group — intentional. The tender surfaces as relevant but not a perfect match.

### Reciprocal Rank Fusion for hybrid search

```
score = 1/(k + rank_bm25) + 1/(k + rank_vector)    where k = 60
```

k=60 comes from Cormack, Clarke, and Buettcher's original RRF paper. It works well across different ranking combinations without tuning, which matters here since I don't have relevance judgments to tune against.

The other option was score-level fusion: normalize BM25 and cosine scores to the same scale, then combine. The problem is BM25 scores are unbounded and change depending on what's in the index. You'd need to know the score distribution, and that shifts every time you ingest new data. RRF avoids this by working with ranks instead of raw scores.

### Zap over slog

Both would work here. I went with Zap because its `Named()` sub-loggers make it easy to tag every log line with the originating package — when you're debugging an ingestion run that touches `ted`, `dados`, `orchestrator`, and `matcher`, being able to filter by component saves real time. `DPanic` is also useful during development: panic in dev, log in prod. `slog` (Go 1.21+) would save a dependency, but structured sub-loggers require more manual wiring.

### Next.js for the frontend

I chose Next.js 14 with the App Router over a Vite + React SPA. The routing conventions map cleanly to the three views I needed (dashboard, tender detail, company profile), and server-rendering the initial page load means no loading skeleton before you see actual data. For a dashboard that's mostly displaying server-side results, SSR is the simpler model.

The UI is Tailwind. I spent some time on the match results page specifically — each match shows a score breakdown (vector, BM25, CPV, value fit) so you can see *why* a tender ranked where it did, not just that it did. Tenders show deadline urgency and source badges (TED vs dados) since those matter for prioritization. All API calls are client-side (`'use client'` components), so the frontend could sit on Vercel while the backend runs elsewhere. That said, this is a backend-heavy project and the frontend reflects that. I put my time into the pipeline and matching engine.

## Concurrency Model

The orchestrator uses `golang.org/x/sync/errgroup` for structured concurrency. In practice, this means 2 goroutines today (TED + dados), scaling linearly as sources are added. The goroutine count is bounded by the number of registered sources, not by data volume, so this doesn't explode.

Results go into a shared `[]model.Tender` slice protected by `sync.Mutex`. The lock is only held during the append — microseconds — not during the HTTP fetch itself, so sources never actually block each other.

Source failures return `nil` from the goroutine rather than an error. `errgroup.Wait()` only surfaces errors that indicate something systemic, not one flaky API returning 503. If TED is down, you still get dados data.

The TED client sleeps 200ms between paginated requests. This is a cooperative rate limit, not a token bucket. For a single-instance service hitting a public API, the goal is politeness, not maximizing throughput. Nobody wants their IP blocked because the ingestion service hammered TED's endpoint.

The shared `errgroup` context handles shutdown: if the parent context is cancelled, in-flight HTTP requests get abandoned via `http.NewRequestWithContext`. No orphaned goroutines.

## Error Handling and Resilience

The general principle: partial success is better than total failure. If one thing breaks, the rest should keep working.

**Source failures are isolated.** If TED is down, dados.gov.pt still runs. The pipeline upserts whatever it got. 200 TED tenders and 0 from dados is still 200 tenders more than if the whole run had crashed.

**Database retries on startup.** PostgreSQL connection retries 3 times with 2s/4s/8s exponential backoff. This mostly exists because Docker Compose starts all containers simultaneously, and the Go service boots faster than Postgres finishes initializing. After startup, GORM's connection pool handles transient issues on its own.

**Embeddings are optional.** No Mistral API key? The server starts anyway with matching disabled and a log warning. During batch embedding, if one tender fails, it gets skipped — the rest still go through. Failed ones show up via `WHERE embedding IS NULL` for the next run to pick up. Didn't want a rate limit at tender #47 to waste the work already done on #1 through #46.

**Elasticsearch going down doesn't break the app.** Tender list and detail endpoints read from PostgreSQL. ES is only needed for `/tenders/search`. If it's unreachable, search returns an error but everything else keeps working.

**Ingestion is idempotent.** Running the same ingestion twice doesn't create duplicates. `ON CONFLICT (source_id)` updates the existing row. Recovery from a failed run is literally just running it again.

**Observability.** Every request gets a unique ID via Chi's `middleware.RequestID`. Zap sub-loggers tag lines with the originating package. A single ingestion run can be traced end-to-end by filtering on request ID across `ted`, `dados`, `orchestrator`, and `matcher` logs. The GORM logger flags queries over 1s as warnings — enough to spot bottlenecks without drowning in query logs.

## What I'd Do Next

**Scheduled ingestion** is the biggest gap. Right now you trigger ingestion via API call. A background goroutine on a 6-hour interval (or just cron) would keep the data fresh without anyone having to remember to hit the endpoint.

**Async indexing.** DB upsert, ES indexing, and embedding generation all happen synchronously right now. Decoupling them with a message queue (Kafka or NATS) would speed up the ingestion path and let each downstream step retry independently. This matters most when Mistral rate-limits you halfway through a large batch.

**Notifications.** The whole point of matching is telling a company when something relevant shows up. Polling a dashboard doesn't cut it. A webhook that fires when a match crosses a score threshold is the obvious next step.

**Multi-tenancy.** Right now the system is single-tenant — one tender pool, one match table, no access boundaries. Production would scope company profiles, match history, and ingestion filters to isolated tenants with row-level security in Postgres. The data model supports this with a `tenant_id` column on the `companies` and `matches` tables; the API layer doesn't enforce it yet.
