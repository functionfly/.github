-- Function Recommendations System
-- Tracks usage patterns, similarity scores, and recommendations

-- 1. Function co-occurrence tracking (collaborative filtering)
-- Records when functions are used together in the same session/time window
CREATE TABLE IF NOT EXISTS function_co_occurrences (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id_a UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_id_b UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    co_occurrence_count INTEGER DEFAULT 1,
    last_co_occurred_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_co_occurrence_pair UNIQUE (function_id_a, function_id_b),
    CONSTRAINT chk_co_occurrence_order CHECK (function_id_a < function_id_b)
);

CREATE INDEX IF NOT EXISTS idx_co_occurrences_func_a ON function_co_occurrences(function_id_a);
CREATE INDEX IF NOT EXISTS idx_co_occurrences_func_b ON function_co_occurrences(function_id_b);
CREATE INDEX IF NOT EXISTS idx_co_occurrences_count ON function_co_occurrences(co_occurrence_count DESC);

-- 2. Function similarity scores (pre-computed for performance)
-- Stores similarity scores between functions based on various strategies
CREATE TABLE IF NOT EXISTS function_similarities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id_a UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    function_id_b UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Similarity scores from different strategies (0.0 to 1.0)
    content_similarity NUMERIC(5, 4) DEFAULT 0,      -- Based on description, inputs, outputs
    collaborative_similarity NUMERIC(5, 4) DEFAULT 0, -- Based on usage patterns
    category_similarity NUMERIC(5, 4) DEFAULT 0,      -- Based on category/tags overlap

    -- Weighted combined score
    combined_similarity NUMERIC(5, 4) DEFAULT 0,

    -- Metadata
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    computation_version INTEGER DEFAULT 1,

    CONSTRAINT uq_similarity_pair UNIQUE (function_id_a, function_id_b),
    CONSTRAINT chk_similarity_order CHECK (function_id_a < function_id_b)
);

CREATE INDEX IF NOT EXISTS idx_similarities_func_a ON function_similarities(function_id_a);
CREATE INDEX IF NOT EXISTS idx_similarities_func_b ON function_similarities(function_id_b);
CREATE INDEX IF NOT EXISTS idx_similarities_combined ON function_similarities(function_id_a, combined_similarity DESC);

-- 3. User function interactions (for personalized recommendations)
-- Tracks how users interact with functions
CREATE TABLE IF NOT EXISTS user_function_interactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Interaction types
    interaction_type VARCHAR(20) NOT NULL CHECK (interaction_type IN (
        'view', 'execute', 'save', 'follow', 'rate', 'copy_code', 'share'
    )),

    -- Context
    session_id VARCHAR(100),
    referrer_function_id UUID REFERENCES registry_functions(id) ON DELETE SET NULL,

    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_interactions_user ON user_function_interactions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_interactions_function ON user_function_interactions(function_id);
CREATE INDEX IF NOT EXISTS idx_user_interactions_type ON user_function_interactions(interaction_type);
CREATE INDEX IF NOT EXISTS idx_user_interactions_session ON user_function_interactions(session_id);
CREATE INDEX IF NOT EXISTS idx_user_interactions_timestamp ON user_function_interactions(timestamp DESC);

-- 4. Function recommendations cache
-- Pre-computed recommendations for quick retrieval
CREATE TABLE IF NOT EXISTS function_recommendations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Recommendation details
    recommended_function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    recommendation_score NUMERIC(5, 4) NOT NULL,
    recommendation_type VARCHAR(20) NOT NULL CHECK (recommendation_type IN (
        'similar', 'frequently_used_together', 'same_category', 'trending', 'personalized'
    )),

    -- Ranking and display
    rank_position INTEGER DEFAULT 0,

    -- Metadata
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,

    CONSTRAINT uq_recommendation_pair UNIQUE (function_id, recommended_function_id, recommendation_type)
);

CREATE INDEX IF NOT EXISTS idx_recommendations_function ON function_recommendations(function_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_recommended ON function_recommendations(recommended_function_id);
CREATE INDEX IF NOT EXISTS idx_recommendations_type ON function_recommendations(recommendation_type);
CREATE INDEX IF NOT EXISTS idx_recommendations_score ON function_recommendations(function_id, recommendation_score DESC);
CREATE INDEX IF NOT EXISTS idx_recommendations_expires ON function_recommendations(expires_at) WHERE expires_at IS NOT NULL;

-- 5. Session function usage (for collaborative filtering)
-- Tracks functions used within a session for pattern detection
CREATE TABLE IF NOT EXISTS session_function_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id VARCHAR(100) NOT NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Usage context
    execution_count INTEGER DEFAULT 1,
    first_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_used_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_session_usage_session ON session_function_usage(session_id);
CREATE INDEX IF NOT EXISTS idx_session_usage_function ON session_function_usage(function_id);
CREATE INDEX IF NOT EXISTS idx_session_usage_user ON session_function_usage(user_id);
CREATE INDEX IF NOT EXISTS idx_session_usage_time ON session_function_usage(last_used_at DESC);

-- 6. Category similarity matrix
-- Pre-computed category relationships for quick lookup
CREATE TABLE IF NOT EXISTS category_similarities (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category_a VARCHAR(50) NOT NULL,
    category_b VARCHAR(50) NOT NULL,
    similarity_score NUMERIC(5, 4) DEFAULT 0,

    -- Based on function overlap and usage patterns
    shared_functions INTEGER DEFAULT 0,
    co_occurrence_count INTEGER DEFAULT 0,

    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_category_pair UNIQUE (category_a, category_b),
    CONSTRAINT chk_category_order CHECK (category_a < category_b)
);

CREATE INDEX IF NOT EXISTS idx_category_sim_a ON category_similarities(category_a);
CREATE INDEX IF NOT EXISTS idx_category_sim_b ON category_similarities(category_b);

-- 7. Function embeddings for content-based similarity
-- Stores vector embeddings for semantic similarity
CREATE TABLE IF NOT EXISTS function_embeddings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    -- Embedding (BYTEA when pgvector not installed; use vector(1536) with CREATE EXTENSION vector if available)
    embedding BYTEA,

    -- Text that was embedded for reference
    embedded_text TEXT,

    -- Metadata
    embedding_model VARCHAR(100) DEFAULT 'text-embedding-ada-002',
    computed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    CONSTRAINT uq_function_embedding UNIQUE (function_id)
);

CREATE INDEX IF NOT EXISTS idx_embeddings_function ON function_embeddings(function_id);

-- 8. Recommendation feedback
-- Tracks user feedback on recommendations for improvement
CREATE TABLE IF NOT EXISTS recommendation_feedback (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    recommended_function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,

    feedback_type VARCHAR(20) NOT NULL CHECK (feedback_type IN (
        'clicked', 'executed', 'dismissed', 'not_relevant', 'helpful'
    )),

    recommendation_type VARCHAR(20),
    timestamp TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_feedback_user ON recommendation_feedback(user_id);
CREATE INDEX IF NOT EXISTS idx_feedback_function ON recommendation_feedback(function_id);
CREATE INDEX IF NOT EXISTS idx_feedback_type ON recommendation_feedback(feedback_type);

-- 9. Function execution events (for recommendation tracking)
-- Tracks function executions with session context for co-occurrence and recommendations
CREATE TABLE IF NOT EXISTS function_execution_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,

    -- Session context for co-occurrence tracking
    session_id VARCHAR(100) NOT NULL,

    -- Execution metadata
    execution_id UUID DEFAULT gen_random_uuid(),
    executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_func_exec_events_function ON function_execution_events(function_id);
CREATE INDEX IF NOT EXISTS idx_func_exec_events_session ON function_execution_events(session_id);
CREATE INDEX IF NOT EXISTS idx_func_exec_events_user ON function_execution_events(user_id) WHERE user_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_func_exec_events_executed ON function_execution_events(executed_at DESC);

-- Function to update co-occurrence counts
CREATE OR REPLACE FUNCTION update_co_occurrence(
    p_function_id_a UUID,
    p_function_id_b UUID
) RETURNS void AS $$
DECLARE
    v_first UUID;
    v_second UUID;
BEGIN
    -- Ensure consistent ordering
    IF p_function_id_a < p_function_id_b THEN
        v_first := p_function_id_a;
        v_second := p_function_id_b;
    ELSE
        v_first := p_function_id_b;
        v_second := p_function_id_a;
    END IF;

    INSERT INTO function_co_occurrences (function_id_a, function_id_b, co_occurrence_count, last_co_occurred_at)
    VALUES (v_first, v_second, 1, NOW())
    ON CONFLICT (function_id_a, function_id_b) DO UPDATE SET
        co_occurrence_count = function_co_occurrences.co_occurrence_count + 1,
        last_co_occurred_at = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to get related functions with scores
CREATE OR REPLACE FUNCTION get_related_functions(
    p_function_id UUID,
    p_limit INTEGER DEFAULT 10,
    p_min_score NUMERIC DEFAULT 0.1
) RETURNS TABLE (
    function_id UUID,
    author VARCHAR,
    name VARCHAR,
    title TEXT,
    description TEXT,
    category TEXT,
    score NUMERIC,
    recommendation_type VARCHAR
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        rf.id,
        rf.author,
        rf.name,
        rf.title,
        rf.description,
        rf.category,
        fr.recommendation_score as score,
        fr.recommendation_type
    FROM function_recommendations fr
    JOIN registry_functions rf ON rf.id = fr.recommended_function_id
    WHERE fr.function_id = p_function_id
      AND fr.recommendation_score >= p_min_score
      AND rf.visibility = 'public'
    ORDER BY fr.recommendation_score DESC
    LIMIT p_limit;
END;
$$ LANGUAGE plpgsql;

-- Function to compute content similarity between two functions
CREATE OR REPLACE FUNCTION compute_content_similarity(
    p_function_id_a UUID,
    p_function_id_b UUID
) RETURNS NUMERIC AS $$
DECLARE
    v_similarity NUMERIC := 0;
    v_tags_a JSONB;
    v_tags_b JSONB;
    v_category_a TEXT;
    v_category_b TEXT;
    v_desc_a TEXT;
    v_desc_b TEXT;
    v_tag_overlap INTEGER;
    v_tag_count_a INTEGER;
    v_tag_count_b INTEGER;
BEGIN
    -- Get function data
    SELECT tags, category, description INTO v_tags_a, v_category_a, v_desc_a
    FROM registry_functions WHERE id = p_function_id_a;

    SELECT tags, category, description INTO v_tags_b, v_category_b, v_desc_b
    FROM registry_functions WHERE id = p_function_id_b;

    -- Category match (weight: 0.3)
    IF v_category_a IS NOT NULL AND v_category_b IS NOT NULL AND v_category_a = v_category_b THEN
        v_similarity := v_similarity + 0.3;
    END IF;

    -- Tag overlap (weight: 0.4)
    IF v_tags_a IS NOT NULL AND v_tags_b IS NOT NULL THEN
        SELECT count(*) INTO v_tag_overlap
        FROM (
            SELECT jsonb_array_elements_text(v_tags_a) AS tag
            INTERSECT
            SELECT jsonb_array_elements_text(v_tags_b)
        ) overlap;

        SELECT jsonb_array_length(v_tags_a), jsonb_array_length(v_tags_b)
        INTO v_tag_count_a, v_tag_count_b;

        IF v_tag_count_a + v_tag_count_b > 0 THEN
            v_similarity := v_similarity + (0.4 * (2.0 * v_tag_overlap::NUMERIC / (v_tag_count_a + v_tag_count_b)));
        END IF;
    END IF;

    -- Description similarity using trigrams (weight: 0.3)
    IF v_desc_a IS NOT NULL AND v_desc_b IS NOT NULL THEN
        v_similarity := v_similarity + (0.3 * similarity(v_desc_a, v_desc_b));
    END IF;

    RETURN LEAST(v_similarity, 1.0);
END;
$$ LANGUAGE plpgsql;

-- Add pgvector extension if available (optional, for advanced embeddings)
-- CREATE EXTENSION IF NOT EXISTS vector;

-- Create trigger to track function executions for co-occurrence
CREATE OR REPLACE FUNCTION track_session_co_occurrence()
RETURNS TRIGGER AS $$
DECLARE
    v_session_id VARCHAR(100);
    v_existing_function_ids UUID[];
BEGIN
    -- Get session_id from the new execution record
    v_session_id := NEW.session_id;

    IF v_session_id IS NOT NULL THEN
        -- Get other functions used in this session
        SELECT array_agg(function_id) INTO v_existing_function_ids
        FROM session_function_usage
        WHERE session_id = v_session_id
          AND function_id != NEW.function_id;

        -- Update co-occurrences
        IF v_existing_function_ids IS NOT NULL THEN
            FOR i IN 1..array_length(v_existing_function_ids, 1) LOOP
                PERFORM update_co_occurrence(NEW.function_id, v_existing_function_ids[i]);
            END LOOP;
        END IF;

        -- Update or insert session usage
        INSERT INTO session_function_usage (session_id, function_id, user_id, execution_count, first_used_at, last_used_at)
        VALUES (v_session_id, NEW.function_id, NEW.user_id, 1, NOW(), NOW())
        ON CONFLICT (session_id, function_id) DO UPDATE SET
            execution_count = session_function_usage.execution_count + 1,
            last_used_at = NOW();
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to function_execution_events for co-occurrence tracking
DROP TRIGGER IF EXISTS trg_track_session_co_occurrence ON function_execution_events;
CREATE TRIGGER trg_track_session_co_occurrence
    AFTER INSERT ON function_execution_events
    FOR EACH ROW
    EXECUTE FUNCTION track_session_co_occurrence();

-- Insert some default category similarities (category_a < category_b for chk_category_order)
INSERT INTO category_similarities (category_a, category_b, similarity_score, shared_functions) VALUES
('string', 'text', 0.8, 0),
('encoding', 'text', 0.7, 0),
('math', 'number', 0.9, 0),
('date', 'time', 0.95, 0),
('array', 'collection', 0.85, 0),
('json', 'object', 0.9, 0),
('sanitization', 'validation', 0.75, 0),
('crypto', 'security', 0.8, 0),
('api', 'http', 0.85, 0),
('file', 'storage', 0.7, 0)
ON CONFLICT (category_a, category_b) DO NOTHING;
