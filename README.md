# Procura

European public procurement data is scattered across dozens of national portals and the EU's TED database. Different schemas, different languages, different query interfaces. Companies hunting for contracts either pay for aggregator subscriptions or burn hours scanning sites manually. Most miss tenders they'd be perfectly qualified for, because keyword search can't bridge the gap between how a company describes itself and how a government buyer writes a contract notice.

Procura pulls tenders from TED and Portugal's dados.gov.pt, normalizes them into a single schema, and indexes them with Elasticsearch (Portuguese + English analyzers). It then generates Mistral AI embeddings and scores tenders against company profiles, ranking by a mix of vector similarity, BM25 text relevance, CPV code overlap, and contract value fit.

<img width="1920" height="958" alt="Screenshot 2026-03-31 at 11 41 57 AM" src="https://github.com/user-attachments/assets/7bd38fa2-efc6-4a85-aa1e-783ddaa40b2f" />

<img width="1920" height="954" alt="Screenshot 2026-03-31 at 11 50 08 AM" src="https://github.com/user-attachments/assets/3f4033b1-a630-4288-9c57-78d6093dd58d" />

<img width="1920" height="955" alt="Screenshot 2026-03-31 at 11 49 15 AM" src="https://github.com/user-attachments/assets/0304d5ce-1fcf-4b8d-95bd-b06a1d2fb95b" />


## Quick Start

```bash
# 1. Clone and configure
cp .env.example .env
# Set your MISTRAL_API_KEY in .env (free tier, no credit card — https://console.mistral.ai/)

# 2. Start all services
docker compose up -d

# 3. Trigger ingestion of Portuguese tenders
curl -X POST http://localhost:8080/api/v1/tenders/ingest

# 4. Browse tenders
open http://localhost:3000
```

## Architecture

See [ARCHITECTURE.md](ARCHITECTURE.md) for detailed system design.

```
[TED API] ──┐                         ┌── [Elasticsearch]
             ├── [Ingestion Service] ──┤
[dados.gov] ─┘       │                └── [PostgreSQL + pgvector]
                      │                          │
              [Normalizer]                       ▼
                                        [Matching Engine]
                                               │
[Frontend] ←── [Chi HTTP API] ←───────── [Match Results]
```

## API

| Method | Endpoint | Description |
|--------|----------|-------------|
| `GET` | `/api/v1/health` | Health check |
| `GET` | `/api/v1/stats` | Ingestion statistics |
| `GET` | `/api/v1/tenders` | List tenders (paginated, filterable) |
| `GET` | `/api/v1/tenders/:id` | Tender detail |
| `GET` | `/api/v1/tenders/search` | Hybrid search (ES with DB fallback) |
| `POST` | `/api/v1/tenders/ingest` | Trigger ingestion |
| `POST` | `/api/v1/companies` | Create company profile |
| `GET` | `/api/v1/companies/:id` | Get company |
| `PUT` | `/api/v1/companies/:id` | Update company |
| `GET` | `/api/v1/companies/:id/matches` | Get matched tenders |
| `POST` | `/api/v1/match` | Trigger matching |
| `POST` | `/api/v1/match/embeddings` | Generate tender embeddings |

### Query Parameters

**GET /api/v1/tenders**
- `page`, `page_size` — Pagination (default: 1, 20)
- `cpv` — Filter by CPV code
- `country` — Filter by buyer country (e.g., `PT`)
- `min_value`, `max_value` — Filter by estimated value range
- `deadline_after` — Filter by deadline (format: `2024-06-01`)
- `q` — Text search on title and description

### Match Result Example

`GET /api/v1/companies/:id/matches` returns scored tenders with individual signal breakdowns:

```json
{
  "matches": [
    {
      "tender": {
        "id": "b85617ef-a411-42fd-8014-0b8bb7104a9b",
        "title": "Serviços de manutenção de edifícios — Lisboa",
        "buyer_name": "Câmara Municipal de Lisboa",
        "estimated_value": 280000,
        "cpv_codes": ["50700000"],
        "deadline": "2026-04-28T00:00:00Z"
      },
      "score": 0.79,
      "vector_score": 0.91,
      "bm25_score": 0.74,
      "cpv_overlap": 0.50,
      "value_fit": 0.88
    }
  ]
}
```

Each signal is exposed so you can see *why* a tender ranked where it did — not just that it did. Vector similarity (0.91) shows strong semantic overlap between the company profile and tender description, while CPV overlap (0.50) is lower because maintenance (`50700000`) and construction (`45000000`) share a division but not a group.

## Tech Stack

| Component | Technology | Why |
|-----------|-----------|-----|
| HTTP Router | Chi | Composable middleware, net/http compatible |
| ORM | GORM | Auto-migration, custom Zap logger |
| Database | PostgreSQL + pgvector | Relational storage + vector similarity search |
| Search | Elasticsearch 8.x | BM25 scoring, Portuguese language analysis |
| Embeddings | Mistral AI mistral-embed | 1024-dim vectors, free tier, no credit card |
| Logging | Zap | Structured, zero-allocation, env-aware |
| Frontend | Next.js 14 + Tailwind | Modern dashboard with search, detail views, company matching |
| Infra | Docker Compose | One-command development setup |

## Development

```bash
# Run Go backend locally (requires PostgreSQL + ES running)
make dev

# Run tests
make test

# Run with coverage
make test-cover

# Lint
make lint

# Build binary
make build
```

## Project Structure

```
cmd/server/          — Application entry point
internal/
  config/            — Environment-based configuration
  database/          — PostgreSQL connection, migrations
  model/             — GORM models (Tender, Company, Match)
  ingestion/         — Data source interface + orchestrator
    ted/             — TED API client + normalizer
    dados/           — dados.gov.pt client + normalizer
  search/            — Elasticsearch client + indexer
  embedding/         — Mistral AI client + pgvector store
  matching/          — Scoring algorithm + matcher
  handler/           — HTTP handlers (tender, company, match, health)
  server/            — Chi router setup
frontend/            — Next.js dashboard
```

## What's Not Here Yet

Ingestion is manual right now. You hit an API endpoint to trigger it. A background goroutine on a 6-hour interval (or just cron) is the obvious next step.

DB upsert, ES indexing, and embedding generation all happen synchronously. Works fine at current scale, but a message queue would let each step fail and retry on its own. Matters most when Mistral rate-limits you mid-batch.

There's no notification system. Companies have to check the dashboard to see new matches, which defeats the purpose. A webhook that fires when a match crosses a score threshold would close that gap.

The data model has `tenant_id` on companies and matches, but the API doesn't enforce scoping yet. It's like the plumbing is there but the access control isn't.

More detail in [ARCHITECTURE.md](ARCHITECTURE.md#what-id-do-next).
