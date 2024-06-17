-- +goose Up
-- +goose StatementBegin
SELECT add_continuous_aggregate_policy('analytic_event_wallet_1d',
    start_offset => NULL,
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('analytic_event_wallet_1w',
    start_offset => NULL,
    end_offset => INTERVAL '1 week',
    schedule_interval => INTERVAL '1 week',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('analytic_event_order_1d',
    start_offset => NULL,
    end_offset => INTERVAL '1 day',
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('analytic_event_order_1w',
    start_offset => NULL,
    end_offset => INTERVAL '1 week',
    schedule_interval => INTERVAL '1 week',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_fill_1w',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 week') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 week',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_balance_operation_net_deposit_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_fill_profile_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_trade_market_1h',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 hour') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 hour',
    if_not_exists => TRUE
);

SELECT add_continuous_aggregate_policy('app_trade_market_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_balance_operation_pnl_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_trade_volume_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);


SELECT add_continuous_aggregate_policy('app_fill_fee_1d',
    start_offset => NULL,
    end_offset => (EXTRACT(EPOCH FROM INTERVAL '1 day') * 1000000)::bigint,
    schedule_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT remove_continuous_aggregate_policy('app_fill_fee_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_trade_volume_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_balance_operation_pnl_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_trade_market_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_trade_market_1h', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_fill_profile_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_balance_operation_net_deposit_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('app_fill_1w', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('analytic_event_order_1w', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('analytic_event_order_1d', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('analytic_event_wallet_1w', if_exists => TRUE);
SELECT remove_continuous_aggregate_policy('analytic_event_wallet_1d', if_exists => TRUE);
-- +goose StatementEnd
