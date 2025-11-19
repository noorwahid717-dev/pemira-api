-- Timeline: suara per jam dipisah channel (ONLINE vs TPS)
-- Parameter: $1 = election_id
-- Output: stacked/clustered bar chart data

WITH params AS (
    SELECT
        e.id AS election_id,
        e.voting_start_at AS start_ts,
        e.voting_end_at   AS end_ts
    FROM elections e
    WHERE e.id = $1
),
buckets AS (
    SELECT
        generate_series(
            date_trunc('hour', (SELECT start_ts FROM params)),
            date_trunc('hour', (SELECT end_ts   FROM params)),
            interval '1 hour'
        ) AS bucket_start
),
votes_per_bucket AS (
    SELECT
        date_trunc('hour', v.cast_at) AS bucket_start,
        v.channel,
        COUNT(*) AS cnt
    FROM votes v
    WHERE v.election_id = $1
    GROUP BY date_trunc('hour', v.cast_at), v.channel
)
SELECT
    b.bucket_start,
    COALESCE(SUM(CASE WHEN vpb.channel = 'ONLINE' THEN vpb.cnt END), 0) AS votes_online,
    COALESCE(SUM(CASE WHEN vpb.channel = 'TPS'    THEN vpb.cnt END), 0) AS votes_tps,
    COALESCE(SUM(vpb.cnt), 0) AS total_votes
FROM buckets b
LEFT JOIN votes_per_bucket vpb
       ON vpb.bucket_start = b.bucket_start
GROUP BY b.bucket_start
ORDER BY b.bucket_start;
