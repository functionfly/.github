-- +migrate Down
ALTER TABLE registry_function_versions
DROP COLUMN IF EXISTS source_code,
DROP COLUMN IF EXISTS updated_at;
