-- +goose Up
-- +goose StatementBegin
-- system metric
CREATE TABLE IF NOT EXISTS tnt_system_metric (
  name TEXT NOT NULL,
  value            DOUBLE PRECISION    NOT NULL,
  instance_name    TEXT                NOT NULL,
  meta             JSONB,
  ts               TIMESTAMPTZ         NOT NULL
);

CREATE INDEX IF NOT EXISTS meta_idx_tnt_system_metric
  ON tnt_system_metric USING GIN(meta);

SELECT create_hypertable('tnt_system_metric', 'ts',
  chunk_time_interval => INTERVAL '1 hour',
  if_not_exists       => TRUE
);

SELECT add_retention_policy('tnt_system_metric', INTERVAL '14 days',
  if_not_exists => TRUE
);

-- app metric
CREATE TABLE IF NOT EXISTS tnt_app_metric (
  name             TEXT                NOT NULL,
  value            DOUBLE PRECISION    NOT NULL,
  instance_name    TEXT                NOT NULL,
  meta             JSONB,
  ts               TIMESTAMPTZ         NOT NULL
);

CREATE INDEX IF NOT EXISTS meta_idx_tnt_app_metric
  ON tnt_app_metric USING GIN(meta);

SELECT create_hypertable('tnt_app_metric', 'ts',
  chunk_time_interval => INTERVAL '1 hour',
  if_not_exists       => TRUE
);

SELECT add_retention_policy('tnt_app_metric', INTERVAL '14 days',
  if_not_exists => TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS meta_idx_tnt_app_metric;
DROP TABLE IF EXISTS tnt_app_metric;

DROP INDEX IF EXISTS meta_idx_tnt_system_metric;
DROP TABLE IF EXISTS tnt_system_metric;
-- +goose StatementEnd
