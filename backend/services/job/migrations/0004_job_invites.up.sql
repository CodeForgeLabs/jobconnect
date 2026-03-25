create table if not exists job_invites (
    id bigserial primary key,
    job_id bigint not null references jobs(id) on delete cascade,
    client_id uuid not null,
    freelancer_id uuid not null,
    created_at timestamptz not null,
    unique (job_id, freelancer_id)
);

create index if not exists idx_job_invites_job_id_created_at on job_invites(job_id, created_at desc);
