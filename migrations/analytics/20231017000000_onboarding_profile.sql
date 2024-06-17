-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS onboarding_profile_view  AS
    SELECT
    ((event->>'response_body')::jsonb->'result'->0->'profile'->'id')::bigint AS "profile_id",
    (event->>'response_body')::jsonb->'result'->0->'profile'->>'wallet' AS "wallet",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_source' AS "utm_source",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_medium' AS "utm_medium",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_campaign' AS "utm_campaign",
    ts AS "timestamp"

    FROM analytic_event
    WHERE
        event->>'url_path' = '/onboarding' AND
        event->>'http_method' = 'POST' AND
        (event->>'request_body')::jsonb->'campaign'->-1->'utm_source' IS NOT NULL;


CREATE OR REPLACE FUNCTION refresh_onboarding_profile_view(job_id int, config jsonb)
RETURNS VOID AS
$$
DECLARE
    current_ts TIMESTAMP;
BEGIN
    SELECT CURRENT_TIMESTAMP INTO current_ts;
    INSERT INTO onboarding_profile_view SELECT
    ((event->>'response_body')::jsonb->'result'->0->'profile'->'id')::bigint AS "profile_id",
    (event->>'response_body')::jsonb->'result'->0->'profile'->>'wallet' AS "wallet",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_source' AS "utm_source",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_medium' AS "utm_medium",
    (event->>'request_body')::jsonb->'campaign'->-1->>'utm_campaign' AS "utm_campaign",
    ts AS "timestamp"

    FROM analytic_event
    WHERE
        (ts > (config->>'last_ts')::timestamp AND ts <= current_ts) AND
        event->>'url_path' = '/onboarding' AND
        LOWER(event->>'http_method') = 'post' AND
        (event->>'request_body')::jsonb->'campaign'->-1->'utm_source' IS NOT NULL;

    PERFORM alter_job(job_id, config => json_build_object('last_ts', current_ts)::jsonb);
END;
$$ LANGUAGE plpgsql;

SELECT add_job(
    'refresh_onboarding_profile_view',
    '24 hours',
    config => json_build_object(
        'last_ts', (SELECT COALESCE(MAX(timestamp), TO_TIMESTAMP(0)) FROM onboarding_profile_view)
    )::jsonb
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS onboarding_profile_view;

DROP FUNCTION IF EXISTS refresh_onboarding_profile_view;

SELECT delete_job(
    (SELECT job_id
     FROM timescaledb_information.jobs
     WHERE proc_name = 'refresh_onboarding_profile_view'
     LIMIT 1)
);
-- +goose StatementEnd
