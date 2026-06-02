alter table contract_status_history
    drop column if exists milestone_id,
    drop column if exists event_type;

