-- +goose NO TRANSACTION
-- +goose Up
CREATE OR REPLACE FUNCTION unix_now() returns BIGINT LANGUAGE SQL STABLE AS
$$
SELECT EXTRACT(EPOCH from current_timestamp) * 1000000
$$;

SELECT set_integer_now_func('app_market', 'unix_now', replace_if_exists => TRUE);

CREATE MATERIALIZED VIEW IF NOT EXISTS funding_rate_hourly WITH (timescaledb.continuous) AS
SELECT
    id AS "market_id",
    TIME_BUCKET((EXTRACT(EPOCH FROM ('1 hour'::interval)) * 1000000)::bigint, archive_timestamp) AS "timestamp",
    LAST(last_funding_rate_basis, archive_timestamp) AS "funding_rate"
FROM app_market
GROUP BY 1, 2
ORDER BY 2 DESC;

CREATE INDEX IF NOT EXISTS funding_rate_hourly_market_id_timestamp_idx ON funding_rate_hourly(market_id, timestamp);

SELECT add_continuous_aggregate_policy('funding_rate_hourly',
    start_offset => (EXTRACT(EPOCH FROM INTERVAL '3 hours') * 1000000)::bigint,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 hour') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 hour'
);

-- +goose Down
DROP MATERIALIZED VIEW IF EXISTS funding_rate_hourly;

SELECT remove_continuous_aggregate_policy('funding_rate_hourly', if_exists => TRUE);
