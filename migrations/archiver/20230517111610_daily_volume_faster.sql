-- +goose Up
-- +goose StatementBegin
CREATE UNIQUE INDEX market_data_view_market_id_unique_idx ON market_data_view(market_id);

CREATE OR REPLACE FUNCTION refresh_materialized_view(job_id INT, config JSONB)
RETURNS VOID AS $$
BEGIN
EXECUTE format('REFRESH MATERIALIZED VIEW CONCURRENTLY %I', config->>'view_name');
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS market_data_view_market_id_unique_idx;
DROP FUNCTION IF EXISTS refresh_materialized_view;
-- +goose StatementEnd