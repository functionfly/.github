-- Create function_favorites table for user function favorites
CREATE TABLE IF NOT EXISTS function_favorites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    function_id UUID NOT NULL REFERENCES registry_functions(id) ON DELETE CASCADE,
    position INTEGER DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_function_favorite UNIQUE (user_id, function_id)
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_function_favorites_user ON function_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_function_favorites_function ON function_favorites(function_id);
CREATE INDEX IF NOT EXISTS idx_function_favorites_user_position ON function_favorites(user_id, position DESC);
CREATE INDEX IF NOT EXISTS idx_function_favorites_created ON function_favorites(created_at DESC);

-- RLS policies
ALTER TABLE function_favorites ENABLE ROW LEVEL SECURITY;

-- Users can manage their own favorites
CREATE POLICY "Users can manage their own favorites" ON function_favorites
    FOR ALL USING (user_id = auth.uid()) WITH CHECK (user_id = auth.uid());

-- Anyone can view favorites (for public profiles)
CREATE POLICY "Anyone can view favorites" ON function_favorites
    FOR SELECT USING (true);

COMMENT ON TABLE function_favorites IS 'Stores user favorites for registry functions';
COMMENT ON COLUMN function_favorites.position IS 'Lower position = higher priority (0 = top favorite)';