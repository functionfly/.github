-- Link local invoice rows to Stripe Checkout / Invoice for idempotency and receipts.
ALTER TABLE invoices
    ADD COLUMN IF NOT EXISTS stripe_invoice_id TEXT,
    ADD COLUMN IF NOT EXISTS external_reference TEXT;

CREATE UNIQUE INDEX IF NOT EXISTS idx_invoices_external_reference_unique
    ON invoices (external_reference)
    WHERE external_reference IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_invoices_stripe_invoice_id
    ON invoices (stripe_invoice_id)
    WHERE stripe_invoice_id IS NOT NULL;
