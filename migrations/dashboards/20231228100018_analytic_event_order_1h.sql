-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analytic_event_order_1h
(
    ts         TIMESTAMPTZ NOT NULL,
    profile_id BIGINT      NOT NULL,
    total      bigint      NOT NULL,
    PRIMARY KEY (ts, profile_id)
);

SELECT create_hypertable(
               'analytic_event_order_1h',
               'ts',
               if_not_exists => true,
               migrate_data => true
       );

CREATE OR REPLACE FUNCTION refresh_analytic_event_order_1h(job_id int, config jsonb)
    RETURNS VOID AS
$$
DECLARE
    last_ts TIMESTAMPTZ;
    next_ts TIMESTAMPTZ;
BEGIN
    SELECT (config ->> 'last_ts')::TIMESTAMPTZ INTO last_ts;
    SELECT time_bucket('1hour', last_ts + interval '1hour')
    INTO next_ts;

    IF next_ts < time_bucket('1hour', now()) THEN
        INSERT INTO analytic_event_order_1h(profile_id, total, ts)
        SELECT (event ->> 'profile_id')::BIGINT as profile_id,
               count(*)                         as total,
               last_ts
        FROM analytic_event
        WHERE ts >= last_ts
          AND ts < next_ts
          AND event ->> 'url_path' LIKE '/orders%'
          AND event ->> 'response_status_code' = '200'
        GROUP BY 1;

        PERFORM alter_job(job_id, config => json_build_object('last_ts', next_ts)::jsonb);
    END IF;
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'refresh_analytic_event_order_1h',
               '1 min',
               fixed_schedule => false,
               config => json_build_object(
                       'last_ts',
                       (SELECT COALESCE(time_bucket('1hour', MIN(ts)), time_bucket('1hour', now())) FROM analytic_event)
                         )::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'refresh_analytic_event_order_1h'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS refresh_analytic_event_order_1h;
DROP TABLE IF EXISTS analytic_event_order_1h;
-- +goose StatementEnd
