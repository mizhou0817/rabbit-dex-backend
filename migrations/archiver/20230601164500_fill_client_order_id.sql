-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_fill ADD COLUMN IF NOT EXISTS client_order_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_fill DROP COLUMN IF EXISTS client_order_id;
-- +goose StatementEnd
