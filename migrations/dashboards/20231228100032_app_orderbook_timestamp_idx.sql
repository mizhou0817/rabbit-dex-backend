-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
-- instant orders for each market
CREATE INDEX IF NOT EXISTS app_orderbook_timestamp_idx
    ON app_orderbook (timestamp DESC) WITH (timescaledb.transaction_per_chunk);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_orderbook_timestamp_idx;
-- +goose StatementEnd
