-- +goose Up
-- +goose StatementBegin

CREATE TABLE app_game_assets (
    blockchain     text,
    profile_id     bigint,
    batch_id       bigint  NOT NULL,
    trading_points float8  NOT NULL,
    staking_points float8  NOT NULL,
    bonus_points   float8  NOT NULL,
    total_points   float8  NOT NULL,
    trading_gold   float8  NOT NULL,
    staking_gold   float8  NOT NULL,
    bonus_gold     float8  NOT NULL,
    total_gold     float8  NOT NULL,
    PRIMARY KEY (blockchain, profile_id)
);

CREATE INDEX app_game_assets_blockchain_batch_id   ON app_game_assets (blockchain, batch_id   desc);
CREATE INDEX app_game_assets_blockchain_total_gold ON app_game_assets (blockchain, total_gold desc);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_game_assets;
-- +goose StatementEnd
