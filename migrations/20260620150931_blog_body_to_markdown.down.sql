-- Rollback: This migration converts JSON to Markdown, which is lossy.
-- The reverse would require TipTap-format JSON, which we no longer have.
-- Leave as a no-op since down-migration cannot be safely automated.
SELECT 1;
