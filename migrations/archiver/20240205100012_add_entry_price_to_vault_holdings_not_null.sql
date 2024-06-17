-- +goose Up
-- +goose StatementBegin
DELETE FROM app_vault_holdings_last;
ALTER TABLE app_vault_holdings_last
    ALTER COLUMN entry_price SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_vault_holdings_last
    ALTER COLUMN entry_price SET NULL;
-- +goose StatementEnd
