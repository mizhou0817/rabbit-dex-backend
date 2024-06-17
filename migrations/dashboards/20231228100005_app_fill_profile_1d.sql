-- +goose Up
-- +goose StatementBegin
SELECT set_integer_now_func('app_fill', 'unix_now', replace_if_exists=> true);

DROP MATERIALIZED VIEW IF EXISTS app_fill_profile_1d;

CREATE MATERIALIZED VIEW app_fill_profile_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) as bucket, profile_id
FROM app_fill
GROUP BY bucket, profile_id
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_profile_1d;
-- +goose StatementEnd
