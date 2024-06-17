-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_order ADD COLUMN IF NOT EXISTS client_order_id TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_order DROP COLUMN IF EXISTS client_order_id;
-- +goose StatementEnd
