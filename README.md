# Advanced Hybrid Search

A microservices-based advanced search system that combines **semantic search** (vector embeddings), **lexical search** (BM25 full-text), and **hybrid search** for intelligent job search functionality. Built with Go, PostgreSQL with pgvector extension, and Ollama for AI-powered embeddings.

## 📋 Overview

This project demonstrates a modern search architecture using:

- **Semantic Search**: Vector similarity search using embeddings for meaning-based queries
- **Lexical Search**: BM25 full-text search for keyword-based queries
- **Hybrid Search**: Combines both approaches for optimal search results
- **Job Dataset**: Pre-populated dataset with job titles, salaries, industries, and more

## 🏗️ Architecture

```
┌─────────────────┐
│  Broker Service │ (Port 6000)
│  API Gateway    │
└────────┬────────┘
         │
         ▼
┌─────────────────┐      ┌──────────────┐      ┌─────────────┐
│ Search Service  │◄────►│  PostgreSQL  │◄────►│   Ollama    │
│  (Port 6001)    │      │  + ParadeDB  │      │  Embeddings │
│                 │      │  (Port 5432) │      │ (Port 11434)│
└─────────────────┘      └──────────────┘      └─────────────┘
```

### Services

1. **Broker Service** (`broker-service/`)
   - Acts as an API gateway and request router
   - Exposes unified endpoints for all search types
   - Routes requests to the search service
   - Runs on port 6000

2. **Search Service** (`search-service/`)
   - Core search functionality
   - Handles semantic, lexical, and hybrid searches
   - Manages job data and embeddings
   - Integrates with PostgreSQL and Ollama
   - Runs on port 6001

3. **PostgreSQL + ParadeDB**
   - Vector database with pgvector extension
   - BM25 full-text search via pg_search extension
   - HNSW index for fast vector similarity search
   - Stores job data with 384-dimensional embeddings

4. **Ollama**
   - Provides embedding generation using `all-minilm` model
   - Converts text queries to vector embeddings
   - Runs locally for privacy and speed

## 🚀 Getting Started

### Prerequisites

- Docker and Docker Compose
- Make (optional, for convenience commands)
- Go 1.25.0+ (for local development)

### Quick Start

1. **Clone the repository**

   ```bash
   git clone <repository-url>
   cd advance-hybrid-search
   ```

2. **Navigate to project directory**

   ```bash
   cd project
   ```

3. **Start all services**

   ```bash
   make up_build
   ```

   This will:
   - Build the broker and search services
   - Start PostgreSQL with ParadeDB extensions
   - Start Ollama for embeddings
   - Create necessary database indexes

4. **Load the job dataset**

   ```bash
   curl http://localhost:6001/store-csv
   ```

   This loads and processes the job salary dataset, generating embeddings for ~10,000 jobs.

5. **Verify services are running**

   ```bash
   # Check broker service
   curl http://localhost:6000/ping

   # Check search service
   curl http://localhost:6001/ping
   ```

### Available Make Commands

```bash
make up                 # Start all services
make up_build           # Build and start all services
make down               # Stop all services
make postgres-up        # Start only PostgreSQL
make broker_start       # Run broker service locally
make search_service_start  # Run search service locally
make build_broker_service  # Build broker service binary
make build_search_service  # Build search service binary
```

## 📡 API Endpoints

### Broker Service (Port 6000)

**POST /broker**

Unified endpoint for all search types. Send requests with different actions:

#### Semantic Search

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "symantic_search",
    "symanticSearchPayload": {
      "query": "software engineer machine learning",
      "page": "0",
      "limit": "10"
    }
  }'
```

#### Lexical Search

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "lexical_search",
    "lexicalSearchPayload": {
      "query": "data scientist",
      "page": 0,
      "limit": 10
    }
  }'
```

#### Hybrid Search

```bash
curl -X POST http://localhost:6000/broker \
  -H "Content-Type: application/json" \
  -d '{
    "action": "hybrid_search",
    "hybridSearchPayload": {
      "query": "senior developer python",
      "page": 0,
      "limit": 10
    }
  }'
```

### Search Service (Port 6001)

Direct endpoints (also accessible via broker):

- **GET /ping** - Health check
- **GET /read-csv** - Read and parse CSV file
- **GET /store-csv** - Load CSV data into database with embeddings
- **POST /get-vector** - Generate embeddings for text
- **POST /symantic-search** - Semantic vector similarity search
- **POST /lexical-search** - BM25 full-text search
- **POST /hybrid-search** - Combined search approach

## 🗄️ Database Schema

```sql
CREATE TABLE jobs (
  id SERIAL PRIMARY KEY,
  created_at TIMESTAMP,
  updated_at TIMESTAMP,
  deleted_at TIMESTAMP,
  title VARCHAR(255),
  experience INTEGER,
  education_level VARCHAR(255),
  skills_count INTEGER,
  industry VARCHAR(255),
  company_size VARCHAR(255),
  location VARCHAR(255),
  remote_work VARCHAR(255),
  certifications INTEGER,
  salary INTEGER,
  embedding VECTOR(384)  -- 384-dimensional vector embedding
);

-- HNSW index for fast vector similarity
CREATE INDEX idx_jobs_embedding_hnsw
ON jobs USING hnsw (embedding vector_ip_ops);

-- BM25 index for full-text search
CREATE INDEX idx_jobs_search_idx_bm25
ON jobs USING bm25(id, title, education_level, industry, company_size, location)
WITH (key_field='id');
```

## 🔍 Search Types Explained

### 1. Semantic Search

- Uses vector embeddings to understand query meaning
- Finds jobs similar in _meaning_, not just keywords
- Example: "ML engineer" will match "Machine Learning Specialist"
- Best for: Conceptual searches, synonyms, related terms

### 2. Lexical Search

- BM25 algorithm for traditional keyword matching
- Frequency and relevance-based scoring
- Best for: Exact term matching, specific keywords

### 3. Hybrid Search

- Combines semantic and lexical approaches
- Reciprocal Rank Fusion (RRF) for result merging
- Best for: Most comprehensive and accurate results

## 🛠️ Technology Stack

- **Language**: Go 1.25
- **Database**: PostgreSQL with ParadeDB (pgvector + pg_search)
- **Embeddings**: Ollama (all-minilm model, 384 dimensions)
- **Container**: Docker & Docker Compose
- **Libraries**:
  - GORM (database ORM)
  - pgvector-go (vector operations)
  - net/http (HTTP server)

## 📊 Dataset

The project uses a job salary prediction dataset with fields:

- Job Title
- Years of Experience
- Education Level
- Skills Count
- Industry
- Company Size
- Location
- Remote Work Option
- Certifications
- Salary

Dataset size: ~10,000 job records

## 🔧 Configuration

### Environment Variables

**Search Service:**

```bash
PORT=80
DSN="host=postgres user=admin password=admin dbname=advance-search port=5432 sslmode=disable"
OLLAMA_EMBEDDING_URL="http://ollama:11434/api/embeddings"
```

**Broker Service:**

```bash
PORT=80
```

### Default Credentials

- **PostgreSQL**:
  - User: `admin`
  - Password: `admin`
  - Database: `advance-search`

## 🚦 Development

### Running Services Locally

1. **Start PostgreSQL and Ollama**:

   ```bash
   cd project
   docker-compose up -d postgres ollama
   ```

2. **Run Search Service**:

   ```bash
   cd search-service
   DSN="host=localhost user=admin password=admin dbname=advance-search port=5432 sslmode=disable" \
   PORT="6001" \
   OLLAMA_EMBEDDING_URL="http://localhost:11434/api/embeddings" \
   go run ./cmd/api/
   ```

3. **Run Broker Service**:
   ```bash
   cd broker-service
   PORT="6000" go run ./cmd/api/
   ```

### Building Binaries

```bash
# Build search service
cd search-service
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o searchApp ./cmd/api/

# Build broker service
cd broker-service
CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o brokerApp ./cmd/api/
```

## 📝 Notes

- First-time setup requires pulling the Ollama model (executed automatically)
- Loading the CSV dataset takes a few minutes depending on hardware
- Vector embeddings are generated using CPU by default
- HNSW index provides fast approximate nearest neighbor search

## 🤝 Contributing

Contributions are welcome! Please feel free to submit issues or pull requests.

## 📄 License

This project is available for educational and development purposes.
