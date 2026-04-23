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
    - Track cache hits, misses, set failures, and downstream recomputation latency.
    - Log candidate counts and cache backend selection at startup.
- **Phase 3c — Explicit Invalidation API**:
    - Add invalidation methods for freelancer profile/work-preference changes, job changes, and review summary changes.
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

## **Phase 4 — Semantic Search & AI (Future)**
Transition from keyword matching to "Intent" matching.

- **Vector Search**: Use OpenAI/Mistral embeddings to match jobs based on description (e.g., a "Backend Go Engineer" job matches "Go Developer" even without exact skill tags).
- **Collaborative Filtering**: "Users who applied for Job X also applied for Job Y."
- **Why**: Deep discovery that simple skill tags can't capture.

---

## **Architecture Overview**
This service follows the **Hexagonal Architecture**:
- `internal/application`: Matching logic and ranking algorithms.
- `internal/adapters/grpc`: Clients for `job`, `user`, and `review` services.
- `internal/domain`: Scoring models and recommendation entities.
