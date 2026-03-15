DROP INDEX IF EXISTS wallet_ledger_entries_wallet_created_idx;
DROP INDEX IF EXISTS wallet_holds_status_idx;
DROP INDEX IF EXISTS wallet_holds_wallet_id_idx;

DROP TABLE IF EXISTS wallet_ledger_entries;
DROP TABLE IF EXISTS wallet_holds;
DROP TABLE IF EXISTS wallet_accounts;
