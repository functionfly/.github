-- Drop trigger and function
DROP TRIGGER IF EXISTS trigger_add_password_to_history ON users;
DROP FUNCTION IF EXISTS add_password_to_history();

-- Drop indexes
DROP INDEX IF EXISTS idx_password_history_created_at;
DROP INDEX IF EXISTS idx_password_history_user_id;

-- Drop table
DROP TABLE IF EXISTS password_history;
