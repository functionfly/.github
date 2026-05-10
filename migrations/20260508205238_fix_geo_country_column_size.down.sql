-- Rollback: Revert geo_country back to VARCHAR(2)
ALTER TABLE registry_function_executions
    ALTER COLUMN geo_country TYPE VARCHAR(2);