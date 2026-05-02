DROP TRIGGER IF EXISTS spend_caps_updated_at ON spend_caps;
DROP TRIGGER IF EXISTS usage_alerts_updated_at ON usage_alerts;
DROP FUNCTION IF EXISTS update_spend_caps_updated_at();
DROP FUNCTION IF EXISTS update_usage_alerts_updated_at();
DROP TABLE IF EXISTS usage_trends;
DROP TABLE IF EXISTS usage_forecasts;
DROP TABLE IF EXISTS spend_caps;
DROP TABLE IF EXISTS usage_alert_history;
DROP TABLE IF EXISTS usage_alerts;
