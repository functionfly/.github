-- User date of birth (signup / profile), stored as DATE
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS date_of_birth DATE;

COMMENT ON COLUMN users.date_of_birth IS 'User date of birth; collected at signup for age verification';
