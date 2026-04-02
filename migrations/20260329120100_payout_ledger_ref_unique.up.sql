-- One payout ledger credit per logical external reference (e.g. registry execution log id).
BEGIN;

CREATE UNIQUE INDEX IF NOT EXISTS idx_payout_ledger_ref_type_id_unique
    ON payout_ledger (reference_type, reference_id)
    WHERE reference_id IS NOT NULL;

COMMIT;
