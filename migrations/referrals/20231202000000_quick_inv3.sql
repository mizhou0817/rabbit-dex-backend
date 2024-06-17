-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION get_last_inv3_data()
    RETURNS TABLE
            (
                id                bigint,
                valid             boolean,
                last_updated      bigint,
                margined_ae_sum   numeric,
                exchange_balance  numeric,
                insurance_balance numeric,
                shard_id          text,
                archive_id        bigint,
                archive_timestamp bigint
            )
AS
$$
DECLARE
    chunkTable regclass;
BEGIN
    SELECT (c.chunk_schema || '.' || c.chunk_name)::regclass, c.*
    FROM timescaledb_information.chunks c
    WHERE c.hypertable_schema = 'public'
      AND c.hypertable_name = 'app_inv3_data'
    ORDER BY c.range_end_integer DESC
    LIMIT 1
    INTO chunkTable;

    RETURN QUERY
        EXECUTE format('SELECT * FROM %s ORDER BY archive_timestamp DESC LIMIT 1', chunkTable);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS get_last_inv3_data CASCADE;
-- +goose StatementEnd