-- Rollback: 20260615000002_add_missing_billing_indexes

-- Drop indexes
DROP INDEX IF EXISTS idx_usage_events_v2_tenant_timestamp;
DROP INDEX IF EXISTS idx_usage_events_v2_tenant_event_type;
DROP INDEX IF EXISTS idx_usage_events_v2_tenant_ai_model;
DROP INDEX IF EXISTS idx_cost_allocation_entries_timestamp;
DROP INDEX IF EXISTS idx_usage_rollups_v2_tenant_period;
DROP INDEX IF EXISTS idx_usage_rollups_v2_tenant_event_type;
DROP INDEX IF EXISTS idx_pending_usage_charges_tenant_status;
DROP INDEX IF EXISTS idx_pending_usage_charges_retry;

-- Drop CHECK constraint
ALTER TABLE invoices DROP CONSTRAINT IF EXISTS invoices_currency_check;

-- Restore FK without ON DELETE (or drop entirely)
DO $$
DECLARE
    constraint_name TEXT;
BEGIN
    SELECT conname INTO constraint_name
    FROM pg_constraint
    WHERE conrelid = 'invoices'::regclass
      AND confrelid = 'subscriptions'::regclass
      AND contype = 'f'
    LIMIT 1;

    IF constraint_name IS NOT NULL THEN
        EXECUTE format('ALTER TABLE invoices DROP CONSTRAINT %I', constraint_name);
        ALTER TABLE invoices
            ADD CONSTRAINT invoices_subscription_id_fkey
            FOREIGN KEY (subscription_id) REFERENCES subscriptions(id);
    END IF;
END
$$;
