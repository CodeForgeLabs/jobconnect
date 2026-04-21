# Job Service – DB Tables

---

## Core

### `jobs`

- **Primary key:** `id` (BIGSERIAL)

**Identity**

- `client_id` (UUID, NOT NULL) — logical reference to the owning client (no DB FK; cross-service)

**Attributes**

- `title` (TEXT, NOT NULL) — max 160 chars enforced in domain
- `description` (TEXT, NOT NULL) — max 10,000 chars enforced in domain
- `required_skills` (JSONB, NOT NULL, default `[]`)
- `job_type` (TEXT, NOT NULL) — check: `fixed` | `hourly`
- `budget_fixed` (DOUBLE PRECISION, NOT NULL, default 0)
- `hourly_rate` (DOUBLE PRECISION, NOT NULL, default 0)
- `currency` (VARCHAR(8), NOT NULL, default `USD`)
- `deadline` (TIMESTAMPTZ, NULL)
- `budget_min` (DOUBLE PRECISION, NOT NULL, default 0) — optional range set via `SetJobBudgetRange`
- `budget_max` (DOUBLE PRECISION, NOT NULL, default 0) — optional range set via `SetJobBudgetRange`

**Lifecycle / governance**

- `status` (TEXT, NOT NULL, default `open`) — check: `open` | `paused` | `filled` | `closed` | `completed` | `canceled`
- `visibility` (TEXT, NOT NULL, default `public`) — check: `public` | `private` | `invite_only`
- `close_reason` (TEXT, NOT NULL, default `''`)
- `settlement_policy` (TEXT, NOT NULL, default `''`) — `refund_remaining` | `no_refund`; set on cancellation
- `cancellation_reason` (TEXT, NOT NULL, default `''`)
- `created_at` (TIMESTAMPTZ, NOT NULL)
- `updated_at` (TIMESTAMPTZ, NOT NULL)
- `closed_at` (TIMESTAMPTZ, NULL)
- `paused_at` (TIMESTAMPTZ, NULL)
- `filled_at` (TIMESTAMPTZ, NULL)
- `completed_at` (TIMESTAMPTZ, NULL)
- `canceled_at` (TIMESTAMPTZ, NULL)

**Constraints**

- `jobs_budget_by_type` — fixed jobs must have `budget_fixed > 0` and `hourly_rate = 0`; hourly jobs must have `hourly_rate > 0` and `budget_fixed = 0`
- `jobs_status_check` — enum check on `status`
- `jobs_visibility_check` — enum check on `visibility`

**Indexes**

- `idx_jobs_client_id_created_at(client_id, created_at DESC)`
- `idx_jobs_status_created_at(status, created_at DESC)`

**Referenced by**

- `job_attachments` via `job_id`
- `job_invites` via `job_id`
- `saved_jobs` via `job_id`

---

## Attachments

### `job_attachments`

- **Primary key:** `id` (BIGSERIAL)
- `job_id` (BIGINT, NOT NULL) — FK → `jobs.id` (ON DELETE CASCADE)

**Attributes**

- `file_name` (TEXT, NOT NULL)
- `content_type` (TEXT, NOT NULL)
- `url` (TEXT, NOT NULL) — presigned/CDN URL; generated at read time
- `storage_key` (TEXT, NOT NULL, default `''`) — stable internal MinIO/S3 key; not exposed publicly
- `size_bytes` (BIGINT, NOT NULL, default 0)

**Indexes**

- `idx_job_attachments_job_id(job_id, id)`
- `idx_job_attachments_storage_key(storage_key)`

---

## Invites

### `job_invites`

- **Primary key:** `id` (BIGSERIAL)
- `job_id` (BIGINT, NOT NULL) — FK → `jobs.id` (ON DELETE CASCADE)

**Attributes**

- `client_id` (UUID, NOT NULL) — who sent the invite
- `freelancer_id` (UUID, NOT NULL) — who received it
- `created_at` (TIMESTAMPTZ, NOT NULL)
- `response_status` (TEXT, NOT NULL, default `unspecified`) — check: `unspecified` | `accepted` | `declined`
- `responded_at` (TIMESTAMPTZ, NULL)

**Unique constraint**

- `(job_id, freelancer_id)` — a freelancer can only be invited once per job

**Indexes**

- `idx_job_invites_job_id_created_at(job_id, created_at DESC)`

---

## Freelancer interactions

### `saved_jobs`

- **Primary key:** `id` (BIGSERIAL)
- `job_id` (BIGINT, NOT NULL) — FK → `jobs.id` (ON DELETE CASCADE)
- `freelancer_id` (UUID, NOT NULL) — logical reference (no DB FK; cross-service)
- `created_at` (TIMESTAMPTZ, NOT NULL)

**Unique constraint**

- `(job_id, freelancer_id)` — a freelancer can only save a job once

**Indexes**

- `idx_saved_jobs_freelancer_created_at(freelancer_id, created_at DESC)`

---

## Pending migrations

All proto-driven schema changes are resolved by migration `0006_status_and_experience_cleanup`, which expands `jobs_status_check` to include `completed` and `canceled` and drops the legacy `experience_level` column and its check constraint.
