-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_balance_operation', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_balance_operation_net_deposit_1d;

CREATE MATERIALIZED VIEW app_balance_operation_net_deposit_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) AS timestamp,
       profile_id,
       SUM(
               CASE
                   WHEN ops_type = 'deposit' THEN amount
                   WHEN ops_type = 'withdrawal' THEN -1 * amount
                   END
       )                                           AS net_deposit
FROM app_balance_operation
GROUP BY 1, 2
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_balance_operation_net_deposit_1d;
-- +goose StatementEnd
