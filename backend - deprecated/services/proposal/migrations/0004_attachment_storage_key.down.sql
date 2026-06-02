drop index if exists idx_proposal_attachments_storage_key;

alter table proposal_attachments
    drop column if exists storage_key;
