-- Rollback payout system migration.

BEGIN;

DROP TRIGGER IF EXISTS trg_payout_requests_updated_at ON payout_requests;
DROP TRIGGER IF EXISTS trg_stripe_connect_accounts_updated_at ON stripe_connect_accounts;
DROP FUNCTION IF EXISTS update_payout_updated_at();

DROP TABLE IF EXISTS payout_ledger;
DROP TABLE IF EXISTS payout_requests;
DROP TABLE IF EXISTS stripe_connect_accounts;

COMMIT;
