-- Add 6-digit verification code for email verification (type-in flow)
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_code VARCHAR(10);

CREATE INDEX IF NOT EXISTS idx_users_verification_code ON users(verification_code) WHERE verification_code IS NOT NULL;
