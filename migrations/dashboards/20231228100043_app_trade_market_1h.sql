-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_trade', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_trade_market_1h;

CREATE MATERIALIZED VIEW app_trade_market_1h
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(3600000000::bigint, timestamp) as timestamp,
       market_id,
       sum(price * size)                          as volume
FROM app_trade
GROUP BY 1, 2
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_trade_market_1h;
-- +goose StatementEnd
