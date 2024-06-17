-- +goose Up
-- +goose StatementBegin
-- instant orders for each market
CREATE TABLE IF NOT EXISTS app_order (
  id                   TEXT      NOT NULL,
  profile_id           BIGINT    NOT NULL,
  market_id            TEXT      NOT NULL,
  order_type           TEXT      NOT NULL,
  status               TEXT      NOT NULL,
  price                NUMERIC,
  size                 NUMERIC,
  initial_size         NUMERIC,
  total_filled_size    NUMERIC,
  side                 TEXT,
  timestamp            BIGINT    NOT NULL,
  reason               TEXT,
  shard_id             TEXT      NOT NULL,
  archive_id           BIGINT    NOT NULL
);

CREATE UNIQUE INDEX  IF NOT EXISTS app_order_id_idx
  ON app_order(id, timestamp);

CREATE INDEX IF NOT EXISTS app_order_profile_id_idx
  ON app_order(profile_id);

CREATE INDEX IF NOT EXISTS app_order_profile_id_market_id_idx
  ON app_order(profile_id, market_id);

CREATE INDEX IF NOT EXISTS app_order_profile_id_status_idx
  ON app_order(profile_id, status);

CREATE INDEX IF NOT EXISTS app_order_market_id_idx
  ON app_order(market_id);

CREATE INDEX IF NOT EXISTS app_order_archive_id_idx
  ON app_order(archive_id);

CREATE UNIQUE INDEX IF NOT EXISTS app_order_shard_id_archive_id_idx
  ON app_order(shard_id, archive_id, timestamp);

SELECT create_hypertable('app_order', 'timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- instant trades for each market
CREATE TABLE IF NOT EXISTS app_trade (
  id                   TEXT      NOT NULL,
  market_id            TEXT      NOT NULL,
  timestamp            BIGINT    NOT NULL,
  price                NUMERIC,
  size                 NUMERIC,
  liquidation          BOOLEAN,
  taker_side           TEXT,
  shard_id             TEXT      NOT NULL,
  archive_id           BIGINT    NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_trade_id_idx
  ON app_trade(id, timestamp);

CREATE INDEX IF NOT EXISTS app_trade_market_id_idx
  ON app_trade(market_id);

CREATE INDEX IF NOT EXISTS app_trade_archive_id_idx
  ON app_trade(archive_id);

CREATE UNIQUE INDEX IF NOT EXISTS app_trade_shard_id_archive_id_idx
  ON app_trade(shard_id, archive_id, timestamp);

SELECT create_hypertable('app_trade', 'timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);

-- instant fill for each market
CREATE TABLE IF NOT EXISTS app_fill (
  id                   TEXT      NOT NULL,
  profile_id           BIGINT    NOT NULL,
  market_id            TEXT      NOT NULL,
  order_id             TEXT      NOT NULL,
  timestamp            BIGINT    NOT NULL,
  trade_id             TEXT      NOT NULL,
  price                NUMERIC,
  size                 NUMERIC,
  side                 TEXT,
  is_maker             BOOLEAN,
  fee                  NUMERIC,
  liquidation          BOOLEAN,
  shard_id             TEXT      NOT NULL,
  archive_id           BIGINT    NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_fill_id_idx
  ON app_fill(id, timestamp);

CREATE INDEX IF NOT EXISTS app_fill_market_id_idx
  ON app_fill(market_id);

CREATE INDEX IF NOT EXISTS app_fill_profile_id_idx
  ON app_fill(profile_id);

CREATE INDEX IF NOT EXISTS app_fill_order_id_idx
  ON app_fill(order_id);

CREATE INDEX IF NOT EXISTS app_fill_archive_id_idx
  ON app_fill(archive_id);

CREATE UNIQUE INDEX IF NOT EXISTS app_fill_shard_id_archive_id_idx
  ON app_fill(shard_id, archive_id, timestamp);

SELECT create_hypertable('app_fill', 'timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);


-- instant balance operations for each market
CREATE TABLE IF NOT EXISTS app_balance_operation (
  id                   TEXT      NOT NULL,
  status               TEXT      NOT NULL,
  reason               TEXT      NOT NULL,
  txhash               TEXT      NOT NULL,
  profile_id           BIGINT    NOT NULL,
  wallet               TEXT      NOT NULL,
  ops_type             TEXT      NOT NULL,
  ops_id2              TEXT      NOT NULL,
  amount               NUMERIC,
  timestamp            BIGINT    NOT NULL,
  exchange_id          TEXT,
  chain_id             NUMERIC,
  contract_address     TEXT,
  shard_id             TEXT      NOT NULL,
  archive_id           BIGINT    NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_balance_operation_id_idx
  ON app_balance_operation(id, timestamp);

CREATE INDEX IF NOT EXISTS app_balance_operation_ops_id2_idx
  ON app_balance_operation(ops_id2);

CREATE INDEX IF NOT EXISTS app_balance_operation_ops_type_status_idx
  ON app_balance_operation(ops_type, status);

CREATE INDEX IF NOT EXISTS app_balance_operation_profile_id_ops_type_status_amount_idx
  ON app_balance_operation(profile_id, ops_type, status, amount);

CREATE INDEX IF NOT EXISTS app_balance_operation_archive_id_idx
  ON app_balance_operation(archive_id);

CREATE UNIQUE INDEX IF NOT EXISTS app_balance_operation_shard_id_archive_id_idx
  ON app_balance_operation(shard_id, archive_id, timestamp);

SELECT create_hypertable('app_balance_operation', 'timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);


-- historical profile cache
CREATE TABLE IF NOT EXISTS app_profile_cache (
  id                      BIGINT     NOT NULL,
  profile_type            TEXT       NOT NULL,
  status                  TEXT       NOT NULL,
  wallet                  TEXT       NOT NULL,
  last_update             BIGINT     NOT NULL,
  balance                 NUMERIC    NOT NULL,
  account_equity          NUMERIC    NOT NULL,
  total_position_margin   NUMERIC    NOT NULL,
  total_order_margin      NUMERIC    NOT NULL,
  total_notional          NUMERIC    NOT NULL,
  account_margin          NUMERIC    NOT NULL,
  withdrawable_balance    NUMERIC    NOT NULL,
  cum_unrealized_pnl      NUMERIC    NOT NULL,
  health                  NUMERIC    NOT NULL,
  account_leverage        NUMERIC    NOT NULL,
  cum_trading_volume      NUMERIC    NOT NULL,
  leverage                JSONB      NOT NULL,
  last_liq_check          BIGINT     NOT NULL,
  shard_id                TEXT       NOT NULL,
  archive_id              BIGINT     NOT NULL,
  archive_timestamp       BIGINT     NOT NULL
);

CREATE INDEX IF NOT EXISTS app_profile_cache_archive_id_idx
  ON app_profile_cache(archive_id);

SELECT create_hypertable('app_profile_cache', 'archive_timestamp',
  chunk_time_interval => 3600000000,
  if_not_exists       => TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_order_id_idx;
DROP INDEX IF EXISTS app_order_profile_id_idx;
DROP INDEX IF EXISTS app_order_profile_id_market_id_idx;
DROP INDEX IF EXISTS app_order_profile_id_status_idx;
DROP INDEX IF EXISTS app_order_market_id_idx;
DROP INDEX IF EXISTS app_order_archive_id_idx;
DROP INDEX IF EXISTS app_order_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_order;

DROP INDEX IF EXISTS app_trade_id_idx;
DROP INDEX IF EXISTS app_trade_market_id_idx;
DROP INDEX IF EXISTS app_trade_archive_id_idx;
DROP INDEX IF EXISTS app_trade_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_trade;

DROP INDEX IF EXISTS app_fill_fill_id_idx;
DROP INDEX IF EXISTS app_fill_market_id_idx;
DROP INDEX IF EXISTS app_fill_profile_id_idx;
DROP INDEX IF EXISTS app_fill_order_id_idx;
DROP INDEX IF EXISTS app_fill_archive_id_idx;
DROP INDEX IF EXISTS app_fill_shard_id_archive_id_idx;
DROP TABLE IF EXISTS app_fill;

DROP INDEX IF EXISTS app_balance_operation_shard_id_archive_id_idx;
DROP INDEX IF EXISTS app_balance_operation_archive_id_idx;
DROP INDEX IF EXISTS app_balance_operation_profile_id_ops_type_status_amount_idx;
DROP INDEX IF EXISTS app_balance_operation_ops_type_status_idx;
DROP INDEX IF EXISTS app_balance_operation_ops_id2_idx;
DROP INDEX IF EXISTS app_balance_operation_id_idx;
DROP TABLE IF EXISTS app_balance_operation;

DROP INDEX IF EXISTS app_profile_cache_archive_id_idx;
DROP TABLE IF EXISTS app_profile_cache;
-- +goose StatementEnd
