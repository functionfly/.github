-- Browser Automation Migration - Down
-- Drops all browser automation tables

DROP TABLE IF EXISTS agent_browser_dead_letters;
DROP TABLE IF EXISTS agent_browser_credentials;
DROP TABLE IF EXISTS agent_browser_usage;
DROP TABLE IF EXISTS agent_browser_sessions;
DROP TABLE IF EXISTS agent_browser_configs;
