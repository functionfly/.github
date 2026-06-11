-- Rollback: Remove reputation farming alerts and trust score weights config tables
-- Created: 20260608190000

DROP TABLE IF EXISTS reputation_events;
DROP TABLE IF EXISTS trust_score_weights_config;
DROP TABLE IF EXISTS reputation_farming_alerts;
