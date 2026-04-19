-- Sunsetting Flywheel Network: drop all flywheel tables and dependent objects
-- Using CASCADE to handle FK constraints from conversations, flywheel_messages, flywheel_replays

DROP TABLE IF EXISTS flywheel_subscriptions CASCADE;
DROP TABLE IF EXISTS flywheel_executions CASCADE;
DROP TABLE IF EXISTS flywheel_agent_collaborations CASCADE;
DROP TABLE IF EXISTS flywheel_challenge_submissions CASCADE;
DROP TABLE IF EXISTS flywheel_challenges CASCADE;
DROP TABLE IF EXISTS flywheel_reputation_events CASCADE;
DROP TABLE IF EXISTS flywheel_reputation_scores CASCADE;
DROP TABLE IF EXISTS flywheel_replies CASCADE;
DROP TABLE IF EXISTS flywheel_threads CASCADE;
DROP TABLE IF EXISTS flywheel_categories CASCADE;
