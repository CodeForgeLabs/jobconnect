alter table contracts
    add column if not exists paused_at timestamptz null,
    add column if not exists ended_at timestamptz null;

create table if not exists contract_status_history (
    id bigserial primary key,
    contract_id bigint not null references contracts(id) on delete cascade,
    status text not null,
    reason text not null default '',
    actor_id uuid not null,
    created_at timestamptz not null
);

create index if not exists idx_contract_status_history_contract_created
    on contract_status_history(contract_id, created_at desc);

create table if not exists contract_hourly_logs (
    id bigserial primary key,
    contract_id bigint not null references contracts(id) on delete cascade,
    freelancer_id uuid not null,
    work_date date not null,
    start_at timestamptz not null,
    end_at timestamptz not null,
    duration_min integer not null check (duration_min > 0),
    note text not null default '',
    status text not null check (status in ('pending', 'approved', 'rejected')),
    review_note text not null default '',
    created_at timestamptz not null,
    client_review_at timestamptz null
);

create index if not exists idx_contract_hourly_logs_contract_created
    on contract_hourly_logs(contract_id, created_at desc);

create table if not exists contract_amendments (
    id bigserial primary key,
    contract_id bigint not null references contracts(id) on delete cascade,
    proposed_by uuid not null,
    summary text not null,
    payload_json jsonb not null default '{}'::jsonb,
    status text not null check (status in ('pending', 'accepted', 'rejected', 'expired')),
    expires_at timestamptz null,
    responded_at timestamptz null,
    created_at timestamptz not null
);

create index if not exists idx_contract_amendments_contract_created
    on contract_amendments(contract_id, created_at desc);
