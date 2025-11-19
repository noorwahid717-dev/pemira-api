-- Analisis jam tersibuk voting (peak hours)
-- Parameter: $1 = election_id
-- Output: ranking hours untuk capacity planning

WITH hourly_votes AS (
    SELECT
        date_trunc('hour', v.cast_at) AS vote_hour,
        COUNT(*) AS total_votes,
        COUNT(*) FILTER (WHERE v.channel = 'ONLINE') AS votes_online,
        COUNT(*) FILTER (WHERE v.channel = 'TPS') AS votes_tps
    FROM votes v
    WHERE v.election_id = $1
    GROUP BY date_trunc('hour', v.cast_at)
),
ranked AS (
    SELECT
        vote_hour,
        total_votes,
        votes_online,
        votes_tps,
        RANK() OVER (ORDER BY total_votes DESC) AS rank_by_total
    FROM hourly_votes
)
SELECT
    vote_hour,
    EXTRACT(HOUR FROM vote_hour) AS hour_of_day,
    TO_CHAR(vote_hour, 'Day') AS day_name,
    total_votes,
    votes_online,
    votes_tps,
    rank_by_total
FROM ranked
ORDER BY total_votes DESC
LIMIT 20;
