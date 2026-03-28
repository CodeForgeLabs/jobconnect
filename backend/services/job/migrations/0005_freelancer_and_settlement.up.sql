alter table job_invites
    add column if not exists response_status text not null default 'unspecified',
    add column if not exists responded_at timestamptz null;

alter table job_invites
    drop constraint if exists job_invites_response_status_check;

alter table job_invites
    add constraint job_invites_response_status_check
    check (response_status in ('unspecified', 'accepted', 'declined'));

create table if not exists saved_jobs (
    id bigserial primary key,
    job_id bigint not null references jobs(id) on delete cascade,
    freelancer_id uuid not null,
    created_at timestamptz not null,
    unique (job_id, freelancer_id)
);

create index if not exists idx_saved_jobs_freelancer_created_at on saved_jobs(freelancer_id, created_at desc);

alter table jobs
    add column if not exists completed_at timestamptz null,
    add column if not exists canceled_at timestamptz null,
    add column if not exists settlement_policy text not null default '',
    add column if not exists cancellation_reason text not null default '';
