-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_vault_last
(
    vault_profile_id     BIGINT PRIMARY KEY,
    manager_profile_id   BIGINT  NOT NULL,
    treasurer_profile_id BIGINT  NOT NULL,
    performance_fee      NUMERIC NOT NULL,
    status               TEXT    NOT NULL,
    total_shares         NUMERIC NOT NULL,
    vault_name           TEXT    NOT NULL,
    manager_name         TEXT    NOT NULL,
    initialised_at       BIGINT  NOT NULL,
    shard_id             TEXT    NOT NULL,
    archive_id           BIGINT  NOT NULL,
    archive_timestamp    BIGINT  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_vault_last_shard_id_archive_id_idx
    ON app_vault_last (shard_id, archive_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_vault_last_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_vault_last;
-- +goose StatementEnd
