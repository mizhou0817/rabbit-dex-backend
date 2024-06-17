-- +goose Up
-- +goose StatementBegin

-- weekly
ALTER TABLE referral_leaderboard_weekly_rank
    ADD COLUMN exchange_id TEXT NOT NULL DEFAULT 'rbx';

CREATE INDEX IF NOT EXISTS referral_leaderboard_weekly_rank_exchange_id_idx ON referral_leaderboard_weekly_rank (exchange_id);

WITH subquery AS (SELECT id, exchange_id
                  from app_profile
                  WHERE exchange_id in ('rbx', 'bfx'))
UPDATE referral_leaderboard_weekly_rank
SET exchange_id = subquery.exchange_id
FROM subquery
WHERE profile_id = subquery.id;

-- monthly
ALTER TABLE referral_leaderboard_monthly_rank
    ADD COLUMN exchange_id TEXT NOT NULL DEFAULT 'rbx';

CREATE INDEX IF NOT EXISTS referral_leaderboard_monthly_rank_exchange_id_idx ON referral_leaderboard_monthly_rank (exchange_id);

WITH subquery AS (SELECT id, exchange_id
                  from app_profile
                  WHERE exchange_id in ('rbx', 'bfx'))
UPDATE referral_leaderboard_monthly_rank
SET exchange_id = subquery.exchange_id
FROM subquery
WHERE profile_id = subquery.id;

-- lifetime
ALTER TABLE referral_leaderboard_lifetime_rank
    ADD COLUMN exchange_id TEXT NOT NULL DEFAULT 'rbx';

CREATE INDEX IF NOT EXISTS referral_leaderboard_lifetime_rank_exchange_id_idx ON referral_leaderboard_lifetime_rank (exchange_id);

WITH subquery AS (SELECT id, exchange_id
                  from app_profile
                  WHERE exchange_id in ('rbx', 'bfx'))
UPDATE referral_leaderboard_lifetime_rank
SET exchange_id = subquery.exchange_id
FROM subquery
WHERE profile_id = subquery.id;

-- function
CREATE OR REPLACE FUNCTION referral_refresh_leaderboard_rank(exchange_id TEXT, period TEXT)
    RETURNS VOID AS
$$
DECLARE
    rankQuery        TEXT;
    insertQuery      TEXT;
    destinationTable TEXT;
BEGIN
    IF period NOT IN ('lifetime', 'weekly', 'monthly') THEN
        RAISE EXCEPTION 'incorrect period %, expected lifetime|weekly|monthly', period;
    END IF;

    if period IN ('weekly', 'monthly') THEN
        rankQuery = $text$
            WITH ranks AS (
                SELECT  l.profile_id                       AS "profile_id",
                        p.exchange_id AS "exchange_id",
                        COALESCE(SUM(f.volume), 0) AS "volume",
                        DENSE_RANK() OVER (ORDER BY COALESCE(SUM(f.volume), 0) DESC) "rank"
                FROM app_referral_link as l
                JOIN app_profile as p on p.id = l.profile_id
                LEFT JOIN LATERAL (
                    SELECT
                        f.profile_id "profile_id",
                        (f.size * f.price) "volume"
                        FROM app_fill f
                        WHERE f.profile_id = l.invited_id
                            AND f.timestamp >= unix_now() - interval_to_micros('%s')
                            AND f.order_id != 'wf3'
                ) f ON true
                WHERE p.exchange_id = '%s'
                GROUP by 1,2
                ORDER BY 3 DESC
            )
        $text$;
    ELSE
        rankQuery = $text$
            WITH ranks AS (
                SELECT v.profile_id "profile_id",
                       p.exchange_id AS "exchange_id",
                       v.volume "volume",
                       DENSE_RANK() OVER (ORDER BY v.volume DESC) "rank"
                FROM referral_volumes as v
                JOIN app_profile as p on p.id = v.profile_id
                WHERE p.exchange_id = '%s'
                ORDER BY v.volume DESC
            )
        $text$;
    END IF;


    IF period = 'weekly' THEN
        destinationTable = 'referral_leaderboard_weekly_rank';
        rankQuery = format(rankQuery, '7 days', exchange_id);
    ELSEIF period = 'monthly' THEN
        destinationTable = 'referral_leaderboard_monthly_rank';
        rankQuery = format(rankQuery, '30 days', exchange_id);
    ELSE
        destinationTable = 'referral_leaderboard_lifetime_rank';
        rankQuery = format(rankQuery, exchange_id);
    END IF;

    insertQuery = format($text$
    INSERT
    INTO %s AS lr(profile_id, exchange_id, current_volume, current_rank)
            (SELECT profile_id, exchange_id, volume, rank FROM ranks)
    ON CONFLICT (profile_id)
        DO UPDATE
        SET previous_volume = lr.current_volume,
            previous_rank   = lr.current_rank,

            exchange_id     = EXCLUDED.exchange_id,
            current_volume  = EXCLUDED.current_volume,
            current_rank    = EXCLUDED.current_rank;
    $text$, destinationTable);

    EXECUTE format('%s %s', rankQuery, insertQuery);
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION referral_refresh_leaderboard_rank(period TEXT)
    RETURNS VOID AS
$$
DECLARE
    rankQuery        TEXT;
    insertQuery      TEXT;
    destinationTable TEXT;
BEGIN
    IF period NOT IN ('lifetime', 'weekly', 'monthly') THEN
        RAISE EXCEPTION 'incorrect period %, expected lifetime|weekly|monthly', period;
    END IF;

    if period IN ('weekly', 'monthly') THEN
        rankQuery = $text$
            WITH ranks AS (
                SELECT  l.profile_id                       AS "profile_id",
                        COALESCE(SUM(f.volume), 0) AS "volume",
                        DENSE_RANK() OVER (ORDER BY COALESCE(SUM(f.volume), 0) DESC) "rank"
                FROM app_referral_link l
                LEFT JOIN LATERAL (
                    SELECT
                        f.profile_id "profile_id",
                        (f.size * f.price) "volume"
                        FROM app_fill f
                        WHERE f.profile_id = l.invited_id
                            AND f.timestamp >= unix_now() - interval_to_micros('%s')
                            AND f.order_id != 'wf3'
                ) f ON true
                GROUP by 1
                ORDER BY 2 DESC
            )
        $text$;
    ELSE
        rankQuery = $text$
            WITH ranks AS (
                SELECT profile_id "profile_id",
                       volume "volume",
                       DENSE_RANK() OVER (ORDER BY volume DESC) "rank"
                FROM referral_volumes
                ORDER BY volume DESC
            )
        $text$;
    END IF;


    IF period = 'weekly' THEN
        destinationTable = 'referral_leaderboard_weekly_rank';
        rankQuery = format(rankQuery, '7 days');
    ELSEIF period = 'monthly' THEN
        destinationTable = 'referral_leaderboard_monthly_rank';
        rankQuery = format(rankQuery, '30 days');
    ELSE
        destinationTable = 'referral_leaderboard_lifetime_rank';
    END IF;

    insertQuery = format($text$
    INSERT
    INTO %s AS lr(profile_id, current_volume, current_rank)
            (SELECT profile_id, volume, rank FROM ranks)
    ON CONFLICT (profile_id)
        DO UPDATE
        SET previous_volume = lr.current_volume,
            previous_rank   = lr.current_rank,

            current_volume  = EXCLUDED.current_volume,
            current_rank    = EXCLUDED.current_rank;
    $text$, destinationTable);

    EXECUTE format('%s %s', rankQuery, insertQuery);
END;
$$ LANGUAGE plpgsql;


DROP INDEX IF EXISTS referral_leaderboard_lifetime_rank_exchange_id_idx;
ALTER TABLE referral_leaderboard_lifetime_rank
    DROP COLUMN exchange_id;

DROP INDEX IF EXISTS referral_leaderboard_monthly_rank_exchange_id_idx;
ALTER TABLE referral_leaderboard_monthly_rank
    DROP COLUMN exchange_id;

DROP INDEX IF EXISTS referral_leaderboard_weekly_rank_exchange_id_idx;
ALTER TABLE referral_leaderboard_weekly_rank
    DROP COLUMN exchange_id;
-- +goose StatementEnd
