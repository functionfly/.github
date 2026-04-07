-- Rollback: Support System Tables
-- Note: This is a placeholder down migration
-- Manual cleanup may be required for enum types

DROP TABLE IF EXISTS support_conversation_participants CASCADE;
DROP TABLE IF EXISTS support_conversation_handoffs CASCADE;
DROP TABLE IF EXISTS support_conversation_feedbacks CASCADE;
DROP TABLE IF EXISTS support_conversation_actions CASCADE;
DROP TABLE IF EXISTS support_messages CASCADE;
DROP TABLE IF EXISTS support_conversations CASCADE;
DROP TABLE IF EXISTS support_tickets CASCADE;
DROP TABLE IF EXISTS support_snippets CASCADE;

DROP TYPE IF EXISTS support_action_type_enum CASCADE;
DROP TYPE IF EXISTS support_feedback_type_enum CASCADE;
DROP TYPE IF EXISTS support_source_enum CASCADE;
DROP TYPE IF EXISTS support_priority_enum CASCADE;
DROP TYPE IF EXISTS support_status_enum CASCADE;
DROP TYPE IF EXISTS support_conversation_type_enum CASCADE;
