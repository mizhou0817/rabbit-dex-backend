-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION last_shard_archive_id(schema text, hypertable text, shard_id text)
    RETURNS bigint
    LANGUAGE plpgsql
    PARALLEL SAFE
AS $$
DECLARE
    chunkTable regclass;
    result bigint;
BEGIN
    SELECT (c.chunk_schema || '.' || c.chunk_name)::regclass, c.*
    FROM timescaledb_information.chunks c
    WHERE c.hypertable_schema = schema
      AND c.hypertable_name = hypertable
    ORDER BY c.range_end_integer DESC
    LIMIT 1
    INTO chunkTable;

    IF chunkTable IS NULL THEN
        RETURN 0;
    END IF;

    EXECUTE format('SELECT MAX(archive_id) FROM %s WHERE shard_id=%L', chunkTable, shard_id) INTO result;
    IF result IS NOT NULL THEN
        RETURN result;
    END IF;

    EXECUTE format('SELECT COALESCE(MAX(archive_id), 0) FROM %s.%s WHERE shard_id=%L', schema, hypertable, shard_id) INTO result;
    RETURN result;
END;
$$;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS last_archive_id;
-- +goose StatementEnd
