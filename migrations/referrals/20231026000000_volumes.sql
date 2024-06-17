-- +goose Up
-- +goose StatementBegin
CREATE INDEX IF NOT EXISTS app_fill_trade_id_idx
    ON app_fill (trade_id);

CREATE INDEX IF NOT EXISTS app_fill_shard_id_archive_id_idx
    ON app_fill (shard_id, archive_id);

CREATE TABLE IF NOT EXISTS referral_volumes
(
    profile_id BIGINT PRIMARY KEY,
    volume     NUMERIC NOT NULL,
    created_at BIGINT  NOT NULL DEFAULT unix_now(),
    updated_at BIGINT  NULL
);
CREATE INDEX IF NOT EXISTS referral_volumes_volume_idx ON referral_volumes (volume);

CREATE EXTENSION IF NOT EXISTS btree_gist;
CREATE TABLE IF NOT EXISTS referral_volumes_integrity
(
    shard_id         TEXT   NOT NULL,
    archive_id_start BIGINT NOT NULL,
    archive_id_end   BIGINT NOT NULL,
    EXCLUDE USING gist (
        "shard_id" WITH =, int8range("archive_id_start", "archive_id_end", '[]') WITH &&
        )
);

CREATE OR REPLACE FUNCTION referral_get_volumes(_shard_id TEXT)
    RETURNS TABLE
            (
                profile_id      BIGINT,
                volume          NUMERIC,
                existing_volume NUMERIC,
                archive_id      BIGINT
            )
AS
$$
DECLARE
    from_archive_id BIGINT;
BEGIN
    SELECT COALESCE(MAX(rvi.archive_id_end), 0)
    FROM referral_volumes_integrity rvi
    WHERE rvi.shard_id = $1
    INTO from_archive_id;

    RETURN QUERY
        SELECT l.profile_id                  AS "profile_id",
               COALESCE(f.size * f.price, 0) AS "volume",
               COALESCE(rv.volume, 0)        AS "existing_volume",
               f.archive_id                     "archive_id"
        FROM app_referral_link l
        LEFT JOIN referral_volumes rv ON rv.profile_id = l.profile_id
        INNER JOIN LATERAL (
          SELECT f.*
          FROM app_fill f
          WHERE f.profile_id = l.invited_id
                AND f.shard_id = $1 AND f.archive_id > from_archive_id
                AND f.order_id != 'wf3'
        ) f ON true
        ORDER BY f.archive_id;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS app_fill_trade_id_idx;
DROP INDEX IF EXISTS app_fill_shard_id_archive_id_idx;

DROP TABLE IF EXISTS referral_volumes CASCADE;
DROP TABLE IF EXISTS referral_volumes_integrity CASCADE;
DROP FUNCTION IF EXISTS referral_get_volumes;
-- +goose StatementEnd
