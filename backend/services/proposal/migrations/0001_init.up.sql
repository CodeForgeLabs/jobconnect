create table if not exists proposals (
    id bigserial primary key,
    job_id bigint not null,
    client_id uuid not null,
    freelancer_id uuid not null,
    cover_letter text not null,
    bid_type text not null check (bid_type in ('fixed', 'hourly')),
    bid_amount double precision not null check (bid_amount > 0),
    estimated_days integer not null check (estimated_days > 0),
    status text not null check (status in ('sent', 'shortlisted', 'rejected', 'hired', 'withdrawn')),
    status_reason text not null default '',
    created_at timestamptz not null,
    updated_at timestamptz not null,
    shortlisted_at timestamptz null,
    rejected_at timestamptz null,
    hired_at timestamptz null,
    withdrawn_at timestamptz null
);

create table if not exists proposal_attachments (
    id bigserial primary key,
    proposal_id bigint not null references proposals(id) on delete cascade,
    file_name text not null,
    content_type text not null,
    url text not null,
    size_bytes bigint not null default 0
);

create index if not exists idx_proposals_job_created_at on proposals(job_id, created_at desc);
create index if not exists idx_proposals_job_status_created_at on proposals(job_id, status, created_at desc);
create index if not exists idx_proposals_job_bid_desc on proposals(job_id, bid_amount desc, created_at desc);
create index if not exists idx_proposals_job_bid_asc on proposals(job_id, bid_amount asc, created_at desc);
create index if not exists idx_proposals_freelancer_created_at on proposals(freelancer_id, created_at desc);
create index if not exists idx_proposals_freelancer_status_created_at on proposals(freelancer_id, status, created_at desc);

create unique index if not exists uq_proposals_active_per_job_freelancer
    on proposals(job_id, freelancer_id)
    where status in ('sent', 'shortlisted');

create index if not exists idx_proposal_attachments_proposal_id on proposal_attachments(proposal_id, id);
