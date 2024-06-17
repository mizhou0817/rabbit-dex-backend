-- +goose Up
-- +goose StatementBegin
-- copy of hypertable app_inv3_data partitioned by last_updated
CREATE OR REPLACE FUNCTION interval_to_micros(i TEXT) returns BIGINT
    LANGUAGE SQL
    STABLE AS
$$
SELECT EXTRACT(EPOCH FROM i::interval)::bigint * 1000000
$$;

CREATE TABLE IF NOT EXISTS app_inv3_data_by_last_updated
(
    id                BIGINT  NOT NULL,
    valid             BOOLEAN NOT NULL,
    last_updated      BIGINT  NOT NULL,
    margined_ae_sum   NUMERIC NOT NULL,
    exchange_balance  NUMERIC NOT NULL,
    insurance_balance NUMERIC NOT NULL,
    shard_id          TEXT    NOT NULL,
    archive_id        BIGINT  NOT NULL,
    archive_timestamp BIGINT  NOT NULL,
    PRIMARY KEY (id, last_updated)
);

CREATE OR REPLACE FUNCTION insert_app_inv3_data_by_last_updated()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_inv3_data_by_last_updated(id, valid, last_updated, margined_ae_sum, exchange_balance,
                                              insurance_balance, shard_id, archive_id, archive_timestamp)
    VALUES (NEW.id, NEW.valid, NEW.last_updated, NEW.margined_ae_sum, NEW.exchange_balance, NEW.insurance_balance,
            NEW.shard_id, NEW.archive_id, NEW.archive_timestamp)
    ON CONFLICT (id, last_updated) DO NOTHING;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_inv3_data_by_last_updated_insert ON app_inv3_data;

CREATE TRIGGER app_inv3_data_by_last_updated_insert
    BEFORE INSERT
    ON app_inv3_data
    FOR EACH ROW
EXECUTE PROCEDURE insert_app_inv3_data_by_last_updated();

SELECT create_hypertable('app_inv3_data_by_last_updated', 'last_updated',
                         chunk_time_interval => 86400000000,
                         if_not_exists => TRUE,
                         migrate_data => TRUE
       );

CREATE OR REPLACE FUNCTION backfill_app_inv3_data_by_last_updated(job_id int, config jsonb)
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

    IF to_ts < stop_ts THEN
        PERFORM delete_job(job_id);
        RETURN;
    END IF;

    INSERT INTO app_inv3_data_by_last_updated(id, valid, last_updated, margined_ae_sum, exchange_balance,
                                              insurance_balance,
                                              shard_id, archive_id, archive_timestamp)
        (SELECT id,
                valid,
                last_updated,
                margined_ae_sum,
                exchange_balance,
                insurance_balance,
                shard_id,
                archive_id,
                archive_timestamp
         FROM app_inv3_data
         WHERE archive_timestamp >= from_ts
           AND archive_timestamp <= to_ts)
    ON CONFLICT (id, last_updated) DO NOTHING;

    PERFORM alter_job(job_id, config => json_build_object('stop_ts', stop_ts, 'current_ts', from_ts)::jsonb);
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'backfill_app_inv3_data_by_last_updated',
               '1 minute',
               config => json_build_object(
                       'stop_ts',
                       (select COALESCE(unix_now(), min(archive_timestamp))
                        FROM app_inv3_data),
                       'current_ts',
                       unix_now())::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'backfill_app_inv3_data_by_last_updated'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS backfill_app_inv3_data_by_last_updated;
DROP TRIGGER IF EXISTS app_inv3_data_by_last_updated_insert ON app_inv3_data;
DROP FUNCTION IF EXISTS insert_app_inv3_data_by_last_updated;
DROP TABLE IF EXISTS app_inv3_data_by_last_updated;
-- +goose StatementEnd
