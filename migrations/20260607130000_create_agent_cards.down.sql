-- Rollback P6: agent_cards table.

DROP INDEX IF EXISTS idx_agent_cards_skills;
DROP INDEX IF EXISTS idx_agent_cards_capabilities;
DROP INDEX IF EXISTS idx_agent_cards_trust;
DROP TABLE IF EXISTS agent_cards;
