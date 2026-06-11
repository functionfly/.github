-- Rollback: Support System Tables
-- Order matters: drop triggers first, then tables, then function, then types

-- Drop triggers (depend on update_updated_at_column function)
DROP TRIGGER IF EXISTS update_emergency_requests_updated_at ON emergency_fix_requests;
DROP TRIGGER IF EXISTS update_staff_availability_updated_at ON staff_availability;
DROP TRIGGER IF EXISTS update_support_conversations_updated_at ON support_conversations;

-- Drop tables (order matters due to FK dependencies)
DROP TABLE IF EXISTS support_conversation_participants CASCADE;
DROP TABLE IF EXISTS emergency_fix_requests CASCADE;
DROP TABLE IF EXISTS support_messages CASCADE;
DROP TABLE IF EXISTS support_conversations CASCADE;
DROP TABLE IF EXISTS staff_availability CASCADE;

-- Drop trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop enum types (in reverse order of creation)
DROP TYPE IF EXISTS support_message_type_enum CASCADE;
DROP TYPE IF EXISTS support_author_type_enum CASCADE;
DROP TYPE IF EXISTS support_priority_enum CASCADE;
DROP TYPE IF EXISTS support_status_enum CASCADE;
DROP TYPE IF EXISTS support_conversation_type_enum CASCADE;
