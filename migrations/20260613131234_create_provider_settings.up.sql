-- Create provider_settings table for platform-wide provider maintenance mode
CREATE TABLE IF NOT EXISTS provider_settings (
  provider VARCHAR(50) PRIMARY KEY,
  disabled BOOLEAN NOT NULL DEFAULT false,
  disabled_reason TEXT,
  disabled_at TIMESTAMPTZ,
  disabled_by VARCHAR(255)
);

-- Seed with current providers (none disabled by default)
INSERT INTO provider_settings (provider, disabled) VALUES
  ('cloudflare', false),
  ('vercel', false),
  ('fly', false),
  ('deno', false),
  ('functionfly-edge', false),
  ('aws-lambda', false)
ON CONFLICT (provider) DO NOTHING;
