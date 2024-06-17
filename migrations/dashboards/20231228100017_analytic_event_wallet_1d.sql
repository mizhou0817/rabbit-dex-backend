-- +goose Up
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS analytic_event_wallet_1d;

CREATE MATERIALIZED VIEW analytic_event_wallet_1d
    WITH (timescaledb.continuous, timescaledb.materialized_only=false) AS
SELECT time_bucket('1day', ts) as ts,
       wallet,
       SUM(total)              as total
FROM analytic_event_wallet_1h
GROUP BY 1, 2
WITH NO DATA;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP MATERIALIZED VIEW IF EXISTS analytic_event_wallet_1d;
-- +goose StatementEnd
