-- Allow CLI (and other clients) to request redirect to a custom URL after OAuth callback (e.g. http://127.0.0.1:port/callback)
ALTER TABLE oauth_states ADD COLUMN IF NOT EXISTS redirect_uri TEXT;
