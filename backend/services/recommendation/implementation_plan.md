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

- **Background Workers**: Periodically pre-calculate recommendations for top active users.
- **Caching Layer**: Store recommendations in Redis/PostgreSQL for instant retrieval.
- **Events**: Invalidate/update recommendations when a new job is posted or a profile is updated (via Pub/Sub).
- **Why**: To maintain <100ms response times at scale.

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
