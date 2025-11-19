-- Dashboard Admin Utama - All-in-One Summary
-- Parameter: $1 = election_id

WITH
vs_agg AS (
    SELECT
        election_id,
        COUNT(*) FILTER (WHERE is_eligible)                    AS total_eligible,
        COUNT(*) FILTER (WHERE is_eligible AND has_voted)      AS total_voted,
        COUNT(*) FILTER (WHERE is_eligible AND NOT has_voted)  AS total_not_voted
    FROM voter_status
    WHERE election_id = $1
    GROUP BY election_id
),
votes_agg AS (
    SELECT
        election_id,
        COUNT(*)                                              AS total_votes,
        COUNT(*) FILTER (WHERE channel = 'ONLINE')             AS total_online,
        COUNT(*) FILTER (WHERE channel = 'TPS')                AS total_tps
    FROM votes
    WHERE election_id = $1
    GROUP BY election_id
),
tps_agg AS (
    SELECT
        election_id,
        COUNT(*)                                             AS total_tps,
        COUNT(*) FILTER (WHERE status = 'ACTIVE')            AS active_tps
    FROM tps
    WHERE election_id = $1
    GROUP BY election_id
)
SELECT
    e.id                  AS election_id,
    e.code                AS election_code,
    e.name                AS election_name,
    e.status              AS election_status,
    e.voting_start_at,
    e.voting_end_at,

    vs.total_eligible,
    vs.total_voted,
    vs.total_not_voted,
    ROUND(
        vs.total_voted::NUMERIC
        / NULLIF(vs.total_eligible, 0) * 100,
        2
    ) AS turnout_percent,

    vagg.total_votes,
    vagg.total_online,
    vagg.total_tps,

    tagg.total_tps,
    tagg.active_tps

FROM elections e
LEFT JOIN vs_agg   vs   ON vs.election_id = e.id
LEFT JOIN votes_agg vagg ON vagg.election_id = e.id
LEFT JOIN tps_agg  tagg ON tagg.election_id = e.id
WHERE e.id = $1;
