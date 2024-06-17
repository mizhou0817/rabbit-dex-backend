-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_referral_counter
(
    profile_id      BIGINT PRIMARY KEY NOT NULL,
    invited_counter BIGINT NOT NULL DEFAULT 0
);

CREATE OR REPLACE FUNCTION updater_app_referral_counter() RETURNS TRIGGER AS
$$
DECLARE
BEGIN
    IF TG_OP = 'INSERT' THEN
        EXECUTE format('INSERT INTO app_referral_counter AS c(profile_id, invited_counter) VALUES(%L, 1) ON CONFLICT(profile_id) DO UPDATE SET invited_counter = c.invited_counter + 1', NEW.profile_id);
        RETURN NEW;
    ELSIF TG_OP = 'DELETE' THEN
        EXECUTE  format('UPDATE app_referral_counter SET invited_counter = invited_counter - 1 WHERE profile_id = %L', OLD.profile_id);
        RETURN OLD;
    END IF;
END;
$$
    LANGUAGE 'plpgsql';


CREATE OR REPLACE TRIGGER app_referral_link_trigger
    BEFORE INSERT OR DELETE
    ON app_referral_link
    FOR EACH ROW
EXECUTE PROCEDURE updater_app_referral_counter();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_referral_counter;
DROP FUNCTION IF EXISTS updater_app_referral_counter CASCADE;
DROP TRIGGER IF EXISTS app_referral_link_trigger ON app_referral_link CASCADE;
-- +goose StatementEnd
