-- +goose Up
-- +goose StatementBegin
CREATE OR REPLACE FUNCTION interval_to_micros(i TEXT) returns BIGINT LANGUAGE SQL STABLE AS
$$
    SELECT EXTRACT(EPOCH FROM i::interval)::bigint * 1000000
$$;


CREATE TABLE IF NOT EXISTS referral_leaderboard_weekly_rank
(
    profile_id      BIGINT PRIMARY KEY NOT NULL,
    previous_rank   BIGINT             NULL DEFAULT 0,
    previous_volume NUMERIC            NULL DEFAULT 0,
    current_volume  NUMERIC            NOT NULL,
    current_rank    BIGINT             NOT NULL
);

CREATE TABLE IF NOT EXISTS referral_leaderboard_monthly_rank
(
    profile_id      BIGINT PRIMARY KEY NOT NULL,
    previous_rank   BIGINT             NULL DEFAULT 0,
    previous_volume NUMERIC            NULL DEFAULT 0,
    current_volume  NUMERIC            NOT NULL,
    current_rank    BIGINT             NOT NULL
);

CREATE TABLE IF NOT EXISTS referral_leaderboard_lifetime_rank
(
    profile_id      BIGINT PRIMARY KEY NOT NULL,
    previous_rank   BIGINT             NULL DEFAULT 0,
    previous_volume NUMERIC            NULL DEFAULT 0,
    current_volume  NUMERIC            NOT NULL,
    current_rank    BIGINT             NOT NULL
);


CREATE OR REPLACE FUNCTION referral_refresh_leaderboard_rank(period TEXT)
    RETURNS VOID AS
$$
DECLARE
    rankQuery TEXT;
    insertQuery TEXT;
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
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP FUNCTION IF EXISTS interval_to_micros(TEXT);
DROP FUNCTION IF EXISTS referral_refresh_leaderboard_rank(TEXT);

DROP TABLE IF EXISTS referral_leaderboard_weekly_rank CASCADE;
DROP TABLE IF EXISTS referral_leaderboard_monthly_rank CASCADE;
DROP TABLE IF EXISTS referral_leaderboard_lifetime_rank CASCADE;
-- +goose StatementEnd
