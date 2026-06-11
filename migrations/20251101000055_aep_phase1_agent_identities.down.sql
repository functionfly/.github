-- Rollback AEP Phase 1: Agent Identity and Quota Configuration
DROP TABLE IF EXISTS agent_quota_configs;
DROP TABLE IF EXISTS agent_identities;
