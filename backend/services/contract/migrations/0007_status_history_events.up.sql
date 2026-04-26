alter table contract_status_history
    add column if not exists event_type text not null default '',
    add column if not exists milestone_id bigint null;

