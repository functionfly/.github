-- Rollback: Copy Trust API keys back from unified api_keys to trust_api_keys
-- Note: This is a partial rollback - only the fields that map directly are restored

UPDATE trust_api_keys tak
SET
    name = ak.name,
    description = ak.description,
    expires_at = ak.expires_at,
    last_used_at = ak.last_used_at,
    is_revoked = ak.is_revoked,
    revoked_at = ak.revoked_at,
    revoked_reason = ak.revoked_reason,
    use_count = ak.use_count,
    allowed_ips = ak.allowed_ips
FROM api_keys ak
WHERE ak.id = tak.id AND ak.key_type = 'trust';

-- Delete migrated Trust keys from api_keys (optional cleanup)
-- Uncomment to clean up:
-- DELETE FROM api_keys WHERE key_type = 'trust';

DO $$
DECLARE
    remaining_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO remaining_count FROM api_keys WHERE key_type = 'trust';
    RAISE NOTICE 'Remaining % Trust API keys in unified api_keys table', remaining_count;
END $$;
