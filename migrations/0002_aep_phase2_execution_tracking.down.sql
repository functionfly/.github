-- Rollback AEP Phase 2: Behavioral Policies, Execution Records, Sessions
DROP TABLE IF EXISTS agent_sessions;
DROP TABLE IF EXISTS agent_execution_records;
DROP TABLE IF EXISTS agent_behavioral_policies;
