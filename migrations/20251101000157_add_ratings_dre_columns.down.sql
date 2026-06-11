-- +migrate Down
ALTER TABLE registry_function_ratings
DROP COLUMN IF EXISTS determinism_score,
DROP COLUMN IF EXISTS replay_integrity_score,
DROP COLUMN IF EXISTS performance_stability_score,
DROP COLUMN IF EXISTS drift_score,
DROP COLUMN IF EXISTS trust_score_v2,
DROP COLUMN IF EXISTS trust_v2_updated_at;
