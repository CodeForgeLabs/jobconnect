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
- caches top recommendations in memory for a short TTL

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
