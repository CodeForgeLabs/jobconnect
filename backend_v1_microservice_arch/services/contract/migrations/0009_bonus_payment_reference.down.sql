drop index if exists idx_contract_bonuses_payment_reference;

alter table contract_bonuses
    drop column if exists payment_reference_id;
