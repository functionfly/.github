-- Add index on github_imports.connection_id for faster webhook sync queries
CREATE INDEX IF NOT EXISTS idx_github_imports_connection_id ON github_imports(connection_id);