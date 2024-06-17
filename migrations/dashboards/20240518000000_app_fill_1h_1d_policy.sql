-- +goose Up
-- +goose StatementBegin
SELECT add_continuous_aggregate_policy('app_fill_1h',
    start_offset => (EXTRACT(EPOCH FROM INTERVAL '3 hours') * 1000000)::bigint,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 hour') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('app_fill_1d',
    start_offset => (EXTRACT(EPOCH FROM INTERVAL '3 days') * 1000000)::bigint,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);
-- +goose StatementEnd
-- +goose NO TRANSACTION
CALL refresh_continuous_aggregate('app_fill_1h', 0::bigint, (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint);
CALL refresh_continuous_aggregate('app_fill_1d', 0::bigint, (EXTRACT(EPOCH FROM NOW()) * 1000000)::bigint);

-- +goose Down
-- +goose StatementBegin
SELECT remove_continuous_aggregate_policy('app_fill_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_fill_1h', if_exists => TRUE);
-- +goose StatementEnd
