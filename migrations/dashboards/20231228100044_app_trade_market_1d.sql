-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_trade', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_trade_market_1d;

CREATE MATERIALIZED VIEW app_trade_market_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) as timestamp,
       market_id,
       sum(volume)                                 as volume
FROM app_trade_market_1h
GROUP BY 1, 2
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_trade_market_1d;
-- +goose StatementEnd
