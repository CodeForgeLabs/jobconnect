drop table if exists contract_amendments;
drop table if exists contract_hourly_logs;
drop table if exists contract_status_history;
alter table contracts
    drop column if exists ended_at,
    drop column if exists paused_at;
