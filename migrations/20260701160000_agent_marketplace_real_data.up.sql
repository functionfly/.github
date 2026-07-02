-- Agent marketplace real-data infrastructure
-- 1. agent_ratings table (per-tenant ratings for agents)
-- 2. Trigger: auto-increment agent_listings.total_calls on each execution
-- 3. Trigger: auto-update agent_listings.rating_score on rating change
-- 4. Trigger: auto-update agent_listings.roi_score from execution success rate

-- ============================================================
-- Agent Ratings Table
-- ============================================================
CREATE TABLE IF NOT EXISTS agent_ratings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    agent_id     TEXT NOT NULL REFERENCES agent_identities(agent_id) ON DELETE CASCADE,
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rating       SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (agent_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_agent_ratings_agent ON agent_ratings (agent_id);
CREATE INDEX IF NOT EXISTS idx_agent_ratings_tenant ON agent_ratings (tenant_id);

-- ============================================================
-- Trigger: auto-update agent_listings.rating_score on rating change
-- ============================================================
CREATE OR REPLACE FUNCTION update_agent_listing_rating()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE agent_listings
    SET rating_score = COALESCE(
        (SELECT ROUND(AVG(rating)::numeric, 2) FROM agent_ratings WHERE agent_id = COALESCE(NEW.agent_id, OLD.agent_id)),
        0
    ),
    updated_at = now()
    WHERE agent_id = COALESCE(NEW.agent_id, OLD.agent_id);
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_agent_listing_rating
    AFTER INSERT OR UPDATE OR DELETE ON agent_ratings
    FOR EACH ROW EXECUTE FUNCTION update_agent_listing_rating();

-- ============================================================
-- Trigger: auto-increment agent_listings.total_calls on execution
-- ============================================================
CREATE OR REPLACE FUNCTION increment_agent_listing_calls()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE agent_listings
    SET total_calls = total_calls + 1,
        updated_at = now()
    WHERE agent_id = NEW.agent_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_increment_agent_listing_calls
    AFTER INSERT ON agent_execution_records
    FOR EACH ROW EXECUTE FUNCTION increment_agent_listing_calls();

-- ============================================================
-- Trigger: auto-update agent_listings.roi_score from success rate
-- roi_score = (successful_executions / total_executions) * 100
-- over the last 30 days, rounded to nearest integer
-- ============================================================
CREATE OR REPLACE FUNCTION update_agent_listing_roi()
RETURNS TRIGGER AS $$
DECLARE
    total BIGINT;
    succeeded BIGINT;
    new_roi NUMERIC(5,2);
BEGIN
    SELECT COUNT(*), COUNT(*) FILTER (WHERE outcome = 'success')
    INTO total, succeeded
    FROM agent_execution_records
    WHERE agent_id = NEW.agent_id
      AND timestamp > now() - INTERVAL '30 days';

    IF total = 0 THEN
        new_roi := 0;
    ELSE
        new_roi := ROUND((succeeded::numeric / total::numeric) * 100, 2);
    END IF;

    UPDATE agent_listings
    SET roi_score = new_roi,
        updated_at = now()
    WHERE agent_id = NEW.agent_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_agent_listing_roi
    AFTER INSERT ON agent_execution_records
    FOR EACH ROW EXECUTE FUNCTION update_agent_listing_roi();

-- ============================================================
-- Backfill: compute current stats from existing data
-- ============================================================

-- Backfill total_calls
UPDATE agent_listings al
SET total_calls = COALESCE(
    (SELECT COUNT(*) FROM agent_execution_records aer WHERE aer.agent_id = al.agent_id),
    al.total_calls
);

-- Backfill roi_score from success rate
UPDATE agent_listings al
SET roi_score = COALESCE(
    (
        SELECT CASE
            WHEN COUNT(*) = 0 THEN al.roi_score  -- keep seed value if no executions
            ELSE ROUND((COUNT(*) FILTER (WHERE outcome = 'success')::numeric / COUNT(*)::numeric) * 100, 2)
        END
        FROM agent_execution_records
        WHERE agent_id = al.agent_id
          AND timestamp > now() - INTERVAL '30 days'
    ),
    al.roi_score
);
