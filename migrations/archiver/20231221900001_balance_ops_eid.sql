
-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_balance_operation ADD COLUMN IF NOT EXISTS exchange_id TEXT;
ALTER TABLE app_balance_operation ADD COLUMN IF NOT EXISTS chain_id NUMERIC;
ALTER TABLE app_balance_operation ADD COLUMN IF NOT EXISTS contract_address TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_balance_operation DROP COLUMN IF EXISTS exchange_id;
ALTER TABLE app_balance_operation DROP COLUMN IF EXISTS chain_id;
ALTER TABLE app_balance_operation DROP COLUMN IF EXISTS contract_address;
-- +goose StatementEnd