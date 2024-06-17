-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS market_id_view
(
    market_id text PRIMARY KEY
);

CREATE OR REPLACE FUNCTION insert_market_id_view()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO market_id_view(market_id)
    VALUES (NEW.id)
    ON CONFLICT (market_id) DO NOTHING;
    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS market_id_view_insert ON app_market;

CREATE TRIGGER market_id_view_insert
    BEFORE INSERT
    ON app_market
    FOR EACH ROW
EXECUTE PROCEDURE insert_market_id_view();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS market_id_view_insert ON app_market;
DROP FUNCTION IF EXISTS insert_market_id_view;
DROP TABLE IF EXISTS market_id_view;
-- +goose StatementEnd