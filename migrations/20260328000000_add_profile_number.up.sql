-- Add profile_number column to users table for early adopter tracking
-- This column stores a sequential number assigned to users based on registration order

ALTER TABLE users ADD COLUMN IF NOT EXISTS profile_number INTEGER;

-- Create unique index on profile_number
CREATE UNIQUE INDEX IF NOT EXISTS idx_users_profile_number ON users(profile_number);

-- Populate profile_number for existing users based on created_at order
-- This assigns sequential numbers starting from 1 for the earliest users
WITH ranked_users AS (
    SELECT id, ROW_NUMBER() OVER (ORDER BY created_at ASC) as rn
    FROM users
    WHERE profile_number IS NULL
)
UPDATE users u
SET profile_number = ru.rn
FROM ranked_users ru
WHERE u.id = ru.id;

-- Create a sequence for new users to automatically get the next profile number
-- Start from the highest existing profile_number + 1, or 1 if no users exist
DO $$
DECLARE
    max_num INTEGER;
BEGIN
    SELECT COALESCE(MAX(profile_number), 0) + 1 INTO max_num FROM users;
    EXECUTE format('CREATE SEQUENCE IF NOT EXISTS users_profile_number_seq START WITH %s INCREMENT BY 1', max_num);
END $$;
