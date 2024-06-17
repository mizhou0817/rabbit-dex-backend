-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS app_vault_balance_operation_last
(
    id                 TEXT PRIMARY KEY,
    ops_type           TEXT    NOT NULL DEFAULT '',
    ops_sub_type       TEXT    NOT NULL DEFAULT '',
    staker_profile_id  BIGINT  NOT NULL DEFAULT 0,
    vault_profile_id   BIGINT  NOT NULL DEFAULT 0,
    vault_wallet       TEXT    NOT NULL DEFAULT '',
    vault_exchange_id  TEXT    NOT NULL DEFAULT '',
    stake_usdt         NUMERIC NOT NULL DEFAULT 0,
    stake_shares       NUMERIC NOT NULL DEFAULT 0,
    unstake_shares     NUMERIC NOT NULL DEFAULT 0,
    unstake_usdt       NUMERIC NOT NULL DEFAULT 0,
    unstake_fee_usdt   NUMERIC NOT NULL DEFAULT 0,
    -- unstake_vault_usdt = unstake_usdt + unstake_fee_usdt
    unstake_vault_usdt NUMERIC NOT NULL DEFAULT 0,
    timestamp          BIGINT  NOT NULL DEFAULT 0,
    status             TEXT    NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS app_vault_balance_operation_last_idx
    ON app_vault_balance_operation_last (staker_profile_id, vault_profile_id, ops_type);

CREATE INDEX IF NOT EXISTS app_vault_balance_operation_last_wallet_idx
    ON app_vault_balance_operation_last (staker_profile_id, vault_wallet, vault_exchange_id, ops_type);

CREATE INDEX IF NOT EXISTS app_vault_balance_operation_last_ts_idx
    ON app_vault_balance_operation_last (staker_profile_id, timestamp);

CREATE OR REPLACE FUNCTION insert_app_vault_balance_operation_last()
    RETURNS TRIGGER
    LANGUAGE PLPGSQL
AS
$$
BEGIN
    IF NEW.ops_type = 'stake' THEN
        INSERT INTO app_vault_balance_operation_last(id, staker_profile_id, vault_wallet, vault_exchange_id, timestamp,
                                                     ops_type, ops_sub_type, stake_usdt, status)
        VALUES (NEW.id, NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.timestamp, 'stake', 'stake', NEW.amount,
                NEW.status)
        ON CONFLICT (id) DO UPDATE
            SET staker_profile_id = EXCLUDED.staker_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                timestamp         = EXCLUDED.timestamp,
                ops_type          = EXCLUDED.ops_type,
                ops_sub_type      = EXCLUDED.ops_sub_type,
                stake_usdt        = EXCLUDED.stake_usdt,
                status            = EXCLUDED.status;
    END IF;
    IF NEW.ops_type = 'stake_shares' THEN
        -- ss_
        INSERT INTO app_vault_balance_operation_last(id, staker_profile_id, vault_wallet, vault_exchange_id, stake_shares)
        VALUES (SUBSTRING(NEW.id, 4), NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET staker_profile_id = EXCLUDED.staker_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                stake_shares      = EXCLUDED.stake_shares;
    END IF;
    IF NEW.ops_type = 'vault_stake' THEN
        -- vs_s_
        INSERT INTO app_vault_balance_operation_last(id, vault_profile_id, vault_wallet, vault_exchange_id, stake_usdt)
        VALUES (SUBSTRING(NEW.id, 6), NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET vault_profile_id  = EXCLUDED.vault_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                stake_usdt        = EXCLUDED.stake_usdt;
    END IF;

    IF NEW.ops_type = 'stake_from_balance' THEN
        INSERT INTO app_vault_balance_operation_last(id, staker_profile_id, vault_wallet, vault_exchange_id, timestamp,
                                                     ops_type, ops_sub_type, stake_usdt, status)
        VALUES (NEW.id, NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.timestamp, 'stake', 'stake_from_balance',
                NEW.amount, NEW.status)
        ON CONFLICT (id) DO UPDATE
            SET staker_profile_id = EXCLUDED.staker_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                timestamp         = EXCLUDED.timestamp,
                ops_type          = EXCLUDED.ops_type,
                ops_sub_type      = EXCLUDED.ops_sub_type,
                stake_usdt        = EXCLUDED.stake_usdt,
                status            = EXCLUDED.status;
    END IF;

    IF NEW.ops_type = 'unstake_shares' THEN
        INSERT INTO app_vault_balance_operation_last(id, staker_profile_id, vault_wallet, vault_exchange_id, timestamp,
                                                     ops_type, ops_sub_type, unstake_shares, status)
        VALUES (NEW.id, NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.timestamp, 'unstake', 'unstake', NEW.amount,
                NEW.status)
        ON CONFLICT (id) DO UPDATE
            SET staker_profile_id = EXCLUDED.staker_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                timestamp         = EXCLUDED.timestamp,
                ops_type          = EXCLUDED.ops_type,
                ops_sub_type      = EXCLUDED.ops_sub_type,
                unstake_shares    = EXCLUDED.unstake_shares,
                status            = EXCLUDED.status;
    END IF;
    IF NEW.ops_type = 'vault_unstake_shares' THEN
        -- vus_
        INSERT INTO app_vault_balance_operation_last(id, vault_profile_id, vault_wallet, vault_exchange_id, unstake_shares)
        VALUES (SUBSTRING(NEW.id, 5), NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET vault_profile_id  = EXCLUDED.vault_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                unstake_shares    = EXCLUDED.unstake_shares;
    END IF;
    IF NEW.ops_type = 'unstake_value' THEN
        -- uv_
        INSERT INTO app_vault_balance_operation_last(id, staker_profile_id, vault_wallet, vault_exchange_id, unstake_usdt)
        VALUES (SUBSTRING(NEW.id, 4), NEW.profile_id, NEW.wallet, NEW.exchange_id, NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET staker_profile_id = EXCLUDED.staker_profile_id,
                vault_wallet      = EXCLUDED.vault_wallet,
                vault_exchange_id = EXCLUDED.vault_exchange_id,
                unstake_usdt      = EXCLUDED.unstake_usdt;
    END IF;
    IF NEW.ops_type = 'unstake_fee' THEN
        -- uf_
        INSERT INTO app_vault_balance_operation_last(id, unstake_fee_usdt)
        VALUES (SUBSTRING(NEW.id, 4), NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET unstake_fee_usdt = EXCLUDED.unstake_fee_usdt;
    END IF;
    IF NEW.ops_type = 'vault_unstake_value' THEN
        -- vuv_
        INSERT INTO app_vault_balance_operation_last(id, unstake_vault_usdt)
        VALUES (SUBSTRING(NEW.id, 5), NEW.amount)
        ON CONFLICT (id) DO UPDATE
            SET unstake_vault_usdt = EXCLUDED.unstake_vault_usdt;
    END IF;

    RETURN NEW;
END;
$$;

DROP TRIGGER IF EXISTS app_vault_balance_operation_last_insert ON app_balance_operation;

CREATE TRIGGER app_vault_balance_operation_last_insert
    BEFORE INSERT OR UPDATE
    ON app_balance_operation
    FOR EACH ROW
EXECUTE PROCEDURE insert_app_vault_balance_operation_last();
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS app_vault_balance_operation_last_insert ON app_balance_operation;
DROP FUNCTION IF EXISTS insert_app_vault_balance_operation_last;
DROP TABLE IF EXISTS app_vault_balance_operation_last;
-- +goose StatementEnd
