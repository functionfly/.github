-- Remove is_active from registry_function_versions
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS is_active;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS deactivated_at;
ALTER TABLE registry_function_versions DROP COLUMN IF EXISTS deactivated_by;

-- Remove is_flagged and flag_reason from registry_functions
ALTER TABLE registry_functions DROP COLUMN IF EXISTS flagged_by;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS flagged_at;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS flag_reason;
ALTER TABLE registry_functions DROP COLUMN IF EXISTS is_flagged;
