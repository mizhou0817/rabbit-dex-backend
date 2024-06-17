-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_trade', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_trade_volume_1d;

CREATE MATERIALIZED VIEW app_trade_volume_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) as bucket, market_id, SUM(size * price) as volume
FROM app_trade
GROUP BY bucket, market_id
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_trade_volume_1d;
-- +goose StatementEnd
