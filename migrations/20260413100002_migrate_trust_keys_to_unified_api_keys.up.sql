-- =====================================================
-- Migrate existing trust_api_keys to unified api_keys
-- =====================================================

-- Copy all trust_api_keys into the unified api_keys table
-- This maps the Trust API key model to the platform API key model

INSERT INTO api_keys (
    id,
    tenant_id,
    user_id,
    name,
    description,
    key_type,
    key_id,
    key_prefix,
    key_hash,
    key_version,
    expires_at,
    last_rotated_at,
    rotation_frequency_days,
    rate_limit_rpm,
    rate_limit_rph,
    rate_limit_rpd,
    is_active,
    metadata,
    created_at,
    updated_at,
    last_used_at,
    created_by,
    partner_id,
    scopes,
    is_revoked,
    revoked_at,
    revoked_reason,
    use_count,
    allowed_ips
)
SELECT
    tak.id,
    tak.partner_id AS tenant_id,           -- partner_id becomes tenant_id
    tak.partner_id AS user_id,             -- partner_id becomes user_id (no separate user)
    tak.name,
    tak.description,
    'trust'::varchar(50) AS key_type,      -- Set type to trust
    tak.key_id,
    tak.key_prefix,
    tak.key_hash,
    1 AS key_version,                      -- Default version for migrated keys
    tak.expires_at,
    tak.created_at AS last_rotated_at,     -- Use created_at as proxy for last_rotated_at
    90 AS rotation_frequency_days,          -- Default 90 days
    60 AS rate_limit_rpm,                  -- Default rate limits
    10000 AS rate_limit_rph,
    100000 AS rate_limit_rpd,
    NOT tak.is_revoked AS is_active,       -- is_active is inverse of is_revoked
    '{}'::jsonb AS metadata,               -- Empty metadata
    tak.created_at,
    tak.updated_at,
    tak.last_used_at,
    tak.created_by,
    tak.partner_id,
    tak.scopes,
    tak.is_revoked,
    tak.revoked_at,
    tak.revoked_reason,
    tak.use_count,
    tak.allowed_ips
FROM trust_api_keys tak
WHERE NOT EXISTS (
    SELECT 1 FROM api_keys ak WHERE ak.id = tak.id
);

-- Log migration summary
DO $$
DECLARE
    migrated_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO migrated_count FROM api_keys WHERE key_type = 'trust';
    RAISE NOTICE 'Migrated % Trust API keys to unified api_keys table', migrated_count;
END $$;
