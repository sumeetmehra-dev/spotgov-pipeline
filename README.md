# Procura

European public procurement data is fragmented across dozens of national portals and the EU's TED database, each with its own schema, language, and query interface. Companies looking for relevant contracts either pay for aggregator subscriptions or spend hours manually scanning multiple sites. Most miss tenders they'd be qualified for because keyword search can't bridge the gap between how a company describes itself and how a buyer writes a contract notice.

Procura pulls tenders from TED and Portugal's dados.gov.pt, normalizes them into a unified schema, indexes them with Elasticsearch (Portuguese + English analyzers), generates Mistral AI embeddings for semantic matching, and scores them against company profiles using a composite ranking that combines vector similarity, BM25 text relevance, CPV code overlap, and contract value fit.

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
  embedding/         — OpenAI client + pgvector store
  matching/          — Scoring algorithm + matcher
  handler/           — HTTP handlers (tender, company, match, health)
  server/            — Chi router setup
frontend/            — Next.js dashboard
```
