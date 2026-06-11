-- Rollback function recommendations system
-- This drops the trigger, functions, and tables created in 000050_create_recommendations.up.sql

-- Drop trigger first (depends on function)
DROP TRIGGER IF EXISTS trg_track_session_co_occurrence ON function_execution_events;

-- Drop functions (order matters due to dependencies)
DROP FUNCTION IF EXISTS track_session_co_occurrence();
DROP FUNCTION IF EXISTS update_co_occurrence(UUID, UUID);
DROP FUNCTION IF EXISTS get_related_functions(UUID, INTEGER, NUMERIC);
DROP FUNCTION IF EXISTS compute_content_similarity(UUID, UUID);

-- Drop tables (order matters due to foreign keys)
DROP TABLE IF EXISTS function_execution_events;
DROP TABLE IF EXISTS recommendation_feedback;
DROP TABLE IF EXISTS function_embeddings;
DROP TABLE IF EXISTS category_similarities;
DROP TABLE IF EXISTS session_function_usage;
DROP TABLE IF EXISTS function_recommendations;
DROP TABLE IF EXISTS user_function_interactions;
DROP TABLE IF EXISTS function_similarities;
DROP TABLE IF EXISTS function_co_occurrences;
