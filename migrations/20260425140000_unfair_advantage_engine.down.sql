-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS swarm_metrics_snapshots;
DROP TABLE IF EXISTS stealth_runs;
DROP TABLE IF EXISTS rd_lab_runs;
DROP TABLE IF EXISTS internal_functions;
DROP TABLE IF EXISTS internal_opportunities;
-- +goose StatementEnd