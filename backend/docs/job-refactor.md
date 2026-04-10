# Job Service Proto Refactor

Date: 2026-04-09

## Changes to `api/proto/job/v1/job.proto`

---

### 1. Removed `experience_level` entirely

`SetJobExperienceLevel` RPC and `ExperienceLevel` enum were removed from the service.

**Reason:** The platform should not definitively label freelancers by experience level. That is a client preference, not a platform-enforced attribute. Keeping it would introduce unfair or inaccurate classifications.

---

### 2. Added `storage_key` to `JobAttachment`

```proto
message JobAttachment {
  ...
  string storage_key = 6;
}
```

**Reason:** `url` is a presigned/CDN URL that can expire or change. `storage_key` is the stable internal MinIO/S3 identifier (e.g. `jobs/123/attachments/456/brief.pdf`) used to regenerate URLs and reference the object reliably. The gRPC handler strips `storage_key` from public-facing responses.

---

### 3. Added lifecycle timestamps and settlement fields to `Job`

```proto
message Job {
  ...
  int64 completed_at_unix_seconds = 23;
  int64 canceled_at_unix_seconds = 24;
  CloseReason close_reason = 25;
  SettlementPolicy settlement_policy = 26;
}
```

**Reason:**
- `completed_at_unix_seconds` — records when `MarkJobCompleted` fired. Needed for contract duration calculations and earnings history.
- `canceled_at_unix_seconds` — records when `CancelJobWithSettlementPolicy` fired. Distinct from `closed_at`; needed for payment settlement timing and dispute windows.
- `close_reason` — records why the job was closed. Needed for platform analytics and refund/dispute flows.
- `settlement_policy` — records which policy (`REFUND_REMAINING` or `NO_REFUND`) was applied on cancellation. Downstream services (payment, wallet) read this instead of re-deriving it.

---

### 4. Added `JOB_STATUS_COMPLETED` and `JOB_STATUS_CANCELED` to `JobStatus`

```proto
enum JobStatus {
  ...
  JOB_STATUS_COMPLETED = 5;
  JOB_STATUS_CANCELED = 6;
}
```

**Reason:** `MarkJobCompleted` and `CancelJobWithSettlementPolicy` RPCs existed but had no corresponding status values. Without these, `status_enum` would return `UNSPECIFIED` after either RPC fired.

---

### 5. Fixed `ListInvitedJobsResponse` — added `InvitedJob` wrapper

Before:
```proto
message ListInvitedJobsResponse {
  repeated Job invites = 1; // lost invite metadata
  string next_page_token = 2;
}
```

After:
```proto
message InvitedJob {
  Job job = 1;
  JobInvite invite = 2;
}

message ListInvitedJobsResponse {
  repeated InvitedJob invites = 1;
  string next_page_token = 2;
}
```

**Reason:** The previous response returned `Job` objects only, discarding invite metadata (`invited_at_unix_seconds`, `response_status`). A freelancer viewing their invites needs to know when they were invited and whether they already responded. `JobInvite` was already defined in the proto but never used — it is now the invite record in the response.

---

### 6. Added `optional Visibility visibility` to `CreateJobRequest`

```proto
message CreateJobRequest {
  ...
  optional Visibility visibility = 11;
}
```

**Reason:** Previously, visibility could only be set after job creation via `SetJobVisibility`. A new job would default to `VISIBILITY_UNSPECIFIED` with no way to publish it atomically at creation time. Making it optional allows clients to set it on creation while keeping it non-required (server defaults to `VISIBILITY_PRIVATE` when omitted).

---

### 7. Removed dual string/enum fields — enum-only going forward

Removed plain string fields that coexisted with their enum equivalents. Field numbers and names are now `reserved` to prevent silent reuse.

| Message | Removed field | Reserved |
|---|---|---|
| `Job` | `string job_type = 6` | `reserved 6; reserved "job_type"` |
| `Job` | `string status = 12` | `reserved 12; reserved "status"` |
| `CreateJobRequest` | `string job_type = 4` | `reserved 4; reserved "job_type"` |
| `UpdateJobRequest` | `optional string job_type = 5` | `reserved 5; reserved "job_type"` |
| `ListMyJobsRequest` | `string status = 1` | `reserved 1; reserved "status"` |
| `ListOpenJobsRequest` | `string job_type = 5` | `reserved 5; reserved "job_type"` |
| `CloseJobRequest` | `string reason = 2` | `reserved 2; reserved "reason"` |

**Reason:** The string fields were added first, before enums were introduced. Enums were added later as the proper replacement but both were kept for backwards compatibility, creating two sources of truth. Enums are strictly better: type-safe, no string parsing/validation, well-defined zero value (`UNSPECIFIED`), and the proto compiler handles unknown values gracefully.

`reserved` on both the number and name ensures the compiler rejects any future field that tries to reuse a retired slot — protecting both binary wire encoding (field numbers) and name-based access (JSON mapping, reflection).

---

## What needs to happen next (downstream)

After running `buf generate --config buf.gen.job.yaml`:

1. `job_server.go` — update all handlers that mapped from the old string fields to use enum fields directly.
2. `settlement.go` — application layer currently accepts `string` for `SettlementPolicy`; align it to use the proto enum after the grpc adapter maps it.
3. `migrations/` — new migration needed to add `completed_at`, `canceled_at`, `settlement_policy`, `cancellation_reason` columns if not already present (check `0005_freelancer_and_settlement.up.sql`).
4. `job_repo.go` — `ListInvitedJobs` query needs to join on `job_invites` to populate the `JobInvite` fields in the response.
