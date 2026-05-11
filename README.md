# Hybrid Search

A Go microservices project for experimenting with job search using semantic vector search, ParadeDB BM25 lexical search, and a simple Reciprocal Rank Fusion hybrid query.

The system is built from two HTTP services:

- `broker-service`: API gateway exposed on `localhost:6000`
- `search-service`: search API exposed on `localhost:6001`

Supporting services are managed with Docker Compose:

- ParadeDB/PostgreSQL on `localhost:5432`
- Ollama on `localhost:11434`

## Architecture

```text
Client
  |
  | HTTP :6000
  v
Broker Service
  |
  | forwards search requests
  v
Search Service
  |                 |
  | SQL             | embeddings API
  v                 v
ParadeDB        Ollama all-minilm
PostgreSQL
```

The search service stores job records in PostgreSQL, generates 384-dimensional embeddings with Ollama's `all-minilm` model, and queries the same `jobs` table with vector and BM25 indexes.

## Repository Layout

```text
.
|-- broker-service/
|   |-- cmd/api/main.go
|   |-- internal/server/
|   |-- broker-service.dockerfile
|   `-- go.mod
|-- search-service/
|   |-- cmd/api/main.go
|   |-- internal/assets/job_salary_prediction_dataset.csv
|   |-- internal/database/
|   |-- internal/server/
|   |-- search-service.dockerfile
|   `-- go.mod
|-- project/
|   |-- docker-compose.yml
|   `-- Makefile
`-- README.md
```

## Prerequisites

- Docker and Docker Compose
- Make, if you want to use the provided shortcuts
- Go 1.25.0 or newer for local development

## Quick Start

Run the full stack from the `project` directory:

```bash
cd project
make up_build
```

`make up_build` builds both Go services, starts the Docker Compose stack, and pulls the Ollama `all-minilm` model inside the Ollama container.

Check the services:

```bash
curl http://localhost:6000/ping
curl http://localhost:6001/ping
```

Load job data into PostgreSQL:

```bash
curl http://localhost:6001/store-csv
```

The CSV file contains 250,000 data rows plus a header. The current importer loads up to about 10,000 rows and generates an embedding for each imported row, so the first import can take a few minutes.

Stop the stack:

```bash
cd project
make down
```

## Make Commands

Run these from `project/`.

```bash
make up                    # Start all containers
make up_build              # Build service binaries, start containers, pull all-minilm
make down                  # Stop containers
make postgres-up           # Start only PostgreSQL/ParadeDB
make postgres_up_build     # Same as postgres-up
make search_service_start  # Run search-service locally on port 6001
make broker_start          # Run broker-service locally on port 6000
make build_search_service  # Build search-service Linux binary
make build_broker_service  # Build broker-service Linux binary
```

## Configuration

Docker Compose configures the services with these defaults.

Search service:

```bash
PORT=80
DSN="host=postgres user=admin password=admin dbname=advance-search port=5432 sslmode=disable"
OLLAMA_EMBEDDING_URL="http://ollama:11434/api/embeddings"
```

Broker service:

```bash
PORT=80
```

PostgreSQL:

```text
host: localhost
port: 5432
user: admin
password: admin
database: advance-search
```

## API

All responses use a common JSON envelope:

```json
{
  "message": "Ok",
  "data": []
}
```

### Broker Service

Base URL:

```text
http://localhost:6000
```

#### `GET /ping`

Health check.

```bash
curl http://localhost:6000/ping
```

#### `POST /broker`

Routes search requests to the search service. The action names and semantic payload spelling currently use `symantic` because that is the value implemented by the service.

Semantic search:

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "symantic_search",
    "symanticSearchPayload": {
      "query": "machine learning engineer healthcare",
      "page": "0",
      "limit": "10"
    }
  }'
```

Lexical search:

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "lexical_search",
    "lexicalSearchPayload": {
      "query": "data analyst telecom",
      "page": 0,
      "limit": 10
    }
  }'
```

Hybrid search:

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "hybrid_search",
    "hybridSearchPayload": {
      "query": "senior python developer",
      "page": 0,
      "limit": 10
    }
  }'
```

### Search Service

Base URL:

```text
http://localhost:6001
```

#### `GET /ping`

Health check.

```bash
curl http://localhost:6001/ping
```

#### `GET /read-csv`

Reads and parses `search-service/internal/assets/job_salary_prediction_dataset.csv`.

```bash
curl http://localhost:6001/read-csv
```

#### `GET /store-csv`

Imports rows from the CSV into PostgreSQL and generates embeddings. Run this before search endpoints.

```bash
curl http://localhost:6001/store-csv
```

#### `POST /get-vector`

Generates an embedding for text through Ollama.

```bash
curl -X POST http://localhost:6001/get-vector \
  -H "Content-Type: application/json" \
  -d '{"text":"machine learning engineer"}'
```

#### `POST /symantic-search`

Semantic vector search. Pagination is passed as query parameters.

```bash
curl -X POST "http://localhost:6001/symantic-search?page=0&limit=10" \
  -H "Content-Type: application/json" \
  -d '{"query":"machine learning engineer healthcare"}'
```

#### `POST /lexical-search`

BM25 lexical search over title, education level, industry, company size, and location.

```bash
curl -X POST http://localhost:6001/lexical-search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "data analyst telecom",
    "page": 0,
    "limit": 10
  }'
```

#### `POST /hybrid-search`

Hybrid search that combines BM25 and vector rankings with Reciprocal Rank Fusion.

```bash
curl -X POST http://localhost:6001/hybrid-search \
  -H "Content-Type: application/json" \
  -d '{
    "query": "senior python developer",
    "page": 0,
    "limit": 10
  }'
```

## Search Behavior

### Semantic Search

Semantic search converts the query to an embedding using Ollama and compares it with stored job embeddings using pgvector distance ordering.

The current job embedding text is built from:

- job title
- education level
- industry
- company size

### Lexical Search

Lexical search uses ParadeDB's BM25 support through the `pg_search` extension. It searches these fields:

- title
- education level
- industry
- company size
- location

### Hybrid Search

Hybrid search gathers candidates from lexical and semantic searches, assigns rank-based scores with Reciprocal Rank Fusion, and returns the combined ranking.

## Database

The `jobs` table is created through GORM auto-migration. Manual migrations then enable extensions and create indexes.

Model fields:

```text
id
created_at
updated_at
deleted_at
title
experience
education_level
skills_count
industry
company_size
location
remote_work
certifications
salary
embedding vector(384)
```

Manual migration:

```sql
CREATE EXTENSION IF NOT EXISTS vector;

CREATE INDEX IF NOT EXISTS idx_jobs_embedding_hnsw
ON jobs USING hnsw (embedding vector_ip_ops);

CREATE EXTENSION IF NOT EXISTS pg_search;

CREATE INDEX IF NOT EXISTS idx_jobs_search_idx_bm25
ON jobs USING bm25(id, title, education_level, industry, company_size, location)
WITH (key_field='id');
```

## Dataset

The bundled CSV is `search-service/internal/assets/job_salary_prediction_dataset.csv`.

Columns:

- `job_title`
- `experience_years`
- `education_level`
- `skills_count`
- `industry`
- `company_size`
- `location`
- `remote_work`
- `certifications`
- `salary`

## Local Development

Start only the infrastructure:

```bash
cd project
docker-compose up -d postgres ollama
docker exec ollama ollama pull all-minilm
```

Run the search service locally:

```bash
cd search-service
DSN="host=localhost user=admin password=admin dbname=advance-search port=5432 sslmode=disable" \
PORT="6001" \
OLLAMA_EMBEDDING_URL="http://localhost:11434/api/embeddings" \
go run ./cmd/api/
```

Run the broker service locally:

```bash
cd broker-service
PORT="6000" go run ./cmd/api/
```

Build binaries manually:

```bash
cd search-service
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o searchApp ./cmd/api/

cd ../broker-service
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o brokerApp ./cmd/api/
```

## Notes

- The service and payload names currently spell semantic search as `symantic`; use the implemented spelling in API calls.
- `/store-csv` is not idempotent. Calling it repeatedly inserts another batch of rows.
- The search service runs migrations at startup.
- The broker assumes the Docker Compose service name `search-service` when forwarding requests.
- There are currently no test files in the repository.
