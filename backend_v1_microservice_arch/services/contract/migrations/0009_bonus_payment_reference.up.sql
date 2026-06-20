alter table contract_bonuses
    add column if not exists payment_reference_id text not null default '';

create index if not exists idx_contract_bonuses_payment_reference
    on contract_bonuses(payment_reference_id);
