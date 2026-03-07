create table if not exists jobs (
    id bigserial primary key,
    client_id uuid not null,
    title text not null,
    description text not null,
    required_skills jsonb not null default '[]'::jsonb,
    job_type text not null check (job_type in ('fixed', 'hourly')),
    budget_fixed double precision not null default 0,
    hourly_rate double precision not null default 0,
    currency varchar(8) not null default 'USD',
    deadline timestamptz null,
    status text not null check (status in ('open', 'closed')) default 'open',
    close_reason text not null default '',
    created_at timestamptz not null,
    updated_at timestamptz not null,
    closed_at timestamptz null,
    constraint jobs_budget_by_type check (
        (job_type = 'fixed' and budget_fixed > 0 and hourly_rate = 0)
        or
        (job_type = 'hourly' and hourly_rate > 0 and budget_fixed = 0)
    )
);

create index if not exists idx_jobs_client_id_created_at on jobs(client_id, created_at desc);
create index if not exists idx_jobs_status_created_at on jobs(status, created_at desc);

create table if not exists job_attachments (
    id bigserial primary key,
    job_id bigint not null references jobs(id) on delete cascade,
    file_name text not null,
    content_type text not null,
    url text not null,
    size_bytes bigint not null default 0
);

create index if not exists idx_job_attachments_job_id on job_attachments(job_id, id);
