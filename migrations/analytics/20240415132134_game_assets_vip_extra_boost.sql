-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_game_assets ADD COLUMN vip_extra_boost float8 NOT NULL DEFAULT 0;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE app_game_assets DROP COLUMN vip_extra_boost;
-- +goose StatementEnd
