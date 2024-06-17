-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_balance_operation', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_balance_operation_pnl_1d;

CREATE MATERIALIZED VIEW app_balance_operation_pnl_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) AS timestamp,
       profile_id,
       sum(amount)                                 as realised_pnl
FROM app_balance_operation
WHERE ops_type = 'pnl'
GROUP BY 1, 2
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_balance_operation_pnl_1d;
-- +goose StatementEnd
