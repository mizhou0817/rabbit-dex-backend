-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_market_last_archive_timestamp
(
    id                text PRIMARY KEY,
    archive_timestamp bigint NOT NULL
);

CREATE OR REPLACE FUNCTION replace_app_market_last_archive_timestamp()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_market_last_archive_timestamp(id, archive_timestamp)
    VALUES (NEW.id, NEW.archive_timestamp)
    ON CONFLICT (id) DO NOTHING;

    UPDATE app_market_last_archive_timestamp
    SET archive_timestamp=NEW.archive_timestamp
    WHERE id = NEW.id
      AND archive_timestamp < NEW.archive_timestamp;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_market_last_archive_timestamp_replace ON app_market;

CREATE TRIGGER app_market_last_archive_timestamp_replace
    BEFORE INSERT
    ON app_market
    FOR EACH ROW
EXECUTE PROCEDURE replace_app_market_last_archive_timestamp();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS app_market_last_archive_timestamp_replace ON app_market;
DROP FUNCTION IF EXISTS replace_app_market_last_archive_timestamp;
DROP TABLE IF EXISTS app_market_last_archive_timestamp;
-- +goose StatementEnd
