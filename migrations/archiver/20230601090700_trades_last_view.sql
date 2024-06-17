-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW market_last_trade_view AS

SELECT market_id, time_bucket('1 hour', TO_TIMESTAMP(timestamp/1e6)) "time", last(price, timestamp) "price"
FROM app_trade
WHERE timestamp >= (EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - INTERVAL '50 hours')) * 1000000)::bigint
GROUP BY 1,2
ORDER BY 1,2;

CREATE INDEX market_last_trade_view_market_id_idx ON market_last_trade_view(market_id);

SELECT add_job(
    'refresh_materialized_view',
    '1 minute',
    config => '{"view_name":"market_last_trade_view"}'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
    (SELECT job_id
     FROM timescaledb_information.jobs
     WHERE config->>'view_name' = 'market_last_trade_view'
     LIMIT 1)
);

DROP INDEX IF EXISTS market_last_trade_view_market_id_idx;
DROP MATERIALIZED VIEW IF EXISTS market_last_trade_view;
-- +goose StatementEnd