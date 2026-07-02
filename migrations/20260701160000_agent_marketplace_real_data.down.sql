DROP TRIGGER IF EXISTS trg_increment_agent_listing_calls ON agent_execution_records;
DROP FUNCTION IF EXISTS increment_agent_listing_calls();
DROP TRIGGER IF EXISTS trg_update_agent_listing_rating ON agent_ratings;
DROP FUNCTION IF EXISTS update_agent_listing_rating();
DROP TABLE IF EXISTS agent_ratings;
