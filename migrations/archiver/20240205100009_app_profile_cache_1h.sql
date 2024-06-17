-- +goose Up
-- +goose StatementBegin
-- historical profile cache
CREATE TABLE IF NOT EXISTS app_profile_cache_1h
(
    id                BIGINT  NOT NULL,
    archive_timestamp BIGINT  NOT NULL,
    account_equity    NUMERIC NOT NULL,
    primary key (id, archive_timestamp)
);

CREATE INDEX IF NOT EXISTS app_profile_cache_1h_archive_timestamp_idx
    ON app_profile_cache_1h (archive_timestamp);

CREATE OR REPLACE FUNCTION upsert_app_profile_cache_1h()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_profile_cache_1h(id, archive_timestamp, account_equity)
    VALUES (NEW.id, time_bucket(interval_to_micros('1 hour'), NEW.archive_timestamp), NEW.account_equity)
    ON CONFLICT (id, archive_timestamp) DO UPDATE
        SET account_equity = EXCLUDED.account_equity;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_profile_cache_1h_trigger ON app_profile_cache;

CREATE TRIGGER app_profile_cache_1h_trigger
    BEFORE INSERT
    ON app_profile_cache
    FOR EACH ROW
EXECUTE PROCEDURE upsert_app_profile_cache_1h();

CREATE OR REPLACE FUNCTION clean_app_profile_cache_1h(job_id int, config jsonb)
    RETURNS VOID AS
$$
BEGIN
    DELETE
    FROM app_profile_cache_1h
    WHERE archive_timestamp < unix_now() - interval_to_micros('29 days');
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'clean_app_profile_cache_1h',
               '1 day'
       );

CREATE OR REPLACE FUNCTION backfill_app_profile_cache_1h(job_id int, config jsonb)
    RETURNS VOID AS
$$
DECLARE
    from_ts bigint;
    to_ts   bigint;
    stop_ts bigint;
BEGIN
    SELECT (config ->> 'stop_ts')::bigint INTO stop_ts;
    SELECT (config ->> 'current_ts')::bigint INTO to_ts;
    SELECT (to_ts - interval_to_micros('1 hour')) INTO from_ts;

    IF from_ts < stop_ts THEN
        PERFORM delete_job(job_id);
        RETURN;
    END IF;

    WITH cache_aggregate AS (
        -- with begin
        select id,
               time_bucket(interval_to_micros('1 hour'), archive_timestamp) as archive_timestamp,
               LAST(account_equity, archive_timestamp)                      as account_equity
        FROM app_profile_cache
        WHERE archive_timestamp >= from_ts
          AND archive_timestamp < to_ts
        GROUP BY 1, 2
        -- with end
    )
    INSERT
    INTO app_profile_cache_1h(id, archive_timestamp, account_equity)
            (SELECT id, archive_timestamp, account_equity from cache_aggregate)
    ON CONFLICT (id, archive_timestamp) DO UPDATE
        SET account_equity = EXCLUDED.account_equity;

    PERFORM alter_job(job_id, config => json_build_object('stop_ts', stop_ts, 'current_ts', from_ts)::jsonb);
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'backfill_app_profile_cache_1h',
               '1 minute',
               config => json_build_object(
                       'stop_ts',
                       time_bucket(interval_to_micros('1 hour'), unix_now() - interval_to_micros('29 days')),
                       'current_ts',
                       time_bucket(interval_to_micros('1 hour'), unix_now()))::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'backfill_app_profile_cache_1h'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS backfill_app_profile_cache_1h;
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'clean_app_profile_cache_1h'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS clean_app_profile_cache_1h;
DROP TRIGGER IF EXISTS app_profile_cache_1h_trigger ON app_profile_cache;
DROP FUNCTION IF EXISTS upsert_app_profile_cache_1h;
DROP TABLE IF EXISTS app_profile_cache_1h;
-- +goose StatementEnd
