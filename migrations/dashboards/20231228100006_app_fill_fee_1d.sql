-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_fill', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_fill_fee_1d;

CREATE MATERIALIZED VIEW app_fill_fee_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) as bucket, market_id, sum(fee) as fee
FROM app_fill
GROUP BY bucket, market_id, profile_id
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_fee_1d;
-- +goose StatementEnd
