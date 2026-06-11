-- Rollback: Remove embed_config and embed_origin columns

DROP INDEX IF EXISTS idx_registry_function_executions_embed_origin;

ALTER TABLE registry_function_executions
    DROP COLUMN IF EXISTS embed_origin;

ALTER TABLE registry_functions
    DROP COLUMN IF EXISTS embed_config;
