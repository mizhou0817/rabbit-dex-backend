-- +goose Up
-- +goose StatementBegin
-- analytic_event
CREATE TABLE IF NOT EXISTS analytic_event (
    event JSONB NOT NULL,
    ts TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE INDEX IF NOT EXISTS event_idx_analytic_event ON analytic_event USING GIN(event);

SELECT create_hypertable('analytic_event', 'ts',
  chunk_time_interval => INTERVAL '1 month',
  if_not_exists       => TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analytic_event;
-- +goose StatementEnd
