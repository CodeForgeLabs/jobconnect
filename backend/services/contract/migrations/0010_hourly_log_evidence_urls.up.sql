alter table contract_hourly_logs
    add column if not exists evidence_urls text[] not null default '{}';

