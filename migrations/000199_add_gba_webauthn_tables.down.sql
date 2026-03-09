-- Drop GoBetterAuth WebAuthn tables

-- Drop tables (will cascade to indexes)
DROP TABLE IF EXISTS gba_webauthn_sessions;
DROP TABLE IF EXISTS gba_webauthn_credentials;