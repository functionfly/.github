-- DNA Mutation Payment Status Tracking
-- Migration: 20260608010000_dna_payment_status
-- Purpose: Track payment status for mutations to handle wallet debit failures
-- Also add new mutation statuses for payment flow

-- First, drop the old check constraint and recreate with new statuses
ALTER TABLE function_dna_mutations DROP CONSTRAINT IF EXISTS function_dna_mutations_status_check;

ALTER TABLE function_dna_mutations
ALTER COLUMN status TYPE TEXT,
ADD CONSTRAINT function_dna_mutations_status_check
CHECK (status IN (
    'proposed', 'accepted', 'rejected', 'deploying', 'deployed', 'rolled_back',
    'accepted_pending_payment', 'payment_failed'
));

-- Add payment tracking columns
ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS payment_status TEXT NOT NULL DEFAULT 'completed'
CHECK (payment_status IN ('pending', 'completed', 'failed', 'reconciled'));

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS payment_retry_count INT NOT NULL DEFAULT 0;

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS payment_failed_at TIMESTAMPTZ;

ALTER TABLE function_dna_mutations
ADD COLUMN IF NOT EXISTS payment_failure_reason TEXT;

-- Index for finding pending payments that need reconciliation
CREATE INDEX IF NOT EXISTS idx_dna_mutations_pending_payment
ON function_dna_mutations(tenant_id, payment_status, created_at)
WHERE payment_status = 'pending';

-- Index for retry processing
CREATE INDEX IF NOT EXISTS idx_dna_mutations_payment_retry
ON function_dna_mutations(payment_status, payment_retry_count)
WHERE payment_status = 'pending';