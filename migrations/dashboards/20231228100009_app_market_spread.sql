-- +goose Up
-- +goose StatementBegin

-- historical market
CREATE TABLE IF NOT EXISTS app_market_spread
(
    id               TEXT   NOT NULL,
    last_update_time BIGINT NOT NULL,
    market_spread    FLOAT  NOT NULL,
    PRIMARY KEY (id, last_update_time)
);

CREATE OR REPLACE FUNCTION insert_app_market_spread()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_market_spread(id, last_update_time, market_spread)
    VALUES (NEW.id, NEW.last_update_time, (NEW.best_ask - NEW.best_bid) / NEW.fair_price)
    ON CONFLICT (id, last_update_time) DO NOTHING;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_market_spread_insert ON app_market;

CREATE TRIGGER app_market_spread_insert
    BEFORE INSERT
    ON app_market
    FOR EACH ROW
EXECUTE PROCEDURE insert_app_market_spread();

SELECT create_hypertable('app_market_spread', 'last_update_time',
                         chunk_time_interval => 86400000000,
                         if_not_exists => TRUE,
                         migrate_data => TRUE
       );

CREATE OR REPLACE FUNCTION backfill_app_market_spread(job_id int, config jsonb)
    RETURNS VOID AS
$$
DECLARE
    from_ts bigint;
    to_ts   bigint;
    stop_ts bigint;
BEGIN
    SELECT (config ->> 'stop_ts')::bigint INTO stop_ts;
    SELECT (config ->> 'current_ts')::bigint INTO to_ts;
    SELECT (to_ts - interval_to_micros('6 hour')) INTO from_ts;

    IF to_ts < stop_ts THEN
        PERFORM delete_job(job_id);
        RETURN;
    END IF;

    INSERT INTO app_market_spread(id, last_update_time, market_spread)
        (SELECT id, last_update_time, (best_ask - best_bid) / fair_price
         FROM app_market
         WHERE archive_timestamp >= from_ts
           AND archive_timestamp <= to_ts)
    ON CONFLICT (id, last_update_time) DO NOTHING;

    PERFORM alter_job(job_id, config => json_build_object('stop_ts', stop_ts, 'current_ts', from_ts)::jsonb);
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'backfill_app_market_spread',
               '1 minute',
               config => json_build_object(
                       'stop_ts',
                       (select COALESCE(unix_now(), min(archive_timestamp))
                        FROM app_market),
                       'current_ts',
                       unix_now())::jsonb
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'backfill_app_market_spread'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS backfill_app_market_spread;
DROP TRIGGER IF EXISTS app_market_spread_insert ON app_market;
DROP FUNCTION IF EXISTS insert_app_market_spread;
DROP TABLE IF EXISTS app_market_spread;
-- +goose StatementEnd
