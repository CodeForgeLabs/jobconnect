create unique index if not exists uq_contracts_pending_offer_per_job_client
    on contracts (job_id, client_id)
    where status = 'pending_acceptance';
