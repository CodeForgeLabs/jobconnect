# Recommendation Service Implementation Plan

The Recommendation Service is the core "Discovery" engine for JobConnect. It improves platform stickiness by proactively suggesting relevant jobs to freelancers and top talent to clients.

## **Phase 1 — Basic Rule-Based Matching (Heuristics)**
The goal is to get the service running with a simple, deterministic algorithm.

- **RPCs**:
    - `GetRecommendedJobs`: Match freelancer's `Skills` against job `RequiredSkills`.
    - `GetRecommendedFreelancers`: Match job's `RequiredSkills` against freelancer's `Skills`.
- **Ranking**:
    - Primary: **Skill Overlap Count** (Higher is better).
    - Secondary: **Recency** (Newest jobs/active profiles first).
- **Implementation**:
    - Act as a gRPC client to `job` and `user` services.
    - Query and filter jobs/freelancers in memory (since the user base is currently small).
- **Why**: Low implementation cost with high immediate value.

---

## **Phase 2 — Trust-Aware Ranking**
Incorporate ratings and transaction data to improve recommendation quality.

- **Ranking Updates**:
    - Boost jobs from **Highly Rated Clients** (using `review` service data).
    - Boost **Highly Rated Freelancers**.
- **Data Integration**:
    - Connect to the `review` service to fetch average ratings and review counts.
- **Why**: Quality control. Users are more likely to trust recommendations if they prioritize reliable partners.

---

## **Phase 3 — Latency & Caching (The "Discovery" Cache)**
As the number of jobs and users grows, real-time filtering becomes too slow.

- **Phase 3a — Redis Cache Backend**:
    - Add a Redis implementation behind the existing `RecommendationCache` port.
    - Keep the in-memory cache available for tests and simple local runs.
    - Cache final ranked job recommendations by freelancer ID.
    - Cache final ranked freelancer recommendations by job ID plus caller scope.
    - Store cache entries as short-lived JSON documents with Redis TTLs.
- **Phase 3b — Cache Observability**:
    - Emit structured service logs for cache hits, misses, disabled cache, recomputation latency, candidate counts, ranked counts, returned counts, and recomputation errors.
    - Emit Redis adapter logs for cache get/decode/set failures.
    - Log cache backend selection at startup.
    - Later, export the same signals to the platform metrics backend once one is chosen.
- **Phase 3c — Explicit Invalidation API**:
    - Add an internal `InvalidateRecommendationCache` RPC for targeted cache invalidation.
    - Support invalidating job recommendations by freelancer/user ID.
    - Support invalidating freelancer recommendations by job ID across all caller-scoped cache entries.
    - Support full recommendation cache clear for broad or emergency invalidations.
    - Keep invalidation policy in the application layer, not inside the Redis adapter.
- **Phase 3d — Kafka-Driven Refresh (Later)**:
    - Consume `job`, `user`, and `review` events from Kafka.
    - Use events to invalidate Redis keys or trigger background recomputation.
    - Example events: `JobCreated`, `JobUpdated`, `JobClosed`, `FreelancerProfileUpdated`, `WorkPreferencesUpdated`, `ReviewCreated`, `ReviewUpdated`, `ReviewDeleted`.
- **Phase 3e — Background Precomputation**:
    - Periodically pre-calculate recommendations for active freelancers and active client jobs.
    - API requests should become cache read first, compute fallback second.
- **Why**: To maintain <100ms response times at scale while keeping Kafka as the future source of freshness signals and Redis as the fast read store.

---

## **Phase 4 — Local Semantic Matching**
Move from token-bag cosine similarity to true sentence embeddings without adding a new microservice, external API, or vendor dependency. Everything runs locally.

**Prerequisite — Phase 3b metrics export.** Before starting Phase 4, finish exporting the structured cache/recompute signals (hits, misses, recomputation latency, candidate/ranked/returned counts, errors) to the platform metrics backend. Without metrics flowing to a dashboard, there is no way to prove that embeddings improved match quality or regressed latency. This is a hard gate, not a nice-to-have.

- **Model choice**:
    - `sentence-transformers/all-MiniLM-L6-v2` — 22 MB on disk, 384-dim output, ~2 ms per sentence on an M3 CPU.
    - Swappable via config. `BAAI/bge-small-en-v1.5` (133 MB, 384-dim) is a drop-in quality upgrade if MiniLM underperforms.
    - Kept behind an `Embedder` port so the ranker stays model-agnostic.
- **Why not OpenAI / Mistral**: external API dependency, cost, and vendor lock-in are not worth it at project scale.
- **Why not LSTM-era encoders**: distilled BERT models (MiniLM family) hit the same "tiny + CPU-only" target with much better retrieval quality. Same goal, better tool.

### **Phase 4a — Embedder Port + Python Subprocess Adapter**
Bake the embedder into the recommendation container. No new compose service, no new process supervised outside Go.

- Add `Embedder` port in `internal/application`:
    - `Embed(ctx context.Context, texts []string) ([][]float32, error)`
- Python worker script shipped inside the recommendation container:
    - Loads `all-MiniLM-L6-v2` at startup via `sentence-transformers`.
    - Reads newline-delimited JSON requests over a Unix domain socket, writes newline-delimited JSON responses (batch in, batch out).
- Go adapter in `internal/infrastructure/embedder/python`:
    - Spawns the Python worker as a long-lived subprocess at service startup.
    - Health checks the socket before marking the service ready.
    - Restarts the subprocess on crash with exponential backoff.
- Recommendation container Dockerfile:
    - Adds Python 3.12 + `sentence-transformers` + PyTorch CPU wheel.
    - Pre-downloads model weights at image build time so runtime start is offline.
    - ~500 MB image size increase is acceptable.
- Config additions:
    - `RECOMMENDATION_EMBEDDER_BACKEND=python` (values: `noop`, `python`).
    - `RECOMMENDATION_EMBEDDER_SOCKET=/tmp/recommendation-embedder.sock`.
    - `RECOMMENDATION_EMBEDDER_MODEL=sentence-transformers/all-MiniLM-L6-v2`.
    - `RECOMMENDATION_EMBEDDER_BATCH_SIZE=32`.
- Graceful degradation:
    - Subprocess dead, socket errored, or embedder disabled → adapter returns a sentinel error, ranker falls back to the existing token-cosine path. The RPC never hard-fails because the embedder is down.

### **Phase 4b — pgvector Store in a Recommendation Postgres DB**
Give the recommendation service its own Postgres DB (first time — it has been stateless so far). No new compose service: pgvector is a Postgres extension, not a separate binary.

- Docker compose:
    - New `jobconnect_recommendation` logical database inside the existing Postgres container.
    - Enable the `vector` extension in that DB.
    - New env: `RECOMMENDATION_POSTGRES_URL`.
- Migrations (owned by recommendation service):
    - `freelancer_embeddings(user_id TEXT PK, text_hash TEXT, embedding vector(384), updated_at TIMESTAMPTZ)`.
    - `job_embeddings(job_id BIGINT PK, text_hash TEXT, embedding vector(384), updated_at TIMESTAMPTZ)`.
    - HNSW indexes on the `embedding` columns with cosine distance.
- `EmbeddingStore` port:
    - `GetEmbedding(ctx, sourceType, sourceID) (Embedding, bool, error)`.
    - `UpsertEmbedding(ctx, sourceType, sourceID, textHash, vector) error`.
    - `SearchByVector(ctx, sourceType, vector, k int) ([]VectorHit, error)`.
- pgvector adapter in `internal/infrastructure/vectorstore/pgvector`.
- Ownership note: embeddings are a **retrieval concern** and belong to the recommendation service. The `user` and `job` services should not carry vector columns or vector-search RPCs.

### **Phase 4c — Lazy Embedding with Text-Hash Dedup**
Compute on first read, cache on the hash, skip re-embedding when text is unchanged.

- Text hash:
    - `sha256(normalize(text))[:16]` where normalize = trim + collapse whitespace + lowercase.
    - Normalization is intentionally coarse so trivial whitespace edits do not trigger re-embed, but substantive edits do.
- Per recommendation request:
    - Hash the query text (freelancer profile for `GetRecommendedJobs`, job description for `GetRecommendedFreelancers`) and every candidate text.
    - Look up `(source_id, text_hash)` in the embedding store. Hash match → reuse vector. Hash mismatch or miss → queue for batch compute.
    - Single batched `Embedder.Embed(texts)` call per request for all misses combined (bounded by `RECOMMENDATION_EMBEDDER_BATCH_SIZE`).
    - Upsert the recomputed vectors with the new hash.
- Never block the ranker: if the embed call fails mid-request, log + fall back to token cosine for that request only. The embedding store state is unchanged.

### **Phase 4d — Swap the Semantic Signal in the Rankers**
Plug embedding cosine into the existing hybrid score. No weight changes — only the semantic term quality improves.

- Replace the `buildTokenVector` + `cosineSimilarity` call inside `rankJobs` / `rankFreelancers` with embedding cosine when vectors are available.
- Keep the token-cosine path as the fallback for "embedder unavailable" and "candidate has no embedding yet".
- Observability: per-request log which semantic path served each candidate (`embedding` vs `token`). Lets you compare quality via existing recommendation logs without a second code path in the API.

### **Phase 4e — Vector-Search Candidate Retrieval**
Use pgvector to pick candidates instead of broad skill-based pulls. Keeps the candidate set sharply relevant even when skill tags do not overlap.

- For `GetRecommendedJobs`:
    - Embed the freelancer profile → `SearchByVector("job", vec, CandidatePageSize)` in pgvector → hydrate hit IDs via `JobService.SearchJobsV2` / `GetJob`.
    - Apply the existing hard filters (visibility, preferences, contract type) on hydrated data.
- For `GetRecommendedFreelancers`:
    - Embed the job description → `SearchByVector("freelancer", vec, CandidatePageSize)` → hydrate via `UserService.ListDiscoverableFreelancers` / `GetMyProfile`.
    - Apply availability + rate filters on hydrated data.
- Graceful fallback: pgvector search fails or returns empty → fall back to the current skill-based candidate pull. The ranker still runs.

### **Phase 4f — Embedding Refresh Policy**
- Near term: lazy-on-read (Phase 4c) handles new users, new jobs, and edited text.
- Mid term (ties into deferred Phase 3d): when Kafka lands, `FreelancerProfileUpdated` / `JobUpdated` / `JobClosed` events trigger re-embed + vector upsert + invalidation of cached recommendations. Removes the "stale vector until next read" window.
- Safety net until Kafka: a background tick can re-validate vectors older than a configurable TTL (e.g., 7 days). Optional; skip if lazy-on-read is sufficient in practice.

### **Phase 4 Test Plan**
- Unit: `Embedder` fake returns deterministic vectors; ranker tests verify that embedding similarity flips ranking order vs token cosine when the two diverge.
- Unit: text-hash dedup — same text twice hits the embedder exactly once; mutated text triggers re-embed.
- Unit: pgvector adapter upsert/search round-trip against a test Postgres container (reuse existing service test infra).
- Unit: adapter degradation — killed subprocess / socket error path returns the sentinel and the ranker falls back cleanly.
- Regression: existing Phase 1b / Phase 2 ranking tests still pass with the embedder wired in (token-cosine fallback path preserved).

### **Phase 4 Non-Goals**
- **Collaborative filtering** ("users who applied for X also applied for Y") stays deferred. It needs an application/interaction event stream that does not exist yet; add after Kafka.
- Pure Go inference via ONNX Runtime. Viable, but cgo build complexity and thin Go tokenizer ecosystem make it a time sink for a project-scale codebase. Python subprocess was chosen deliberately for fast iteration. Migration to a single-binary Go+ONNX build later is a pure adapter swap under the `Embedder` port.

---

## **Architecture Overview**
This service follows the **Hexagonal Architecture**:
- `internal/application`: Matching logic and ranking algorithms.
- `internal/adapters/grpc`: Clients for `job`, `user`, and `review` services.
- `internal/domain`: Scoring models and recommendation entities.
