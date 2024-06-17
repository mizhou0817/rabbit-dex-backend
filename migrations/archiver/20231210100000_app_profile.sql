-- +goose Up
-- +goose StatementBegin
-- profile table
CREATE TABLE IF NOT EXISTS app_profile
(
    id                BIGINT NOT NULL,
    profile_type      TEXT   NOT NULL,
    status            TEXT   NOT NULL,
    wallet            TEXT   NOT NULL,
    created_at        BIGINT NOT NULL,
    exchange_id       TEXT   NOT NULL,
    shard_id          TEXT   NOT NULL,
    archive_id        BIGINT NOT NULL,
    archive_timestamp BIGINT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_profile_id_created_at_idx
    ON app_profile (id, created_at);

CREATE UNIQUE INDEX IF NOT EXISTS app_profile_shard_id_archive_id_idx
  ON app_profile(shard_id, archive_id, created_at);

SELECT create_hypertable('app_profile', 'created_at',
                         chunk_time_interval => 28 * 24 * 3600000000,
                         if_not_exists => TRUE
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_profile;
DROP INDEX IF EXISTS app_profile_id_created_at_idx;
-- +goose StatementEnd
