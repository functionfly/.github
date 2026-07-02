-- Function marketplace real-data infrastructure
-- 1. function_user_ratings table (per-tenant star ratings for functions)
-- 2. Trigger: auto-increment function_listings.call_volume on execution
-- 3. Trigger: sync function_listings.rating_score from registry_function_ratings
-- 4. Trigger: update function_listings.rating_score from user ratings

-- ============================================================
-- Function User Ratings Table (1-5 stars + review, per tenant)
-- ============================================================
CREATE TABLE IF NOT EXISTS function_user_ratings (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id  UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    tenant_id    UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    rating       SMALLINT NOT NULL CHECK (rating >= 1 AND rating <= 5),
    review       TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (function_id, tenant_id)
);

CREATE INDEX IF NOT EXISTS idx_function_user_ratings_function ON function_user_ratings (function_id);
CREATE INDEX IF NOT EXISTS idx_function_user_ratings_tenant ON function_user_ratings (tenant_id);

-- ============================================================
-- Trigger: auto-increment function_listings.call_volume on execution
-- ============================================================
CREATE OR REPLACE FUNCTION increment_function_listing_calls()
RETURNS TRIGGER AS $$
BEGIN
    UPDATE function_listings
    SET call_volume = call_volume + 1,
        updated_at = now()
    WHERE function_id = NEW.function_id;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_increment_function_listing_calls
    AFTER INSERT ON registry_function_executions
    FOR EACH ROW EXECUTE FUNCTION increment_function_listing_calls();

-- ============================================================
-- Trigger: sync rating_score from registry_function_ratings
-- Converts overall_score (0-100) to 0-5 scale
-- ============================================================
CREATE OR REPLACE FUNCTION sync_function_listing_rating()
RETURNS TRIGGER AS $$
DECLARE
    computed_score NUMERIC(3,2);
    user_avg NUMERIC(3,2);
    final_score NUMERIC(3,2);
BEGIN
    -- Get computed score from registry_function_ratings (0-100 → 0-5)
    IF NEW.overall_score IS NOT NULL AND NEW.overall_score > 0 THEN
        computed_score := ROUND((NEW.overall_score / 20.0)::numeric, 2);
    ELSE
        computed_score := NULL;
    END IF;

    -- Get user rating average (0-5)
    SELECT ROUND(AVG(rating)::numeric, 2) INTO user_avg
    FROM function_user_ratings
    WHERE function_id = NEW.function_id;

    -- Weighted blend: 60% computed, 40% user ratings (when both exist)
    IF computed_score IS NOT NULL AND user_avg IS NOT NULL THEN
        final_score := ROUND((computed_score * 0.6 + user_avg * 0.4)::numeric, 2);
    ELSIF computed_score IS NOT NULL THEN
        final_score := computed_score;
    ELSIF user_avg IS NOT NULL THEN
        final_score := user_avg;
    ELSE
        final_score := 0;
    END IF;

    UPDATE function_listings
    SET rating_score = final_score,
        updated_at = now()
    WHERE function_id = NEW.function_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_sync_function_listing_rating
    AFTER INSERT OR UPDATE ON registry_function_ratings
    FOR EACH ROW EXECUTE FUNCTION sync_function_listing_rating();

-- ============================================================
-- Trigger: update rating_score when user rating changes
-- ============================================================
CREATE OR REPLACE FUNCTION update_function_listing_user_rating()
RETURNS TRIGGER AS $$
DECLARE
    computed_score NUMERIC(3,2);
    user_avg NUMERIC(3,2);
    final_score NUMERIC(3,2);
    target_function_id UUID;
BEGIN
    target_function_id := COALESCE(NEW.function_id, OLD.function_id);

    -- Get computed score from registry_function_ratings (0-100 → 0-5)
    SELECT ROUND((overall_score / 20.0)::numeric, 2) INTO computed_score
    FROM registry_function_ratings
    WHERE function_id = target_function_id
      AND overall_score > 0;

    -- Get user rating average
    SELECT ROUND(AVG(rating)::numeric, 2) INTO user_avg
    FROM function_user_ratings
    WHERE function_id = target_function_id;

    IF computed_score IS NOT NULL AND user_avg IS NOT NULL THEN
        final_score := ROUND((computed_score * 0.6 + user_avg * 0.4)::numeric, 2);
    ELSIF computed_score IS NOT NULL THEN
        final_score := computed_score;
    ELSIF user_avg IS NOT NULL THEN
        final_score := user_avg;
    ELSE
        final_score := 0;
    END IF;

    UPDATE function_listings
    SET rating_score = final_score,
        updated_at = now()
    WHERE function_id = target_function_id;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trg_update_function_listing_user_rating
    AFTER INSERT OR UPDATE OR DELETE ON function_user_ratings
    FOR EACH ROW EXECUTE FUNCTION update_function_listing_user_rating();

-- ============================================================
-- Backfill: compute current call_volume from existing executions
-- ============================================================
UPDATE function_listings fl
SET call_volume = COALESCE(
    (SELECT COUNT(*) FROM registry_function_executions rfe WHERE rfe.function_id = fl.function_id),
    fl.call_volume
);

-- Backfill rating_score from registry_function_ratings
UPDATE function_listings fl
SET rating_score = COALESCE(
    (
        SELECT CASE
            WHEN rfr.overall_score IS NULL OR rfr.overall_score = 0 THEN fl.rating_score
            ELSE ROUND((rfr.overall_score / 20.0)::numeric, 2)
        END
        FROM registry_function_ratings rfr
        WHERE rfr.function_id = fl.function_id
    ),
    fl.rating_score
);
