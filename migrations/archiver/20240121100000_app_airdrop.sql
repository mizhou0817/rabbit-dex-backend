-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_airdrop
(
    title             TEXT PRIMARY KEY,
    start_timestamp   BIGINT NOT NULL,
    end_timestamp     BIGINT NOT NULL,
    shard_id          TEXT   NOT NULL,
    archive_id        BIGINT NOT NULL,
    archive_timestamp BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_airdrop_shard_id_archive_id_idx
    ON app_airdrop (shard_id, archive_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_airdrop_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_airdrop;
-- +goose StatementEnd
