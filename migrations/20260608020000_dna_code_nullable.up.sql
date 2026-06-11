-- Make DNA mutation code columns nullable for security
-- Instead of storing full AI-generated code, we store only hashes
-- Code can be retrieved from the function registry when needed

ALTER TABLE function_dna_mutations ALTER COLUMN original_code DROP NOT NULL;
ALTER TABLE function_dna_mutations ALTER COLUMN mutated_code DROP NOT NULL;
