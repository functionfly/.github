ALTER TABLE feedback DROP CONSTRAINT IF EXISTS feedback_feedback_type_check;
ALTER TABLE feedback ADD CONSTRAINT feedback_feedback_type_check CHECK (
  feedback_type IN ('bug', 'feature', 'improvement', 'general', 'launch_waitlist')
);
