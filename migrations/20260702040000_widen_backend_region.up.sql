-- Widen backends.region from VARCHAR(10) to VARCHAR(100)
-- The previous constraint was too small for real region strings like 'eu-central-1' (12 chars)
ALTER TABLE backends ALTER COLUMN region TYPE VARCHAR(100);
