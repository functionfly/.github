-- Add function version changelog table
-- This table tracks changelogs for each version of a function

CREATE TABLE IF NOT EXISTS function_version_changelogs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL,
    function_version_id UUID NOT NULL,
    version VARCHAR(50) NOT NULL,
    previous_version TEXT,
    change_type VARCHAR(50) NOT NULL,
    category VARCHAR(50) NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    changes JSONB DEFAULT '[]',
    author TEXT NOT NULL,
    author_id UUID,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Index for querying changelogs by function
CREATE INDEX IF NOT EXISTS idx_function_version_changelogs_function_id
ON function_version_changelogs(function_id);

-- Index for querying changelogs by version
CREATE INDEX IF NOT EXISTS idx_function_version_changelogs_version
ON function_version_changelogs(version);

-- Index for querying changelogs by function version ID
CREATE INDEX IF NOT EXISTS idx_function_version_changelogs_function_version_id
ON function_version_changelogs(function_version_id);

-- Index for querying changelogs by category
CREATE INDEX IF NOT EXISTS idx_function_version_changelogs_category
ON function_version_changelogs(category);

-- Foreign key to function versions
ALTER TABLE function_version_changelogs
ADD CONSTRAINT fk_function_version_changelogs_function_version
FOREIGN KEY (function_version_id)
REFERENCES registry_function_versions(id)
ON DELETE CASCADE;
