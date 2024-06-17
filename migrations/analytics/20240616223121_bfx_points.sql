-- +goose Up
-- +goose StatementBegin

CREATE TABLE app_bfx_points (
    profile_id       BIGINT  NOT NULL,
    batch_id         BIGINT  NOT NULL,
    exchange_id      TEXT    NOT NULL,
    bonus_points     FLOAT8  NOT NULL,
    bfx_points_total FLOAT8  NOT NULL,
    timestamp        BIGINT  NOT NULL,
    PRIMARY KEY (profile_id, batch_id)
);

CREATE INDEX app_bfx_points_profile_id_batch_id ON app_bfx_points (profile_id, batch_id desc);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_bfx_points;
-- +goose StatementEnd
