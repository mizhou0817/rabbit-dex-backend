-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_balance_operation ADD COLUMN IF NOT EXISTS due_block bigint;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_balance_operation DROP COLUMN IF EXISTS due_block;
-- +goose StatementEnd
