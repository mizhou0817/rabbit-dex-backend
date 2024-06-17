-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_referral_code
(
    profile_id        BIGINT                      NOT NULL PRIMARY KEY,
    short_code        TEXT                        NOT NULL UNIQUE,
    model             TEXT                        NOT NULL,
    model_fee_percent NUMERIC                     NULL CHECK (model_fee_percent > 0 AND model_fee_percent <= 0.70),
    amend_counter     INTEGER                     NOT NULL DEFAULT 0,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
CREATE INDEX IF NOT EXISTS app_referral_code_short_code_idx ON app_referral_code (short_code);
CREATE UNIQUE INDEX IF NOT EXISTS app_referral_code_short_code_key ON app_referral_code ((UPPER(short_code)));

CREATE TABLE IF NOT EXISTS app_referral_blacklist (
    profile_id        BIGINT                      NOT NULL PRIMARY KEY,
    timestamp         TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);

CREATE FUNCTION referral_check_blacklist() RETURNS trigger AS $func$
    BEGIN
        IF (SELECT EXISTS(SELECT 1 FROM app_referral_blacklist WHERE profile_id = NEW.profile_id)) THEN
            RAISE EXCEPTION 'blacklisted profile_id %', NEW.profile_id;
        END IF;
        RETURN NEW;
    END;
$func$ LANGUAGE plpgsql;

CREATE TRIGGER referral_blacklist BEFORE INSERT OR UPDATE ON app_referral_code
    FOR EACH ROW EXECUTE PROCEDURE referral_check_blacklist();


CREATE TABLE IF NOT EXISTS app_referral_link
(
    profile_id BIGINT                      NOT NULL,
    invited_id BIGINT                      NOT NULL,
    timestamp  TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    UNIQUE (invited_id)
);
CREATE INDEX IF NOT EXISTS app_referral_link_profile_id_idx ON app_referral_link (profile_id);
CREATE INDEX IF NOT EXISTS app_referral_link_invited_id_idx ON app_referral_link (invited_id);

CREATE TABLE IF NOT EXISTS app_referral_runner
(
    proc_name    TEXT   NOT NULL PRIMARY KEY,
    timestamp    TIMESTAMP WITHOUT TIME ZONE NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS app_referral_code;
DROP TABLE IF EXISTS app_referral_link;
DROP TABLE IF EXISTS app_referral_runner;
-- +goose StatementEnd
