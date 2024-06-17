-- +goose Up
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_1w;

CREATE MATERIALIZED VIEW app_fill_1w
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket(604800000000::bigint, bucket) AS timestamp,
       SUM(cnt)                                  as total,
       COUNT(distinct (profile_id))              as total_unique
FROM app_fill_1d
GROUP BY 1
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS app_fill_1w;
-- +goose StatementEnd
