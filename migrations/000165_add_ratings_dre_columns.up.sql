-- Add DRE v2 sub-score columns to registry_function_ratings (match RegistryFunctionRating struct)
-- +migrate Up
ALTER TABLE registry_function_ratings
ADD COLUMN IF NOT EXISTS determinism_score DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS replay_integrity_score DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS performance_stability_score DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS drift_score DOUBLE PRECISION NOT NULL DEFAULT 1,
ADD COLUMN IF NOT EXISTS trust_score_v2 DOUBLE PRECISION NOT NULL DEFAULT 0,
ADD COLUMN IF NOT EXISTS trust_v2_updated_at TIMESTAMP WITH TIME ZONE;
