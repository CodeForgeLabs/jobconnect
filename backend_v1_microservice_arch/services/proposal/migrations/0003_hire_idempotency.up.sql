create table if not exists proposal_hire_requests (
    id bigserial primary key,
    client_id uuid not null,
    proposal_id bigint not null references proposals(id) on delete cascade,
    request_id text not null,
    created_at timestamptz not null,
    unique (client_id, proposal_id, request_id)
);

create unique index if not exists uq_proposals_single_hired_per_job
    on proposals(job_id)
    where status = 'hired';
