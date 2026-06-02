drop index if exists uq_proposals_active_per_job_freelancer;

create unique index if not exists uq_proposals_active_per_job_freelancer
    on proposals(job_id, freelancer_id)
    where status in ('sent', 'shortlisted');

alter table proposals
    drop constraint if exists proposals_status_check;

alter table proposals
    add constraint proposals_status_check
    check (status in ('sent', 'shortlisted', 'rejected', 'hired', 'withdrawn'));

alter table proposals
    drop column if exists offer_sent_at;
