-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_market ADD COLUMN IF NOT EXISTS icon_url TEXT;
ALTER TABLE app_market ADD COLUMN IF NOT EXISTS market_title TEXT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_market DROP COLUMN IF EXISTS icon_url;
ALTER TABLE app_market DROP COLUMN IF EXISTS market_title;
-- +goose StatementEnd
