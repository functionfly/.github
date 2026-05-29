-- Create function_likes table to track user likes on registry functions
CREATE TABLE IF NOT EXISTS function_likes (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id uuid NOT NULL,
    user_id uuid NOT NULL,
    liked_at timestamptz NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_function_likes_function ON function_likes(function_id);
CREATE INDEX IF NOT EXISTS idx_function_likes_user ON function_likes(user_id);
CREATE INDEX IF NOT EXISTS idx_function_likes_liked_at ON function_likes(liked_at);
CREATE UNIQUE INDEX IF NOT EXISTS idx_function_likes_unique ON function_likes(function_id, user_id);

-- Foreign key to registry_functions (informational, can be added later)
-- ALTER TABLE function_likes ADD CONSTRAINT fk_function FOREIGN KEY (function_id) REFERENCES registry_functions(id);