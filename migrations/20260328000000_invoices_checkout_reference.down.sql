DROP INDEX IF EXISTS idx_invoices_stripe_invoice_id;
DROP INDEX IF EXISTS idx_invoices_external_reference_unique;
ALTER TABLE invoices DROP COLUMN IF EXISTS external_reference;
ALTER TABLE invoices DROP COLUMN IF EXISTS stripe_invoice_id;
