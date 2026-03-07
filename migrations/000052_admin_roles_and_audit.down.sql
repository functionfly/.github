-- Rollback admin roles and audit system
DROP INDEX IF EXISTS idx_audit_events_action;
DROP INDEX IF EXISTS idx_audit_events_timestamp;
DROP INDEX IF EXISTS idx_audit_events_resource_type_id;
DROP INDEX IF EXISTS idx_audit_events_tenant_id;
DROP INDEX IF EXISTS idx_audit_events_actor_user_id;

DROP TABLE IF EXISTS pricing_tiers;
DROP TABLE IF EXISTS audit_events;

ALTER TABLE users DROP COLUMN IF EXISTS role;