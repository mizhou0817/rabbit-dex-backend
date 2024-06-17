-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS ignored_profile_ids
(
    profile_id BIGINT PRIMARY KEY
);

INSERT INTO ignored_profile_ids(profile_id)
VALUES (0),
       (19),
       (20),
       (14910),
       (14919);

-- 01-kpis-by-campaign.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (13),
       (20)
ON CONFLICT (profile_id) DO NOTHING;

-- 02-all-trader-wallets-with-kpis.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (0),
       (13),
       (14),
       (15),
       (19),
       (20),
       (704),
       (1561),
       (2104),
       (11002),
       (11007),
       (11831),
       (12071),
       (12072),
       (12472),
       (12483),
       (12505),
       (13436),
       (14910),
       (14919),
       (15679)
ON CONFLICT (profile_id) DO NOTHING;

-- 1. market-info.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (13),
       (14),
       (15),
       (19),
       (20),
       (1561),
       (2104),
       (11002),
       (704),
       (11007),
       (11831),
       (12071),
       (12072),
       (12483),
       (13436),
       (12472),
       (12505),
       (14910),
       (14919),
       (15679),
       (16510),
       (16955),
       (17217)
ON CONFLICT (profile_id) DO NOTHING;

-- 2. market-info.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (13),
       (20),
       (2104),
       (12071),
       (12072),
       (12472),
       (12483),
       (13436),
       (14910),
       (14919),
       (15679),
       (16510),
       (16955),
       (17217)
ON CONFLICT (profile_id) DO NOTHING;

-- 3. market-info.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (13),
       (14),
       (15),
       (17),
       (19),
       (20),
       (1561),
       (2104),
       (11002),
       (704),
       (11007),
       (11831),
       (12071),
       (12072),
       (12483),
       (13436),
       (12505)
ON CONFLICT (profile_id) DO NOTHING;

-- playground.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (0),
       (13),
       (14),
       (15),
       (17),
       (19),
       (20),
       (1561),
       (2104),
       (11002),
       (704),
       (11007),
       (11831),
       (12071),
       (12072),
       (12483),
       (13436),
       (12472)
ON CONFLICT (profile_id) DO NOTHING;

-- user-research.json
INSERT INTO ignored_profile_ids(profile_id)
VALUES (0),
       (13),
       (14),
       (15),
       (19),
       (20),
       (14910),
       (14919),
       (1561),
       (11002),
       (11007),
       (704),
       (11831),
       (2104),
       (12071),
       (12072),
       (16510),
       (16955),
       (12483),
       (13436),
       (12472),
       (12505),
       (15679),
       (17217)
ON CONFLICT (profile_id) DO NOTHING;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS ignored_profile_ids;
-- +goose StatementEnd
