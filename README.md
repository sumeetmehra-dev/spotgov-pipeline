# SpotGov Pipeline

AI-powered procurement tender ingestion and matching platform. Ingests public tenders from the EU's TED database and Portugal's dados.gov.pt, indexes them with Elasticsearch (Portuguese + English analyzers), generates vector embeddings for semantic matching, and scores them against company profiles.

<img width="1920" height="961" alt="Screenshot 2026-03-26 at 6 11 54 PM" src="https://github.com/user-attachments/assets/7c4219ea-561c-4f78-bd4f-d65772eb54d7" />

<img width="1920" height="960" alt="Screenshot 2026-03-26 at 6 12 54 PM" src="https://github.com/user-attachments/assets/d99d9bdd-1a99-4e86-94f3-f6ce2ee05d55" />

<img width="1920" height="962" alt="Screenshot 2026-03-26 at 6 13 25 PM" src="https://github.com/user-attachments/assets/2a44edde-fe55-414c-a703-7e771e642c0e" />


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
| `GET` | `/api/v1/tenders/search` | Text search |
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
| Frontend | Next.js 14 + Tailwind | Minimal, server-rendered dashboard |
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
