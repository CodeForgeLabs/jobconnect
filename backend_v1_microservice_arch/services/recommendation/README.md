# Recommendation Service

Phase 1 delivers job recommendations for freelancers. Phase 1b adds freelancer
recommendations for clients. Phase 2 adds review-aware trust ranking through
the `review` service.

## What It Does

- reads freelancer profile and work preferences from `user`
- gathers public open jobs from `job`
- filters obvious mismatches before ranking
- ranks jobs with a hybrid score:
  - semantic profile-to-job text similarity
  - skill overlap
  - budget and rate fit
  - freshness
  - client review trust
- ranks freelancers with semantic, skill, rate, availability, and freelancer
  review trust signals
- caches top recommendations in memory or Redis for a short TTL

`GetRecommendedFreelancers` ranks discoverable freelancers for a client-owned
job. The gateway forwards the caller's authorization metadata so downstream job
ownership checks still apply, and freelancer recommendation cache entries are
scoped to the caller.

## Runtime

Default gRPC listen address:

- `:50064`

Default downstream addresses:

- `JOB_SERVICE_ADDR=localhost:50053`
- `USER_SERVICE_ADDR=localhost:50052`
- `REVIEW_SERVICE_ADDR=localhost:50056`

Tunable recommendation settings:

- `RECOMMENDATION_DEFAULT_LIMIT=10`
- `RECOMMENDATION_MAX_LIMIT=25`
- `RECOMMENDATION_CANDIDATE_PAGE_SIZE=100`
- `RECOMMENDATION_PER_SKILL_PAGE_SIZE=25`
- `RECOMMENDATION_MAX_SKILL_QUERIES=5`
- `RECOMMENDATION_CACHE_TTL=2m`
- `RECOMMENDATION_CACHE_BACKEND=memory` (`memory` or `redis`)
- `RECOMMENDATION_REDIS_ADDR=localhost:6379`
- `RECOMMENDATION_REDIS_PASSWORD=`
- `RECOMMENDATION_REDIS_DB=0`

Docker Compose runs the recommendation service with Redis enabled so cached
recommendations are shared across service instances. The Redis adapter stores
derived recommendation payloads only; Kafka events can later invalidate or
refresh those keys without changing the API path.

## Observability

The service logs cache hits, misses, disabled-cache paths, recomputation latency,
candidate counts, ranked counts, returned counts, and recomputation errors.
Redis cache get/decode/set failures are logged inside the Redis adapter.

## Cache Invalidation

Internal callers can use `InvalidateRecommendationCache` to clear cached entries
after upstream `job`, `user`, or `review` changes. The RPC supports user-scoped
job recommendation invalidation, job-scoped freelancer recommendation
invalidation across caller-specific cache entries, and full recommendation cache
clear for broad refresh events.

## Gateway Endpoint

Authenticated freelancer endpoint:

- `GET /api/v1/recommendations/jobs?limit=10`

Authenticated client endpoint:

- `GET /api/v1/recommendations/jobs/{job_id}/freelancers?limit=10`

Response body:

```json
{
  "recommendations": [
    {
      "job_id": 101,
      "match_score": 0.92,
      "match_reason": "Matches your skills in Go, gRPC"
    }
  ]
}
```

Freelancer recommendation response:

```json
{
  "recommendations": [
    {
      "user_id": "freelancer-uuid",
      "match_score": 0.88,
      "match_reason": "Matches required skills: Go, PostgreSQL"
    }
  ]
}
```

## Roadmap

### Phase 4 — Local Semantic Matching (planned)

Full plan lives in `implementation_plan.md`. Summary:

- Replace token-cosine semantic scoring with `sentence-transformers/all-MiniLM-L6-v2` (22 MB, 384-dim). Runs fully locally on CPU; no external API, no new microservice.
- Python worker shipped inside the recommendation container. Go binary supervises it and talks over a Unix socket through an `Embedder` port. Swap the model (or later migrate to pure-Go ONNX) without touching the ranker.
- Vectors stored in a new recommendation-owned Postgres database using pgvector (Postgres extension, not a separate service) with an HNSW cosine-distance index.
- Embeddings are computed lazily on first read and deduped by `sha256(normalized_text)` so trivial edits do not trigger re-embedding.
- Vector search replaces the broad skill-based candidate pull; existing hard filters (visibility, availability, rate) still run on hydrated data. Token-cosine remains as the graceful-degradation path when the embedder or vector store is unavailable.

**Prerequisite:** Phase 3b metrics export must ship first so that Phase 4 quality and latency changes are measurable.

### Deferred
- Phase 3d — Kafka-driven refresh (re-embed + cache invalidation on upstream changes).
- Phase 3e — Background precomputation worker for active users and active jobs.
- Collaborative filtering ("users who applied for X also applied for Y") — needs an interaction event stream; revisit after Kafka.
