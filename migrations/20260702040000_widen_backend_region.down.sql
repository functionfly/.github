-- Revert backends.region to VARCHAR(10)
-- NOTE: This will fail if any rows have region values longer than 10 characters
ALTER TABLE backends ALTER COLUMN region TYPE VARCHAR(10);
