-- Drop Flywheel Network Database Schema
-- This migration removes all tables for the Flywheel Network feature

-- Drop tables in reverse order to handle foreign key constraints

DROP TABLE IF EXISTS suspensions;
DROP TABLE IF EXISTS abuse_tracking;
DROP TABLE IF EXISTS flywheel_replays;
DROP TABLE IF EXISTS agent_reputations;
DROP TABLE IF EXISTS challenge_submissions;
DROP TABLE IF EXISTS flywheel_messages;
DROP TABLE IF EXISTS flywheel_threads;
DROP TABLE IF EXISTS challenges;
DROP TABLE IF EXISTS debates;
DROP TABLE IF EXISTS agent_attachments;
DROP TABLE IF EXISTS reputation_profiles;
DROP TABLE IF EXISTS flywheel_solutions;
DROP TABLE IF EXISTS flywheel_problems;
