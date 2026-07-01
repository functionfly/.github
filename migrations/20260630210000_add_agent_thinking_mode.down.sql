ALTER TABLE agent_identities
  DROP COLUMN IF EXISTS thinking_mode,
  DROP COLUMN IF EXISTS thinking_budget;
