-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
-- instant orders for each market
CREATE INDEX IF NOT EXISTS app_market_last_update_time_idx
    ON app_market (last_update_time DESC) WITH (timescaledb.transaction_per_chunk);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_market_last_update_time_idx;
-- +goose StatementEnd
