-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS app_profile_exchange_id_wallet_idx
    ON app_profile (exchange_id, wallet);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_profile_exchange_id_wallet_idx;
-- +goose StatementEnd
