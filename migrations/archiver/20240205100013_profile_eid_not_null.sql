-- +goose Up
-- +goose StatementBegin
UPDATE app_profile
SET exchange_id = 'rbx'
WHERE exchange_id is null;

ALTER TABLE app_profile
    ALTER COLUMN exchange_id SET NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_profile
    ALTER COLUMN exchange_id SET NULL;
-- +goose StatementEnd
