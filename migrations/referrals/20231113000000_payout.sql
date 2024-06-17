-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS referral_payout
(
    id         TEXT PRIMARY KEY NOT NULL,
    profile_id BIGINT           NOT NULL,
    market_id  TEXT             NOT NULL,
    amount     NUMERIC          NOT NULL,
    processed  BOOLEAN          NOT NULL DEFAULT FALSE,
    timestamp  BIGINT           NOT NULL DEFAULT unix_now()
);

CREATE TABLE IF NOT EXISTS referral_payout_bonus_integrity
(
    profile_id BIGINT NOT NULL,
    level      BIGINT NOT NULL,
    timestamp  BIGINT NOT NULL DEFAULT unix_now(),

    UNIQUE (profile_id, level)
);


CREATE TABLE IF NOT EXISTS referral_fills_integrity
(
    shard_id         TEXT   NOT NULL,
    archive_id_start BIGINT NOT NULL,
    archive_id_end   BIGINT NOT NULL,
    EXCLUDE USING gist (
        "shard_id" WITH =, int8range("archive_id_start", "archive_id_end", '[]') WITH &&
        )
);

CREATE OR REPLACE FUNCTION referral_get_fills(shard_id TEXT)
    RETURNS TABLE
            (
                referrer_id       BIGINT,
                invited_id        BIGINT,
                profile_id        BIGINT,
                trade_id          TEXT,
                fee               NUMERIC,
                is_maker          BOOLEAN,
                model             TEXT,
                model_fee_percent NUMERIC,
                profile_volume    NUMERIC,
                archive_id        BIGINT
            )
AS
$$
DECLARE
    from_archive_id BIGINT;
    last_trade_id TEXT;
BEGIN
    SELECT COALESCE(MAX(rvi.archive_id_end), 0)
    FROM referral_fills_integrity rvi
    WHERE rvi.shard_id = $1
    INTO from_archive_id;

    SELECT COALESCE(MAX(f.trade_id), '')
    FROM app_fill f
    WHERE f.shard_id = $1 AND f.archive_id = from_archive_id
    INTO last_trade_id;

    RETURN QUERY
        WITH trade_ids AS (SELECT f.trade_id
                           FROM app_fill f
                                    INNER JOIN app_referral_link l ON l.invited_id = f.profile_id
                           WHERE (f.shard_id = $1 AND f.archive_id > from_archive_id)
                             AND f.order_id != 'wf3'
                             AND f.trade_id != last_trade_id
                           GROUP BY 1)

        SELECT l.profile_id           "referrer_id",
               l.invited_id           "invited_id",
               f.profile_id,
               f.trade_id,
               f.fee,
               f.is_maker,
               arc.model              "model",
               arc.model_fee_percent  "model_fee_percent",
               COALESCE(fv.volume, 0) "profile_volume",
               f.archive_id           "archive_id"
        FROM trade_ids tid
                 LEFT JOIN app_fill f ON tid.trade_id = f.trade_id
                 LEFT JOIN app_referral_link l ON f.profile_id = l.invited_id
                 LEFT JOIN referral_volumes fv ON fv.profile_id = l.profile_id
                 LEFT JOIN app_referral_code arc ON arc.profile_id = l.profile_id
        ORDER BY f.timestamp, f.trade_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS referral_payout;
DROP FUNCTION IF EXISTS referral_get_fills;
-- +goose StatementEnd
