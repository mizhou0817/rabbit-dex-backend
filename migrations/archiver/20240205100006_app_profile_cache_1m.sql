-- +goose Up
-- +goose StatementBegin
-- historical profile cache
CREATE OR REPLACE FUNCTION interval_to_micros(i TEXT) returns BIGINT
    LANGUAGE SQL
    STABLE AS
$$
SELECT EXTRACT(EPOCH FROM i::interval)::bigint * 1000000
$$;

CREATE TABLE IF NOT EXISTS app_profile_cache_1m
(
    id                BIGINT  NOT NULL,
    archive_timestamp BIGINT  NOT NULL,
    account_equity    NUMERIC NOT NULL,
    primary key (id, archive_timestamp)
);

CREATE INDEX IF NOT EXISTS app_profile_cache_1m_archive_timestamp_idx
    ON app_profile_cache_1m (archive_timestamp);

CREATE OR REPLACE FUNCTION upsert_app_profile_cache_1m()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_profile_cache_1m(id, archive_timestamp, account_equity)
    VALUES (NEW.id, time_bucket(interval_to_micros('1 minute'), NEW.archive_timestamp), NEW.account_equity)
    ON CONFLICT (id, archive_timestamp) DO UPDATE
        SET account_equity = EXCLUDED.account_equity;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_profile_cache_1m_trigger ON app_profile_cache;

CREATE TRIGGER app_profile_cache_1m_trigger
    BEFORE INSERT
    ON app_profile_cache
    FOR EACH ROW
EXECUTE PROCEDURE upsert_app_profile_cache_1m();

CREATE OR REPLACE FUNCTION clean_app_profile_cache_1m(job_id int, config jsonb)
    RETURNS VOID AS
$$
BEGIN
    DELETE
    FROM app_profile_cache_1m
    WHERE archive_timestamp < unix_now() - interval_to_micros('1 hour');
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'clean_app_profile_cache_1m',
               '1 hour'
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'clean_app_profile_cache_1m'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS clean_app_profile_cache_1m;
DROP TRIGGER IF EXISTS app_profile_cache_1m_trigger ON app_profile_cache;
DROP FUNCTION IF EXISTS upsert_app_profile_cache_1m;
DROP TABLE IF EXISTS app_profile_cache_1m;
-- +goose StatementEnd
