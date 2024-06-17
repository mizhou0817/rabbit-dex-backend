-- +goose Up
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS analytic_event_wallet_1w;

CREATE MATERIALIZED VIEW analytic_event_wallet_1w
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket('1week', ts) as ts,
       SUM(total)               as total,
       COUNT(distinct (wallet)) as total_unique
FROM analytic_event_wallet_1d
GROUP BY 1
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS analytic_event_wallet_1w;
-- +goose StatementEnd
