-- Migration: Add categorization tables for auto-categorization and intelligent tagging
-- Created: 2026-03-08

-- Create function_categories table to store categorization results
CREATE TABLE IF NOT EXISTS function_categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL UNIQUE,
    primary_category TEXT NOT NULL,
    secondary_category TEXT,
    tags TEXT[] DEFAULT '{}',
    confidence DECIMAL(5,4) NOT NULL DEFAULT 0,
    reasoning TEXT,
    code_patterns TEXT[] DEFAULT '{}',
    input_types TEXT[] DEFAULT '{}',
    output_types TEXT[] DEFAULT '{}',
    manually_edited BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for efficient querying
CREATE INDEX IF NOT EXISTS idx_function_categories_function_id ON function_categories(function_id);
CREATE INDEX IF NOT EXISTS idx_function_categories_primary_category ON function_categories(primary_category);
CREATE INDEX IF NOT EXISTS idx_function_categories_secondary_category ON function_categories(secondary_category);
CREATE INDEX IF NOT EXISTS idx_function_categories_manually_edited ON function_categories(manually_edited);

-- Create GIN index for tags array
CREATE INDEX IF NOT EXISTS idx_function_categories_tags ON function_categories USING GIN(tags);

-- Add foreign key constraint to registry_functions
ALTER TABLE function_categories
ADD CONSTRAINT fk_function_categories_function
FOREIGN KEY (function_id) REFERENCES registry_functions(id) ON DELETE CASCADE;

-- Add categorization metadata columns to factory_versions if table exists
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'factory_versions') THEN
    ALTER TABLE factory_versions
      ADD COLUMN IF NOT EXISTS auto_category TEXT,
      ADD COLUMN IF NOT EXISTS auto_tags TEXT[] DEFAULT '{}',
      ADD COLUMN IF NOT EXISTS category_confidence DECIMAL(5,4) DEFAULT 0;
    CREATE INDEX IF NOT EXISTS idx_factory_versions_auto_category ON factory_versions(auto_category);
  END IF;
END $$;

-- Add comment to table
COMMENT ON TABLE function_categories IS 'Stores auto-categorization results for factory-generated functions';
COMMENT ON COLUMN function_categories.primary_category IS 'Primary category ID from taxonomy';
COMMENT ON COLUMN function_categories.secondary_category IS 'Optional secondary category ID';
COMMENT ON COLUMN function_categories.tags IS 'Auto-extracted tags based on code analysis';
COMMENT ON COLUMN function_categories.confidence IS 'Confidence score for categorization (0-1)';
COMMENT ON COLUMN function_categories.reasoning IS 'Explanation of categorization decision';
COMMENT ON COLUMN function_categories.code_patterns IS 'Detected code patterns used for categorization';
COMMENT ON COLUMN function_categories.manually_edited IS 'Whether categorization was manually overridden';
