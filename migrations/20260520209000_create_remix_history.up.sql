-- Create remix_history table to track function remix relationships
CREATE TABLE IF NOT EXISTS remix_history (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    source_function_id uuid NOT NULL,
    target_function_id uuid NOT NULL,
    remixed_by_user_id uuid NOT NULL,
    remixed_at timestamptz NOT NULL,
    customization text,
    cost_usd double precision DEFAULT 0,
    created_at timestamptz DEFAULT NOW()
);

CREATE INDEX idx_remix_source ON remix_history(source_function_id);
CREATE INDEX idx_remix_target ON remix_history(target_function_id);
CREATE INDEX idx_remix_user ON remix_history(remixed_by_user_id);
CREATE INDEX idx_remix_at ON remix_history(remixed_at);
CREATE UNIQUE INDEX idx_remix_unique ON remix_history(target_function_id);