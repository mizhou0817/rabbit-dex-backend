-- +goose Up
-- +goose StatementBegin
CREATE TABLE app_profile_data (
    profile_id BIGINT NOT NULL PRIMARY KEY,
    version BIGINT NOT NULL,
    data JSONB NOT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_profile_data;
-- +goose StatementEnd
