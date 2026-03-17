# Review Service — Technical Review

## Overview

The Review service is a standalone gRPC microservice that enables **mutual reviews** between clients and freelancers after a contract engagement. It follows the same hexagonal (ports-and-adapters) architecture used by the existing Job and Contract services.

**Port:** `:50055` (configurable via `REVIEW_GRPC_LISTEN_ADDR`)

---

## Files Created

```
api/proto/review/v1/review.proto        ← gRPC service + message definitions

services/review/
├── cmd/reviewd/main.go                 ← Service entrypoint
├── gen/review/v1/                      ← Auto-generated protobuf Go code
├── go.mod / go.sum                     ← Go module dependencies
├── migrations/
│   ├── 0001_init.up.sql                ← Schema creation
│   └── 0001_init.down.sql              ← Rollback
└── internal/
    ├── adapters/grpc/
    │   ├── auth.go                     ← JWT extraction from gRPC metadata
    │   ├── review_server.go            ← RPC handler implementations
    │   └── server.go                   ← gRPC registration wrapper
    ├── application/
    │   ├── ports.go                    ← ReviewRepository + Clock interfaces
    │   ├── create_review.go            ← CreateReview use case
    │   ├── get_review.go               ← GetReview use case
    │   ├── list_reviews_by_user.go     ← ListReviewsByUser use case
    │   ├── list_reviews_by_contract.go ← ListReviewsByContract use case
    │   └── get_user_rating_summary.go  ← GetUserRatingSummary use case
    ├── config/config.go                ← Env-based configuration
    ├── domain/review.go                ← Review struct + validation
    └── infrastructure/
        ├── clock/clock.go              ← UTC clock
        ├── db/postgres.go              ← Connection pool
        ├── db/helpers.go               ← ErrNotFound sentinel
        ├── db/review_repo.go           ← Postgres repository
        └── tokens/jwt.go               ← JWT parser
```

---

## Endpoints (RPCs)

| RPC | Auth | Description |
|-----|------|-------------|
| `CreateReview` | ✅ JWT required | Submit a review for the other party on a contract. The reviewer's identity and role are extracted from the JWT token. |
| `GetReview` | ❌ Public | Fetch a single review by its ID. |
| `ListReviewsByUser` | ❌ Public | Paginated list of all reviews **received by** a given user (for public profile pages). |
| `ListReviewsByContract` | ❌ Public | List the review(s) associated with a specific contract (up to 2: one from each party). |
| `GetUserRatingSummary` | ❌ Public | Returns the average rating and total review count for a user. |

### Request/Response Examples

**CreateReview** (requires `authorization: Bearer <token>` in metadata):
```json
// Request
{
  "contract_id": 1,
  "reviewee_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
  "rating": 5,
  "title": "Excellent work",
  "comment": "Delivered ahead of schedule with high quality code."
}
// Response
{
  "review": {
    "id": 1,
    "contract_id": 1,
    "reviewer_id": "d043c58d-f012-4325-8ca9-327bb14fccad",
    "reviewee_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee",
    "reviewer_role": "client",
    "rating": 5,
    "title": "Excellent work",
    "comment": "Delivered ahead of schedule with high quality code.",
    "created_at_unix_seconds": 1742130000
  }
}
```

**GetUserRatingSummary**:
```json
// Request
{ "user_id": "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee" }
// Response
{ "average_rating": 4.75, "total_reviews": 12 }
```

---

## Database Schema

```sql
CREATE TABLE reviews (
    id              BIGSERIAL PRIMARY KEY,
    contract_id     BIGINT NOT NULL,
    reviewer_id     UUID NOT NULL,
    reviewee_id     UUID NOT NULL,
    reviewer_role   TEXT NOT NULL CHECK (reviewer_role IN ('client','freelancer')),
    rating          INT NOT NULL CHECK (rating BETWEEN 1 AND 5),
    title           TEXT NOT NULL,
    comment         TEXT NOT NULL,
    created_at      TIMESTAMPTZ NOT NULL,
    CONSTRAINT unique_review_per_contract UNIQUE (contract_id, reviewer_id)
);
```

**Indexes:**
- `idx_reviews_reviewee` — fast lookups for "all reviews received by user X" (used on profile pages)
- `idx_reviews_contract` — fast lookups for "reviews on contract Y"

**Key constraint:** `UNIQUE (contract_id, reviewer_id)` ensures each party can only leave **one** review per contract.

---

## Architecture Decisions

### 1. Reviews are tied to contracts, not jobs
A single job posting can result in multiple contracts with different freelancers. Tying reviews to the contract ensures each review is about a specific working relationship, not a general job.

### 2. Both parties review independently
A client can review the freelancer, and the freelancer can review the client — producing up to **two** review records per contract. Each is stored as a separate row. This is the industry-standard approach (Upwork, Fiverr, etc.).

### 3. Reviewer identity comes from the JWT
The `CreateReview` RPC does **not** accept a `reviewer_id` in the request body. Instead, it extracts the caller's UUID and role from the JWT token in the `authorization` metadata header. This prevents impersonation.

### 4. Immutable reviews
Reviews cannot be edited or deleted after submission. This is a deliberate trust/integrity decision. Users see what was written at the time of submission.

### 5. Rating scale: 1–5 integer
Simple, universally understood. Enforced both at the domain validation layer and via a PostgreSQL `CHECK` constraint.

### 6. No cross-service calls
The Review service does **not** call the Contract service to verify that a contract exists or is in "ended" status. It trusts the caller. This keeps the service fully independent and avoids coupling.

---

## Business Assumptions

1. **Only parties to a contract can review each other.** The current implementation trusts that the caller is a legitimate party. In production, you'd ideally verify this against the Contract service.
2. **Reviews are public.** `GetReview`, `ListReviewsByUser`, and `GetUserRatingSummary` do not require authentication. This aligns with marketplace platforms where reviews are visible to everyone.
3. **One review per party per contract.** Enforced by the `UNIQUE (contract_id, reviewer_id)` constraint. Attempting a second review returns an error.
4. **No moderation pipeline.** Reviews are published immediately. There is no approval queue or content moderation.
5. **No notification system.** When a review is created, the reviewee is not notified. This would be handled by an event bus or notification service in the future.

---

## Room for Improvements

### Short-term
| Improvement | Rationale |
|-------------|-----------|
| **Contract existence check** | Call the Contract service (or share DB) to verify the contract exists and is in "ended" status before allowing a review. |
| **Review window** | Only allow reviews within N days of contract end. Prevents stale reviews months later. |
| **UpdateReview / DeleteReview RPCs** | Allow users to edit or retract reviews within a grace period (e.g., 24 hours). |
| **Content moderation** | Filter profanity, spam, or abusive language before persisting. |
| **`.env` and `README.md`** | Create env file and setup guide (same pattern as Job/User services). |

### Medium-term
| Improvement | Rationale |
|-------------|-----------|
| **Event-driven notifications** | Publish a `ReviewCreated` event so other services (notification, email) can react. |
| **Review response** | Let the reviewee post a single public reply to a review (common on Google, Airbnb). |
| **Weighted rating** | Factor in recency or contract value when computing average ratings. |
| **Pagination on `ListReviewsByContract`** | Currently returns all reviews (max 2). If the model changes, pagination may be needed. |

### Long-term
| Improvement | Rationale |
|-------------|-----------|
| **Review analytics dashboard** | Aggregate trends (rating over time, sentiment analysis). |
| **Cross-service rating sync** | Push updated average ratings to the User service for profile display without requiring a separate RPC call. |
| **Rate limiting** | Prevent abuse (e.g., creating/deleting accounts to leave multiple reviews). |
| **Audit log** | Track review creation timestamps and any future edits for compliance. |
