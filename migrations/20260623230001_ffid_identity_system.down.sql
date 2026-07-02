-- Remove identity columns from employees
ALTER TABLE employees DROP COLUMN IF EXISTS trust_score;
ALTER TABLE employees DROP COLUMN IF EXISTS reputation_total;
ALTER TABLE employees DROP COLUMN IF EXISTS identity_signature;
ALTER TABLE employees DROP COLUMN IF EXISTS clearance_level_num;

-- Drop tables
DROP TABLE IF EXISTS reputation_history;
DROP TABLE IF EXISTS career_timeline_events;
DROP TABLE IF EXISTS achievement_progress;
DROP TABLE IF EXISTS achievement_definitions;
