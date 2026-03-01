-- Rollback StateFabric tables

-- Drop in reverse order to handle dependencies

-- Drop usage metrics
DROP TABLE IF EXISTS state_usage_metrics;

-- Drop agent memory indexes
DROP TABLE IF EXISTS agent_memory_indexes;

-- Drop agent memories (will cascade from state_id if exists)
DROP TABLE IF EXISTS agent_memories;

-- Drop state triggers
DROP TABLE IF EXISTS state_triggers;

-- Drop enums
DROP TYPE IF EXISTS trigger_type;
DROP TYPE IF EXISTS memory_type;

-- Drop state permissions
DROP TABLE IF EXISTS state_permissions;

-- Drop state snapshots
DROP TABLE IF EXISTS state_snapshots;

-- Drop state events
DROP TABLE IF EXISTS state_events;

-- Drop state values
DROP TABLE IF EXISTS state_values;

-- Drop states
DROP TABLE IF EXISTS states;
