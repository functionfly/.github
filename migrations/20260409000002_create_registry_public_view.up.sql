-- Create a SECURITY DEFINER view for public registry access
-- This bypasses RLS completely for read-only public function listings

-- ============================================
-- 1. Drop existing view if exists
-- ============================================
DROP VIEW IF EXISTS v_registry_functions_public CASCADE;

-- ============================================
-- 2. Create the public view with security definer
-- ============================================
-- This view is owned by the table owner and bypasses RLS
CREATE VIEW v_registry_functions_public AS
SELECT
    id,
    author,
    name,
    latest_version,
    title,
    description,
    category,
    tags,
    visibility,
    price_per_call,
    popularity_score,
    reliability_score,
    deterministic_score,
    tenant_id,
    owner_user_id,
    created_at,
    updated_at
FROM registry_functions
WHERE visibility = 'public';

-- ============================================
-- 3. Grant appropriate permissions
-- ============================================
-- Allow all users to select from this view
-- Note: Actual grants depend on your database user setup
GRANT SELECT ON v_registry_functions_public TO PUBLIC;

-- ============================================
-- 4. Create function to list public functions via the view
-- ============================================
-- This function uses SECURITY DEFINER to bypass RLS
CREATE OR REPLACE FUNCTION list_public_registry_functions(
    p_author VARCHAR DEFAULT NULL,
    p_category VARCHAR DEFAULT NULL,
    p_limit INTEGER DEFAULT 500,
    p_offset INTEGER DEFAULT 0
)
RETURNS TABLE (
    id UUID,
    author VARCHAR,
    name VARCHAR,
    latest_version VARCHAR,
    title VARCHAR,
    description TEXT,
    category VARCHAR,
    tags JSONB,
    visibility VARCHAR,
    price_per_call NUMERIC,
    popularity_score INTEGER,
    reliability_score NUMERIC,
    deterministic_score NUMERIC,
    tenant_id UUID,
    owner_user_id UUID,
    created_at TIMESTAMPTZ,
    updated_at TIMESTAMPTZ
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        rf.id,
        rf.author,
        rf.name,
        rf.latest_version,
        rf.title,
        rf.description,
        rf.category,
        rf.tags,
        rf.visibility,
        rf.price_per_call,
        rf.popularity_score,
        rf.reliability_score,
        rf.deterministic_score,
        rf.tenant_id,
        rf.owner_user_id,
        rf.created_at,
        rf.updated_at
    FROM registry_functions rf
    WHERE rf.visibility = 'public'
        AND (p_author IS NULL OR rf.author = p_author)
        AND (p_category IS NULL OR rf.category = p_category)
    ORDER BY rf.created_at DESC
    LIMIT p_limit
    OFFSET p_offset;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 5. Create count function
-- ============================================
CREATE OR REPLACE FUNCTION count_public_registry_functions(
    p_author VARCHAR DEFAULT NULL,
    p_category VARCHAR DEFAULT NULL
)
RETURNS INTEGER AS $$
DECLARE
    v_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO v_count
    FROM registry_functions rf
    WHERE rf.visibility = 'public'
        AND (p_author IS NULL OR rf.author = p_author)
        AND (p_category IS NULL OR rf.category = p_category);

    RETURN v_count;
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- ============================================
-- 6. Verify the view works
-- ============================================
SELECT
    'v_registry_functions_public' as view_name,
    COUNT(*) as public_function_count
FROM v_registry_functions_public;
