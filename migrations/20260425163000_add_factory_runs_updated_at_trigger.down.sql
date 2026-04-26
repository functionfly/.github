-- +migrate Down
DROP TRIGGER IF EXISTS trigger_factory_runs_updated_at ON factory_runs;
DROP FUNCTION IF EXISTS update_factory_runs_updated_at();
