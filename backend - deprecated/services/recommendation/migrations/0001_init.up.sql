CREATE EXTENSION IF NOT EXISTS vector;

CREATE TABLE IF NOT EXISTS freelancer_embeddings (
    user_id     TEXT PRIMARY KEY,
    text_hash   TEXT NOT NULL,
    embedding   vector(384) NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS job_embeddings (
    job_id      BIGINT PRIMARY KEY,
    text_hash   TEXT NOT NULL,
    embedding   vector(384) NOT NULL,
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_freelancer_embeddings_hnsw
    ON freelancer_embeddings USING hnsw (embedding vector_cosine_ops);

CREATE INDEX IF NOT EXISTS idx_job_embeddings_hnsw
    ON job_embeddings USING hnsw (embedding vector_cosine_ops);
