drop table if exists contract_bonuses;
drop index if exists idx_contract_hourly_logs_invoice;
drop table if exists contract_hourly_invoices;
alter table contract_hourly_logs
    drop column if exists invoice_id;
