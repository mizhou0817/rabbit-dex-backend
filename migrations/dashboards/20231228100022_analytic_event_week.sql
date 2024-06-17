-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS analytic_event_week
(
    ts TIMESTAMPTZ PRIMARY KEY
);

CREATE OR REPLACE FUNCTION refresh_analytic_event_week(job_id int, config jsonb)
    RETURNS VOID AS
$$
DECLARE
    last_ts TIMESTAMPTZ;
    next_ts TIMESTAMPTZ;
BEGIN
    SELECT (config ->> 'last_ts')::TIMESTAMPTZ INTO last_ts;
    SELECT time_bucket('1week', last_ts + interval '1week')
    INTO next_ts;

    IF next_ts < time_bucket('1week', now()) THEN
        INSERT INTO analytic_event_week(ts)
        VALUES (last_ts);

        PERFORM alter_job(job_id, config => json_build_object('last_ts', next_ts)::jsonb);
    END IF;
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'refresh_analytic_event_week',
               '1 hour',
               fixed_schedule => false,
               config => json_build_object(
                       'last_ts',
                       (SELECT COALESCE(time_bucket('1week', MIN(ts)), time_bucket('1week', now())) FROM analytic_event)
                         )::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'refresh_analytic_event_week'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS refresh_analytic_event_week;
DROP TABLE IF EXISTS analytic_event_week;
-- +goose StatementEnd
