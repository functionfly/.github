-- Migration to drop cost_allocation_entries table
-- Reverses the detailed cost tracking implementation

DROP TABLE IF EXISTS cost_allocation_reports;
DROP TABLE IF EXISTS cost_allocation_entries;
