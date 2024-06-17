-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS app_orderbook_shard_id_archive_id_idx
  ON app_orderbook(shard_id, archive_id);

CREATE INDEX IF NOT EXISTS app_position_shard_id_archive_id_idx
  ON app_position(shard_id, archive_id);

CREATE INDEX IF NOT EXISTS app_market_shard_id_archive_id_idx
  ON app_market(shard_id, archive_id);

CREATE INDEX IF NOT EXISTS app_inv3_data_shard_id_archive_id_idx
  ON app_inv3_data(shard_id, archive_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_orderbook_shard_id_archive_id_idx;
DROP INDEX IF EXISTS app_position_shard_id_archive_id_idx;
DROP INDEX IF EXISTS app_market_shard_id_archive_id_idx;
DROP INDEX IF EXISTS app_inv3_data_shard_id_archive_id_idx;
-- +goose StatementEnd
