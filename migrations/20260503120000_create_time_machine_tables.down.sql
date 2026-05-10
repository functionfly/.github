-- Drop Time Machine tables in reverse dependency order

DROP TABLE IF EXISTS time_machine_audit_certificates CASCADE;
DROP TABLE IF EXISTS time_machine_reconciliations CASCADE;
DROP TABLE IF EXISTS time_machine_replay_items CASCADE;
DROP TABLE IF EXISTS time_machine_schedules CASCADE;
DROP TABLE IF EXISTS time_machine_replays CASCADE;
