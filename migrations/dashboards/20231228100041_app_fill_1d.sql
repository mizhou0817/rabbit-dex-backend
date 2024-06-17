-- +goose Up
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_1d;

CREATE MATERIALIZED VIEW app_fill_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(86400000000::bigint, timestamp) as bucket,
       profile_id,
       sum(volume)                                 as volume,
       sum(fee)                                    as fee,
       SUM(trades)                                 as trades,
       SUM(cnt)                                    as cnt
FROM app_fill_1h
GROUP BY bucket, profile_id
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_1d;
-- +goose StatementEnd
