## Proposal service (gRPC)

### MVP scope implemented in this phase
- Submit proposal
- Modify proposal
- Withdraw proposal
- View proposal status
- Client view/list proposals
- Client shortlist/reject/hire transitions
- Filter/sort proposal lists

### Planned runtime dependencies
- Postgres database for proposal persistence
- Job service gRPC read client for job existence/open-state checks
- Shared JWT secret for access-token parsing

### Environment variables
- `PROPOSAL_GRPC_LISTEN_ADDR` (default `:50054`)
- `PROPOSAL_POSTGRES_URL` (required)
- `AUTH_JWT_SECRET` (required)
- `JOB_SERVICE_GRPC_ADDR` (required)
