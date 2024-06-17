-- +goose Up
-- +goose StatementBegin
CREATE MATERIALIZED VIEW market_data_view AS
WITH current_24h AS (
    SELECT
        SUM(price * size) "average_daily_volume",
        MAX(price) "last_trade_price_24high",
        MIN(price) "last_trade_price_24low",
        LAST(price, timestamp) - FIRST(price, timestamp) "last_trade_price_24h_change_premium",
        (LAST(price, timestamp) - FIRST(price, timestamp))/FIRST(price, timestamp) "last_trade_price_24h_change_basis",
        market_id
    FROM app_trade
    WHERE timestamp >= (EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - INTERVAL '24 hours')) * 1000000)::bigint
    GROUP BY market_id
),
previous_24h AS (
    SELECT
        SUM(price * size) "average_daily_volume",
        market_id
    FROM app_trade
    WHERE (
        timestamp >= (EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - INTERVAL '48 hours')) * 1000000)::bigint AND
        timestamp <  (EXTRACT(EPOCH FROM (CURRENT_TIMESTAMP - INTERVAL '24 hours')) * 1000000)::bigint
    )
    GROUP BY market_id
)
SELECT
    c.average_daily_volume,
    c.last_trade_price_24high,
    c.last_trade_price_24low,
    c.last_trade_price_24h_change_premium,
    c.last_trade_price_24h_change_basis,
    c.average_daily_volume - p.average_daily_volume AS "average_daily_volume_change_premium",
    (c.average_daily_volume - p.average_daily_volume) / p.average_daily_volume AS "average_daily_volume_change_basis",
    c.market_id
FROM current_24h c
LEFT JOIN previous_24h p ON c.market_id = p.market_id;

CREATE OR REPLACE FUNCTION refresh_materialized_view(job_id INT, config JSONB)
RETURNS VOID AS $$
BEGIN
EXECUTE format('REFRESH MATERIALIZED VIEW %I', config->>'view_name');
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
    'refresh_materialized_view',
    '1 seconds',
    config => '{"view_name":"market_data_view"}'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
    (SELECT job_id
     FROM timescaledb_information.jobs
     WHERE proc_name = 'refresh_materialized_view'
     LIMIT 1)
);

DROP FUNCTION refresh_materialized_view;

DROP MATERIALIZED VIEW market_data_view;
-- +goose StatementEnd