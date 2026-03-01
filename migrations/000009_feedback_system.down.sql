-- Drop Feedback System tables and indexes
DROP TRIGGER IF EXISTS update_feedback_updated_at ON feedback;
DROP INDEX IF EXISTS idx_feedback_attachments_feedback_id;
DROP INDEX IF EXISTS idx_feedback_created_at;
DROP INDEX IF EXISTS idx_feedback_status;
DROP INDEX IF EXISTS idx_feedback_type;
DROP INDEX IF EXISTS idx_feedback_user_email;
DROP INDEX IF EXISTS idx_feedback_user_id;
DROP TABLE IF EXISTS feedback_attachments;
DROP TABLE IF EXISTS feedback;