-- +goose Up
-- +goose StatementBegin
ALTER TABLE app_order ADD COLUMN IF NOT EXISTS archive_timestamp BIGINT NOT NULL DEFAULT 0;
ALTER TABLE app_trade ADD COLUMN IF NOT EXISTS archive_timestamp BIGINT NOT NULL DEFAULT 0;
ALTER TABLE app_fill ADD COLUMN IF NOT EXISTS archive_timestamp BIGINT NOT NULL DEFAULT 0;
ALTER TABLE app_balance_operation ADD COLUMN IF NOT EXISTS archive_timestamp BIGINT NOT NULL DEFAULT 0;

-- historical orderbook
CREATE TABLE IF NOT EXISTS app_orderbook (
  order_id             TEXT       NOT NULL,
  timestamp            BIGINT     NOT NULL,
  market_id            TEXT       NOT NULL,
  trader_id            BIGINT     NOT NULL,
  price                NUMERIC    NOT NULL,
  size                 NUMERIC    NOT NULL,
  side                 TEXT       NOT NULL,
  reverse              BIGINT     NOT NULL,
  shard_id             TEXT       NOT NULL,
  archive_id           BIGINT     NOT NULL,
  archive_timestamp    BIGINT     NOT NULL
);

CREATE INDEX IF NOT EXISTS app_orderbook_archive_id_idx
  ON app_orderbook(archive_id);

SELECT create_hypertable('app_orderbook', 'archive_timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- historical position
CREATE TABLE IF NOT EXISTS app_position (
  id                   TEXT     NOT NULL,
  market_id            TEXT     NOT NULL,
  profile_id           BIGINT   NOT NULL,
  size                 NUMERIC  NOT NULL,
  side                 TEXT     NOT NULL,
  entry_price          NUMERIC  NOT NULL,
  unrealized_pnl       NUMERIC  NOT NULL,
  notional             NUMERIC  NOT NULL,
  margin               NUMERIC  NOT NULL,
  liquidation_price    NUMERIC  NOT NULL,
  fair_price           NUMERIC  NOT NULL,
  shard_id             TEXT     NOT NULL,
  archive_id           BIGINT   NOT NULL,
  archive_timestamp    BIGINT   NOT NULL
);

CREATE INDEX IF NOT EXISTS app_position_archive_id_idx
  ON app_position(archive_id);

SELECT create_hypertable('app_position', 'archive_timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- historical market
CREATE TABLE IF NOT EXISTS app_market (
  id                                     TEXT       NOT NULL,
  status                                 TEXT       NOT NULL,
  min_initial_margin                     NUMERIC    NOT NULL,
  forced_margin                          NUMERIC    NOT NULL,
  liquidation_margin                     NUMERIC    NOT NULL,
  min_tick                               NUMERIC    NOT NULL,
  min_order                              NUMERIC    NOT NULL,
  best_bid                               NUMERIC    NOT NULL,
  best_ask                               NUMERIC    NOT NULL,
  market_price                           NUMERIC    NOT NULL,
  index_price                            NUMERIC    NOT NULL,
  last_trade_price                       NUMERIC    NOT NULL,
  fair_price                             NUMERIC    NOT NULL,
  last_trade_price_24high                NUMERIC    NOT NULL,
  last_trade_price_24low                 NUMERIC    NOT NULL,
  average_daily_volume                   NUMERIC    NOT NULL,
  instant_funding_rate                   NUMERIC    NOT NULL,
  instant_daily_volume                   NUMERIC    NOT NULL,
  last_funding_rate_basis                NUMERIC    NOT NULL,
  last_trade_price_24h_change_premium    NUMERIC    NOT NULL,
  last_trade_price_24h_change_basis      NUMERIC    NOT NULL,
  average_daily_volume_change_premium    NUMERIC    NOT NULL,
  average_daily_volume_change_basis      NUMERIC    NOT NULL,
  last_update_time                       BIGINT     NOT NULL,
  last_update_sequence                   BIGINT     NOT NULL,
  average_daily_volume_q                 NUMERIC    NOT NULL,
  last_funding_update_time               BIGINT     NOT NULL,
  shard_id                               TEXT       NOT NULL,
  archive_id                             BIGINT     NOT NULL,
  archive_timestamp                      BIGINT     NOT NULL
);

CREATE INDEX IF NOT EXISTS app_market_archive_id_idx
  ON app_market(archive_id);

SELECT create_hypertable('app_market', 'archive_timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- historical inv3_data
CREATE TABLE IF NOT EXISTS app_inv3_data (
  id                   BIGINT     NOT NULL,
  valid                BOOLEAN    NOT NULL,
  last_updated         BIGINT     NOT NULL,
  margined_ae_sum      NUMERIC    NOT NULL,
  exchange_balance     NUMERIC    NOT NULL,
  insurance_balance    NUMERIC    NOT NULL,
  shard_id             TEXT       NOT NULL,
  archive_id           BIGINT     NOT NULL,
  archive_timestamp    BIGINT     NOT NULL
);

CREATE INDEX IF NOT EXISTS app_inv3_data_archive_id_idx
  ON app_inv3_data(archive_id);

SELECT create_hypertable('app_inv3_data', 'archive_timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_orderbook_archive_id_idx;
DROP TABLE IF EXISTS app_orderbook;

DROP INDEX IF EXISTS app_position_archive_id_idx;
DROP TABLE IF EXISTS app_position;

DROP INDEX IF EXISTS app_market_archive_id_idx;
DROP TABLE IF EXISTS app_market;

DROP INDEX IF EXISTS app_inv3_data_archive_id_idx;
DROP TABLE IF EXISTS app_inv3_data;

ALTER TABLE app_balance_operation DROP COLUMN IF EXISTS archive_timestamp;
ALTER TABLE app_fill DROP COLUMN IF EXISTS archive_timestamp;
ALTER TABLE app_trade DROP COLUMN IF EXISTS archive_timestamp;
ALTER TABLE app_order DROP COLUMN IF EXISTS archive_timestamp;
-- +goose StatementEnd
