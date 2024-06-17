-- +goose Up
-- +goose StatementBegin
-- adding fields as nullable because updating will be too long in time
ALTER TABLE app_order
    ADD COLUMN IF NOT EXISTS trigger_price NUMERIC,
    ADD COLUMN IF NOT EXISTS size_percent  NUMERIC,
    ADD COLUMN IF NOT EXISTS time_in_force TEXT,
    ADD COLUMN IF NOT EXISTS created_at    BIGINT,
    ADD COLUMN IF NOT EXISTS updated_at    BIGINT;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_order
    DROP COLUMN IF EXISTS trigger_price,
    DROP COLUMN IF EXISTS size_percent,
    DROP COLUMN IF EXISTS time_in_force,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS updated_at;
-- +goose StatementEnd
