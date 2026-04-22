alter table contract_amendments
    drop column if exists response_note,
    drop column if exists responded_by;

