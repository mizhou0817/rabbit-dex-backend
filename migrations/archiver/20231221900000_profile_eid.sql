-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_profile
    ADD COLUMN IF NOT EXISTS exchange_id TEXT;
UPDATE app_profile
SET exchange_id = 'rbx';
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_profile
    DROP COLUMN IF EXISTS exchange_id;
-- +goose StatementEnd
