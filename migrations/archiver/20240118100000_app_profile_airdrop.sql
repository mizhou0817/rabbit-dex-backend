-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_profile_airdrop
(
    profile_id                 BIGINT  NOT NULL,
    airdrop_title              TEXT    NOT NULL,
    status                     TEXT    NOT NULL,
    total_volume_for_airdrop   NUMERIC NOT NULL,
    total_volume_after_airdrop NUMERIC NOT NULL,
    total_rewards              NUMERIC NOT NULL,
    claimable                  NUMERIC NOT NULL,
    claimed                    NUMERIC NOT NULL,
    last_fill_timestamp        JSONB   NOT NULL,
    initial_rewards            NUMERIC NOT NULL,
    shard_id                   TEXT    NOT NULL,
    archive_id                 BIGINT  NOT NULL,
    archive_timestamp          BIGINT  NOT NULL,
    PRIMARY KEY (profile_id, airdrop_title)
);

CREATE UNIQUE INDEX IF NOT EXISTS app_profile_airdrop_shard_id_archive_id_idx
    ON app_profile_airdrop (shard_id, archive_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_profile_airdrop_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_profile_airdrop;
-- +goose StatementEnd
