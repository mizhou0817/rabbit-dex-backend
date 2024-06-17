-- +goose Up
-- +goose StatementBegin

CREATE TABLE app_bfx_game_assets (
    blockchain          TEXT    NOT NULL,
    profile_id          BIGINT  NOT NULL,
    batch_id            BIGINT  NOT NULL,
    trading_points      FLOAT8  NOT NULL,
    staking_points      FLOAT8  NOT NULL,
    bonus_points        FLOAT8  NOT NULL,
    referral_points     FLOAT8  NOT NULL,
    total_points        FLOAT8  NOT NULL,
    vip_extra_boost     FLOAT8  NOT NULL DEFAULT 0,
    wallet              TEXT    NOT NULL,
    liquidations        FLOAT8  NOT NULL,
    referral_boost      FLOAT8  NOT NULL,
    trading_level       TEXT    NOT NULL,
    trading_boost       FLOAT8  NOT NULL,
    cumulative_volume   FLOAT8  NOT NULL,
    timestamp           BIGINT  NOT NULL,
    average_positions   JSONB   NOT NULL,
);

CREATE INDEX app_bfx_game_assets_blockchain_profile_id_batch_id ON app_bfx_game_assets (blockchain, profile_id, batch_id desc);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_bfx_game_assets;
-- +goose StatementEnd
