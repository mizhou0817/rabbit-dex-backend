-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_vault_holdings_last
    ADD COLUMN entry_price NUMERIC NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_vault_holdings_last
    DROP COLUMN entry_price;
-- +goose StatementEnd
