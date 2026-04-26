# Proposal Service Postman Documentation

## What this package contains

- `docs/postman/proposal-gateway.postman_collection.json`
- `docs/postman/proposal-local.postman_environment.json`

This package documents the Proposal service from a Postman perspective.

## Import steps

1. Open Postman.
2. Import `proposal-gateway.postman_collection.json`.
3. Import `proposal-local.postman_environment.json`.
4. Select environment `JobConnect Local`.
5. Run `Auth Bootstrap > Login Client` and `Auth Bootstrap > Login Freelancer` first.

The login requests automatically store bearer tokens into:
- `clientToken`
- `freelancerToken`

## User flow in Postman

### Freelancer flow

1. Submit proposal (gRPC-only currently, see matrix below).
2. `GET /api/v1/proposals/me/jobs/{jobId}/has-applied`
3. `GET /api/v1/proposals/me/jobs/{jobId}`
4. `GET /api/v1/proposals/me`
5. `POST /api/v1/proposals/{proposalId}/attachments/upload-url`
6. Modify/Withdraw proposal (gRPC-only currently, see matrix below).

### Client flow

1. `GET /api/v1/proposals/client`
2. `GET /api/v1/proposals/{proposalId}`
3. `GET /api/v1/proposals/jobs/{jobId}/counts`
4. `GET /api/v1/proposals/client/counts`
5. `POST /api/v1/proposals/{proposalId}/decision`
6. `GET /api/v1/proposals/{proposalId}/attachments/{attachmentId}/download-url`
7. Open the offer composer in the client UI (no backend mutation yet).
8. Send the offer through Contract API `CreateContract`, which marks the proposal as `hired` only after the offer is created.

## RPC coverage matrix

| Proposal RPC | Actor | Gateway HTTP route | Postman request included | Notes |
|---|---|---|---|---|
| SubmitProposal | Freelancer | Not exposed | No | Use gRPC client (grpcurl/Postman gRPC tab) |
| ModifyProposal | Freelancer | Not exposed | No | Use gRPC client |
| WithdrawProposal | Freelancer | Not exposed | No | Use gRPC client |
| GetProposal | Client/Freelancer owner | `GET /api/v1/proposals/{proposalId}` | Yes | |
| GetMyProposalForJob | Freelancer | `GET /api/v1/proposals/me/jobs/{jobId}` | Yes | |
| HasAppliedToJob | Freelancer | `GET /api/v1/proposals/me/jobs/{jobId}/has-applied` | Yes | |
| ListProposalsByJob | Client | Not exposed directly | No | Closest is ListClientProposals |
| ListMyProposals | Freelancer | `GET /api/v1/proposals/me` | Yes | Supports `status`, `job_id`, `sort_by`, `page_size`, `page_token` |
| ListClientProposals | Client | `GET /api/v1/proposals/client` | Yes | Supports `status`, `job_id`, `freelancer_id`, `sort_by`, `page_size`, `page_token` |
| CountProposalsByJob | Client | `GET /api/v1/proposals/jobs/{jobId}/counts` | Yes | |
| CountClientProposalInbox | Client | `GET /api/v1/proposals/client/counts` | Yes | |
| GetProposalAttachmentUploadUrl | Freelancer | `POST /api/v1/proposals/{proposalId}/attachments/upload-url` | Yes | |
| GetProposalAttachmentDownloadUrl | Client/Freelancer owner | `GET /api/v1/proposals/{proposalId}/attachments/{attachmentId}/download-url` | Yes | |
| SetProposalStatus | Client | `POST /api/v1/proposals/{proposalId}/decision` | Yes | Decisions: `shortlisted`, `rejected` |
| InternalHireProposal | Internal Contract service | Not public | No | Triggered when a sent offer marks the selected proposal as `hired` |

## Sample request payloads

### Set proposal decision

`POST /api/v1/proposals/{proposalId}/decision`

```json
{
  "decision": "shortlisted",
  "reason": "Strong fit for the role"
}
```

### Get attachment upload URL

`POST /api/v1/proposals/{proposalId}/attachments/upload-url`

```json
{
  "file_name": "portfolio.pdf",
  "content_type": "application/pdf"
}
```

## Status rules summary

- Freelancer can modify only when proposal status is `sent`.
- Freelancer can withdraw when status is `sent` or `shortlisted`.
- Client decisions are `shortlisted` or `rejected`.
- Opening the offer composer does not change proposal state.
- Proposal transitions to `hired` only after Contract `CreateContract` succeeds.

## Optional gRPC testing notes

If you want Postman gRPC requests for gRPC-only RPCs (`SubmitProposal`, `ModifyProposal`, `WithdrawProposal`, `InternalHireProposal`), create a gRPC request tab in Postman and target `localhost:50054` with proto file `api/proto/proposal/v1/proposal.proto`.
