-- Enable the vector extension
CREATE EXTENSION IF NOT EXISTS vector;

-- Create index on the "embeddings" column
CREATE INDEX IF NOT EXISTS idx_jobs_embedding_hnsw 
ON jobs USING hnsw (embedding vector_ip_ops);

-- Enable the pg_search extension for bm25 full text search
CREATE EXTENSION IF NOT EXISTS pg_search;


-- Create bm25 index on search_text column
CREATE INDEX IF NOT EXISTS idx_jobs_search_idx_bm25
ON jobs USING bm25(id, title, education_level, industry, company_size, location)
WITH (key_field='id');