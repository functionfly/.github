-- Add company_name column to users (nullable)
ALTER TABLE users ADD COLUMN IF NOT EXISTS company_name VARCHAR(255);
