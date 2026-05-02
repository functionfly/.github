-- Reverse payout system enhancements

BEGIN;

DROP TABLE IF EXISTS payout_fee_deductions CASCADE;
DROP TABLE IF EXISTS payout_fee_config CASCADE;
DROP TABLE IF EXISTS payout_velocity_tracking CASCADE;
DROP TABLE IF EXISTS payout_schedule_preferences CASCADE;

-- Remove added columns from payout_requests
ALTER TABLE payout_requests DROP COLUMN IF EXISTS requires_approval;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approval_threshold_usd;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approved_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approved_at;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approval_notes;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS second_approval_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS second_approval_at;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejected_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejected_at;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS metadata;

COMMIT;
