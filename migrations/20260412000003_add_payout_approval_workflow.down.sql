-- Down migration: Remove payout approval workflow
DROP TABLE IF EXISTS payout_approval_audit;
DROP TABLE IF EXISTS payout_approval_rules;

ALTER TABLE payout_requests DROP COLUMN IF EXISTS requires_approval;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approval_threshold_usd;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approved_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approved_at;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS approval_notes;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS second_approval_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS second_approval_at;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejection_reason;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejected_by;
ALTER TABLE payout_requests DROP COLUMN IF EXISTS rejected_at;
