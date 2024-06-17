-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_vault_holdings_last
(
    vault_profile_id  BIGINT  NOT NULL,
    staker_profile_id BIGINT  NOT NULL,
    shares            NUMERIC NOT NULL,
    entry_nav         NUMERIC NOT NULL,
    shard_id          TEXT    NOT NULL,
    archive_id        BIGINT  NOT NULL,
    archive_timestamp BIGINT  NOT NULL,
    PRIMARY KEY (vault_profile_id, staker_profile_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS app_vault_holdings_last_shard_id_archive_id_idx
    ON app_vault_holdings_last (shard_id, archive_id);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_vault_holdings_last;
-- +goose StatementEnd
