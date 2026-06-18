-- Remove TTL fields from state_fabrics table

ALTER TABLE state_fabrics DROP COLUMN IF EXISTS expires_at;
ALTER TABLE state_fabrics DROP COLUMN IF EXISTS ttl_days;
