-- +migrate Up
-- +migrate StatementBegin
CREATE OR REPLACE FUNCTION update_factory_runs_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_factory_runs_updated_at
    BEFORE UPDATE ON factory_runs
    FOR EACH ROW
    EXECUTE FUNCTION update_factory_runs_updated_at();
-- +migrate StatementEnd

-- +migrate Down
-- +migrate StatementBegin
DROP TRIGGER IF EXISTS trigger_factory_runs_updated_at ON factory_runs;
DROP FUNCTION IF EXISTS update_factory_runs_updated_at();
-- +migrate StatementEnd
