-- +goose NO TRANSACTION
-- +goose Up
-- +goose StatementBegin
-- adding index on app_profile_cache
CREATE INDEX IF NOT EXISTS app_profile_cache_id_archive_timestamp_idx ON app_profile_cache(id, archive_timestamp) WITH (timescaledb.transaction_per_chunk);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_profile_cache_id_archive_timestamp_idx;
-- +goose StatementEnd
