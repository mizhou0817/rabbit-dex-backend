-- +goose Up
-- +goose StatementBegin
-- historical profile cache
CREATE TABLE IF NOT EXISTS app_profile_cache_1d
(
    id                BIGINT  NOT NULL,
    archive_timestamp BIGINT  NOT NULL,
    account_equity    NUMERIC NOT NULL,
    primary key (id, archive_timestamp)
);

CREATE INDEX IF NOT EXISTS app_profile_cache_1d_archive_timestamp_idx
    ON app_profile_cache_1d (archive_timestamp);

CREATE OR REPLACE FUNCTION upsert_app_profile_cache_1d()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_profile_cache_1d(id, archive_timestamp, account_equity)
    VALUES (NEW.id, time_bucket(interval_to_micros('1 day'), NEW.archive_timestamp), NEW.account_equity)
    ON CONFLICT (id, archive_timestamp) DO UPDATE
        SET account_equity = EXCLUDED.account_equity;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_profile_cache_1d_trigger ON app_profile_cache;

CREATE TRIGGER app_profile_cache_1d_trigger
    BEFORE INSERT
    ON app_profile_cache
    FOR EACH ROW
EXECUTE PROCEDURE upsert_app_profile_cache_1d();

CREATE OR REPLACE FUNCTION clean_app_profile_cache_1d(job_id int, config jsonb)
    RETURNS VOID AS
$$
BEGIN
    DELETE
    FROM app_profile_cache_1d
    WHERE archive_timestamp < unix_now() - interval_to_micros('366 days');
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'clean_app_profile_cache_1d',
               '1 day'
       );

CREATE OR REPLACE FUNCTION backfill_app_profile_cache_1d(job_id int, config jsonb)
    RETURNS VOID AS
$$
DECLARE
    from_ts bigint;
    to_ts   bigint;
    stop_ts bigint;
BEGIN
    SELECT (config ->> 'stop_ts')::bigint INTO stop_ts;
    SELECT (config ->> 'current_ts')::bigint INTO to_ts;
    SELECT (to_ts - interval_to_micros('1 day')) INTO from_ts;

    IF from_ts < stop_ts THEN
        PERFORM delete_job(job_id);
        RETURN;
    END IF;

    WITH cache_aggregate AS (
        -- with begin
        select id,
               time_bucket(interval_to_micros('1 day'), archive_timestamp) as archive_timestamp,
               LAST(account_equity, archive_timestamp)                     as account_equity
        FROM app_profile_cache
        WHERE archive_timestamp >= from_ts
          AND archive_timestamp < to_ts
        GROUP BY 1, 2
        -- with end
    )
    INSERT
    INTO app_profile_cache_1d(id, archive_timestamp, account_equity)
            (SELECT id, archive_timestamp, account_equity from cache_aggregate)
    ON CONFLICT (id, archive_timestamp) DO UPDATE
        SET account_equity = EXCLUDED.account_equity;

    PERFORM alter_job(job_id, config => json_build_object('stop_ts', stop_ts, 'current_ts', from_ts)::jsonb);
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'backfill_app_profile_cache_1d',
               '1 minute',
               config => json_build_object(
                       'stop_ts',
                       time_bucket(interval_to_micros('1 day'), unix_now() - interval_to_micros('366 days')),
                       'current_ts',
                       time_bucket(interval_to_micros('1 day'), unix_now()))::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'backfill_app_profile_cache_1d'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS backfill_app_profile_cache_1d;
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'clean_app_profile_cache_1d'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS clean_app_profile_cache_1d;
DROP TRIGGER IF EXISTS app_profile_cache_1d_trigger ON app_profile_cache;
DROP FUNCTION IF EXISTS upsert_app_profile_cache_1d;
DROP TABLE IF EXISTS app_profile_cache_1d;
-- +goose StatementEnd
