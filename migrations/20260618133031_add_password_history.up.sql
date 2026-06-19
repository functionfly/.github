-- Create password history table to prevent password reuse
CREATE TABLE IF NOT EXISTS password_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for fast lookup by user_id
CREATE INDEX IF NOT EXISTS idx_password_history_user_id ON password_history(user_id);

-- Index for cleanup queries
CREATE INDEX IF NOT EXISTS idx_password_history_created_at ON password_history(created_at);

-- Function to automatically add password to history when password is changed
CREATE OR REPLACE FUNCTION add_password_to_history()
RETURNS TRIGGER AS $$
BEGIN
    IF OLD.password_hash IS DISTINCT FROM NEW.password_hash THEN
        INSERT INTO password_history (user_id, password_hash, created_at)
        VALUES (OLD.id, OLD.password_hash, NOW());
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically add password to history on user update
DROP TRIGGER IF EXISTS trigger_add_password_to_history ON users;
CREATE TRIGGER trigger_add_password_to_history
    AFTER UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION add_password_to_history();
