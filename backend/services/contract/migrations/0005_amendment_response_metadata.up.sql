alter table contract_amendments
    add column if not exists responded_by uuid null,
    add column if not exists response_note text not null default '';

