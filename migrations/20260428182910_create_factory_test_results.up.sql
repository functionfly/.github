CREATE TABLE IF NOT EXISTS factory_test_results (
    id          UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    function_id UUID        NOT NULL,
    stage       TEXT        NOT NULL,
    passed      BOOLEAN     NOT NULL,
    status      TEXT        NOT NULL,
    score       DECIMAL(5,2) NOT NULL DEFAULT 0,
    duration_ms INTEGER     NOT NULL DEFAULT 0,
    error       TEXT        NOT NULL DEFAULT '',
    details     JSONB       NOT NULL DEFAULT '{}',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_factory_test_results_function_id ON factory_test_results (function_id);
CREATE INDEX IF NOT EXISTS idx_factory_test_results_stage ON factory_test_results (stage);
