-- Migration: Drop analytics churn tables

DROP TRIGGER IF EXISTS trigger_update_churn_event_updated_at ON subscription_churn_events;
DROP FUNCTION IF EXISTS update_churn_event_updated_at();

DROP TABLE IF EXISTS subscription_churn_events;
DROP TABLE IF EXISTS revenue_recognition;
DROP TABLE IF EXISTS mrr_snapshots;
DROP TABLE IF EXISTS cohort_analysis;
DROP TABLE IF EXISTS failed_payment_analytics;