-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_vault_aggregate_last
(
    vault_profile_id  BIGINT PRIMARY KEY,
    share_price       NUMERIC NOT NULL,
    apy_total         NUMERIC NOT NULL,
    apy_usdt          NUMERIC NOT NULL,
    apy_rbx           NUMERIC NOT NULL,
    archive_timestamp BIGINT  NOT NULL
);

CREATE TABLE IF NOT EXISTS app_vault_aggregate_history
(
    vault_profile_id  BIGINT,
    share_price       NUMERIC NOT NULL,
    apy_total         NUMERIC NOT NULL,
    apy_usdt          NUMERIC NOT NULL,
    apy_rbx           NUMERIC NOT NULL,
    archive_timestamp BIGINT  NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS app_vault_aggregate_history_idx
    ON app_vault_aggregate_history (vault_profile_id, archive_timestamp);

SELECT create_hypertable('app_vault_aggregate_history', 'archive_timestamp',
                         chunk_time_interval => 3600000000,
                         if_not_exists => TRUE
       );

CREATE OR REPLACE FUNCTION insert_app_vault_aggregate_history()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_vault_aggregate_history(vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx,
                                            archive_timestamp)
        (SELECT vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp
         FROM app_vault_aggregate_last
         WHERE vault_profile_id = NEW.vault_profile_id)
    ON CONFLICT (vault_profile_id, archive_timestamp) DO UPDATE
        SET share_price = EXCLUDED.share_price,
            apy_total   = EXCLUDED.apy_total,
            apy_usdt    = EXCLUDED.apy_usdt,
            apy_rbx     = EXCLUDED.apy_rbx;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_vault_aggregate_last_insert ON app_vault_aggregate_last;

CREATE TRIGGER app_vault_aggregate_last_insert
    AFTER INSERT OR UPDATE
    ON app_vault_aggregate_last
    FOR EACH ROW
EXECUTE PROCEDURE insert_app_vault_aggregate_history();


CREATE OR REPLACE FUNCTION refresh_app_vault_aggregate_last(job_id int, config jsonb)
    RETURNS VOID AS
$$
BEGIN
    WITH vault_aggregate AS (
        -- with begin
        SELECT v.vault_profile_id,
               c.account_equity / v.total_shares as share_price,
               0                                 as apy_total,
               0                                 as apy_usdt,
               0                                 as apy_rbx,
               c.archive_timestamp               as archive_timestamp
        FROM app_vault_last as v
                 JOIN app_profile_cache_last as c on c.id = v.vault_profile_id
        WHERE c.archive_timestamp >= v.archive_timestamp
          AND v.total_shares > 0
          AND c.account_equity > 0
          AND unix_now() > v.initialised_at
        -- with end
    )
    INSERT
    INTO app_vault_aggregate_last(vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp)
        (SELECT vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp from vault_aggregate)
    ON CONFLICT(vault_profile_id) DO UPDATE
        SET share_price       = EXCLUDED.share_price,
            archive_timestamp = EXCLUDED.archive_timestamp;

    WITH vault_aggregate AS (
        -- with begin
        SELECT v.vault_profile_id,
               c.account_equity / v.total_shares as share_price,
               c.archive_timestamp               as archive_timestamp
        FROM app_vault_last as v
                 JOIN app_profile_cache_last as c on c.id = v.vault_profile_id
        WHERE c.archive_timestamp >= v.archive_timestamp
          AND v.total_shares > 0
          AND c.account_equity > 0
          AND unix_now() > v.initialised_at
          AND GREATEST(0, (((c.account_equity / v.total_shares) - 1) *
                           (365 / (1 + (unix_now() - v.initialised_at) / 86400000000)))) <= 0
        -- with end
    )
    INSERT
    INTO app_vault_aggregate_last(vault_profile_id, share_price, archive_timestamp)
        (SELECT vault_profile_id, share_price, archive_timestamp from vault_aggregate)
    ON CONFLICT(vault_profile_id) DO UPDATE
        SET share_price       = EXCLUDED.share_price,
            archive_timestamp = EXCLUDED.archive_timestamp;

    WITH vault_aggregate AS (
        -- with begin
        SELECT v.vault_profile_id,
               c.account_equity / v.total_shares                                          as share_price,
               unix_now()                                                                 as now,
               v.initialised_at                                                           as initialised_at,
               (1 + (unix_now() - v.initialised_at) / 86400000000)                        as days,
               (((c.account_equity / v.total_shares) - 1) *
                (365 / (1 + (unix_now() - v.initialised_at) / 86400000000)))              as apy_total_raw,
               GREATEST(0, (((c.account_equity / v.total_shares) - 1) *
                            (365 / (1 + (unix_now() - v.initialised_at) / 86400000000)))) as apy_total,
               GREATEST(0, (((c.account_equity / v.total_shares) - 1) *
                            (365 / (1 + (unix_now() - v.initialised_at) / 86400000000)))) as apy_usdt,
               0                                                                          as apy_rbx,
               c.archive_timestamp                                                        as archive_timestamp
        FROM app_vault_last as v
                 JOIN app_profile_cache_last as c on c.id = v.vault_profile_id
        WHERE c.archive_timestamp >= v.archive_timestamp
          AND v.total_shares > 0
          AND c.account_equity > 0
          AND unix_now() > v.initialised_at
          AND GREATEST(0, (((c.account_equity / v.total_shares) - 1) *
                           (365 / (1 + (unix_now() - v.initialised_at) / 86400000000)))) > 0
        -- with end
    )
    INSERT
    INTO app_vault_aggregate_last(vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp)
        (SELECT vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp from vault_aggregate)
    ON CONFLICT(vault_profile_id) DO UPDATE
        SET share_price       = EXCLUDED.share_price,
            apy_total         = EXCLUDED.apy_total,
            apy_usdt          = EXCLUDED.apy_usdt,
            apy_rbx           = EXCLUDED.apy_rbx,
            archive_timestamp = EXCLUDED.archive_timestamp;

    WITH vault_aggregate AS (
        -- with begin
        SELECT v.vault_profile_id,
               1                   as share_price,
               0                   as apy_total,
               0                   as apy_usdt,
               0                   as apy_rbx,
               c.archive_timestamp as archive_timestamp
        FROM app_vault_last as v
                 JOIN app_profile_cache_last as c on c.id = v.vault_profile_id
        WHERE v.vault_profile_id NOT IN (SELECT vault_profile_id FROM app_vault_aggregate_last)
        -- with end
    )
    INSERT
    INTO app_vault_aggregate_last(vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp)
        (SELECT vault_profile_id, share_price, apy_total, apy_usdt, apy_rbx, archive_timestamp from vault_aggregate)
    ON CONFLICT(vault_profile_id) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
               'refresh_app_vault_aggregate_last',
               '10 min'
       );
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
SELECT delete_job(
               (SELECT job_id
                FROM timescaledb_information.jobs
                WHERE proc_name = 'refresh_app_vault_aggregate_last'
                LIMIT 1)
       );
DROP FUNCTION IF EXISTS refresh_app_vault_aggregate_last;
DROP TRIGGER IF EXISTS app_vault_aggregate_last_insert ON app_vault_aggregate_last;
DROP FUNCTION IF EXISTS insert_app_vault_aggregate_history;
DROP TABLE IF EXISTS app_vault_aggregate_history;
DROP TABLE IF EXISTS app_vault_aggregate_last;
-- +goose StatementEnd
