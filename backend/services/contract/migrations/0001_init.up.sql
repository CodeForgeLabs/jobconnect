create table if not exists contracts (
    id bigserial primary key,
    client_id uuid not null,
    freelancer_id uuid not null,
    job_id bigint not null,
    proposal_id bigint null,
    contract_type text not null check (contract_type in ('fixed', 'hourly')),
    status text not null check (status in ('pending_acceptance', 'active', 'declined', 'paused', 'ended')),
    title text not null,
    description text not null default '',
    currency varchar(8) not null default 'USD',
    hourly_rate double precision not null default 0,
    fixed_total double precision not null default 0,
    weekly_hour_limit integer not null default 0,
    milestones jsonb not null default '[]'::jsonb,
    created_at timestamptz not null,
    updated_at timestamptz not null,
    activated_at timestamptz null,
    declined_at timestamptz null,
    constraint contracts_amount_by_type check (
        (contract_type = 'fixed' and fixed_total > 0)
        or
        (contract_type = 'hourly' and hourly_rate > 0)
    )
);

create index if not exists idx_contracts_client_created_at on contracts(client_id, created_at desc);
create index if not exists idx_contracts_freelancer_created_at on contracts(freelancer_id, created_at desc);
create index if not exists idx_contracts_client_status_created_at on contracts(client_id, status, created_at desc);
create index if not exists idx_contracts_freelancer_status_created_at on contracts(freelancer_id, status, created_at desc);
create index if not exists idx_contracts_job_id on contracts(job_id);
create unique index if not exists uq_contracts_proposal_id_nonzero on contracts(proposal_id) where proposal_id is not null;
