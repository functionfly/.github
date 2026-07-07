-- Slack Status Monitor Schema
-- Migration: 20260706120000_slack_status_monitor.down.sql

DROP TABLE IF EXISTS slack_alert_log;
DROP TABLE IF EXISTS monitored_components;
DROP TABLE IF EXISTS slack_config;

ALTER TABLE IF EXISTS notification_preferences DROP COLUMN IF EXISTS slack_enabled;
