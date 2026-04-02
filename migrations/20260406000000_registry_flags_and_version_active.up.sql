-- Add is_flagged and flag_reason to registry_functions for admin moderation
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS is_flagged BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS flag_reason TEXT;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS flagged_at TIMESTAMPTZ;
ALTER TABLE registry_functions ADD COLUMN IF NOT EXISTS flagged_by UUID;

CREATE INDEX IF NOT EXISTS idx_registry_functions_is_flagged ON registry_functions(is_flagged);

-- Add is_active to registry_function_versions for version lifecycle management
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS is_active BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS deactivated_at TIMESTAMPTZ;
ALTER TABLE registry_function_versions ADD COLUMN IF NOT EXISTS deactivated_by UUID;

CREATE INDEX IF NOT EXISTS idx_registry_function_versions_is_active ON registry_function_versions(is_active);
