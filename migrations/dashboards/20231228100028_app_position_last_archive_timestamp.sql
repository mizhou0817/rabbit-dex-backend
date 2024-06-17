-- +goose Up
-- +goose StatementBegin
-- last archived app position per market id
CREATE TABLE IF NOT EXISTS app_position_last_archive_timestamp
(
    market_id         text PRIMARY KEY,
    archive_timestamp bigint NOT NULL
);

CREATE OR REPLACE FUNCTION replace_app_position_last_archive_timestamp()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_position_last_archive_timestamp(market_id, archive_timestamp)
    VALUES (NEW.market_id, NEW.archive_timestamp)
    ON CONFLICT (market_id) DO NOTHING;

    UPDATE app_position_last_archive_timestamp
    SET archive_timestamp=NEW.archive_timestamp
    WHERE market_id = NEW.market_id
      AND archive_timestamp < NEW.archive_timestamp;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_position_last_archive_timestamp_replace ON app_position;

CREATE TRIGGER app_position_last_archive_timestamp_replace
    BEFORE INSERT
    ON app_position
    FOR EACH ROW
EXECUTE PROCEDURE replace_app_position_last_archive_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS app_position_last_archive_timestamp_replace ON app_position;
DROP FUNCTION IF EXISTS replace_app_position_last_archive_timestamp;
DROP TABLE IF EXISTS app_position_last_archive_timestamp;
-- +goose StatementEnd
