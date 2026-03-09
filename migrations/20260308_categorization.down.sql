-- Migration: Remove categorization tables
-- Created: 2026-03-08

-- Drop foreign key constraint first
ALTER TABLE function_categories DROP CONSTRAINT IF EXISTS fk_function_categories_function;

-- Drop indexes
DROP INDEX IF EXISTS idx_function_categories_function_id;
DROP INDEX IF EXISTS idx_function_categories_primary_category;
DROP INDEX IF EXISTS idx_function_categories_secondary_category;
DROP INDEX IF EXISTS idx_function_categories_manually_edited;
DROP INDEX IF EXISTS idx_function_categories_tags;
DROP INDEX IF EXISTS idx_factory_versions_auto_category;

-- Drop table
DROP TABLE IF EXISTS function_categories;

-- Remove categorization columns from factory_versions
ALTER TABLE factory_versions
DROP COLUMN IF EXISTS auto_category,
DROP COLUMN IF EXISTS auto_tags,
DROP COLUMN IF EXISTS category_confidence;
