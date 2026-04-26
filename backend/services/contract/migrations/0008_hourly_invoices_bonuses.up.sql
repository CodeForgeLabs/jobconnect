alter table contract_hourly_logs
    add column if not exists invoice_id bigint null;

create table if not exists contract_hourly_invoices (
    id bigserial primary key,
    contract_id bigint not null references contracts(id) on delete cascade,
    client_id uuid not null,
    freelancer_id uuid not null,
    week_start timestamptz not null,
    week_end timestamptz not null,
    status text not null check (status in ('draft', 'submitted', 'in_review', 'approved', 'disputed', 'charged', 'paid', 'failed')),
    billable_minutes integer not null default 0 check (billable_minutes >= 0),
    hourly_rate double precision not null default 0,
    amount_minor bigint not null default 0 check (amount_minor >= 0),
    dispute_id text not null default '',
    created_at timestamptz not null,
    submitted_at timestamptz null,
    approved_at timestamptz null,
    paid_at timestamptz null,
    failed_at timestamptz null,
    unique (contract_id, week_start)
);

create index if not exists idx_contract_hourly_invoices_contract_created
    on contract_hourly_invoices(contract_id, created_at desc);

create index if not exists idx_contract_hourly_logs_invoice
    on contract_hourly_logs(invoice_id);

create table if not exists contract_bonuses (
    id bigserial primary key,
    contract_id bigint not null references contracts(id) on delete cascade,
    client_id uuid not null,
    freelancer_id uuid not null,
    amount_minor bigint not null check (amount_minor > 0),
    note text not null default '',
    status text not null check (status in ('pending', 'paid', 'failed')),
    created_at timestamptz not null,
    paid_at timestamptz null,
    failed_at timestamptz null
);

create index if not exists idx_contract_bonuses_contract_created
    on contract_bonuses(contract_id, created_at desc);
