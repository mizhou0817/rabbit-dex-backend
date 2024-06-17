-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_profile_cache_last
(
    id                    BIGINT PRIMARY KEY,
    profile_type          TEXT    NOT NULL,
    status                TEXT    NOT NULL,
    wallet                TEXT    NOT NULL,
    last_update           BIGINT  NOT NULL,
    balance               NUMERIC NOT NULL,
    account_equity        NUMERIC NOT NULL,
    total_position_margin NUMERIC NOT NULL,
    total_order_margin    NUMERIC NOT NULL,
    total_notional        NUMERIC NOT NULL,
    account_margin        NUMERIC NOT NULL,
    withdrawable_balance  NUMERIC NOT NULL,
    cum_unrealized_pnl    NUMERIC NOT NULL,
    health                NUMERIC NOT NULL,
    account_leverage      NUMERIC NOT NULL,
    cum_trading_volume    NUMERIC NOT NULL,
    leverage              JSONB   NOT NULL,
    last_liq_check        BIGINT  NOT NULL,
    shard_id              TEXT    NOT NULL,
    archive_id            BIGINT  NOT NULL,
    archive_timestamp     BIGINT  NOT NULL
);

CREATE INDEX IF NOT EXISTS app_profile_cache_last_account_equity_idx
    ON app_profile_cache_last (account_equity);

CREATE INDEX IF NOT EXISTS app_profile_cache_last_account_margin_idx
    ON app_profile_cache_last (account_margin);

CREATE OR REPLACE FUNCTION replace_app_profile_cache_last()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    INSERT INTO app_profile_cache_last(id, profile_type, status, wallet, last_update, balance, account_equity,
                                       total_position_margin, total_order_margin, total_notional, account_margin,
                                       withdrawable_balance, cum_unrealized_pnl, health, account_leverage,
                                       cum_trading_volume, leverage, last_liq_check, shard_id, archive_id,
                                       archive_timestamp)
    VALUES (NEW.id, NEW.profile_type, NEW.status, NEW.wallet, NEW.last_update, NEW.balance, NEW.account_equity,
            NEW.total_position_margin, NEW.total_order_margin, NEW.total_notional, NEW.account_margin,
            NEW.withdrawable_balance, NEW.cum_unrealized_pnl, NEW.health, NEW.account_leverage, NEW.cum_trading_volume,
            NEW.leverage, NEW.last_liq_check, NEW.shard_id, NEW.archive_id, NEW.archive_timestamp)
    ON CONFLICT (id) DO NOTHING;

    UPDATE app_profile_cache_last
    SET profile_type=NEW.profile_type,
        status=NEW.status,
        wallet=NEW.wallet,
        last_update=NEW.last_update,
        balance=NEW.balance,
        account_equity=NEW.account_equity,
        total_position_margin=NEW.total_position_margin,
        total_order_margin=NEW.total_order_margin,
        total_notional=NEW.total_notional,
        account_margin=NEW.account_margin,
        withdrawable_balance=NEW.withdrawable_balance,
        cum_unrealized_pnl=NEW.cum_unrealized_pnl,
        health=NEW.health,
        account_leverage=NEW.account_leverage,
        cum_trading_volume=NEW.cum_trading_volume,
        leverage=NEW.leverage,
        last_liq_check=NEW.last_liq_check,
        shard_id=NEW.shard_id,
        archive_id=NEW.archive_id,
        archive_timestamp=NEW.archive_timestamp
    WHERE id = NEW.id
      AND archive_timestamp < NEW.archive_timestamp;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_profile_cache_last_replace ON app_profile_cache;

CREATE TRIGGER app_profile_cache_last_replace
    BEFORE INSERT
    ON app_profile_cache
    FOR EACH ROW
EXECUTE PROCEDURE replace_app_profile_cache_last();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS app_profile_cache_last_replace ON app_profile_cache;
DROP FUNCTION IF EXISTS replace_app_profile_cache_last;
DROP TABLE IF EXISTS app_profile_cache_last;
-- +goose StatementEnd
