# Recommendation Service

Phase 1 delivers job recommendations for freelancers.

## What It Does

- reads freelancer profile and work preferences from `user`
- gathers public open jobs from `job`
- filters obvious mismatches before ranking
- ranks jobs with a hybrid score:
  - semantic profile-to-job text similarity
  - skill overlap
  - budget and rate fit
  - freshness
- caches top recommendations in memory for a short TTL

`GetRecommendedFreelancers` is intentionally left unimplemented in Phase 1.

## Runtime

Default gRPC listen address:

- `:50064`

Default downstream addresses:

- `JOB_SERVICE_ADDR=localhost:50053`
- `USER_SERVICE_ADDR=localhost:50052`

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
