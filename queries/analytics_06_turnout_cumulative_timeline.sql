-- Timeline partisipasi kumulatif dengan persentase turnout
-- Parameter: $1 = election_id
-- Output: cumulative turnout over time untuk grafik line

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
        COUNT(*) AS votes_in_hour
    FROM votes v
    WHERE v.election_id = $1
    GROUP BY date_trunc('hour', v.cast_at)
),
total_eligible AS (
    SELECT COUNT(*) AS total
    FROM voter_status
    WHERE election_id = $1
      AND is_eligible = TRUE
),
bucket_with_cum AS (
    SELECT
        b.bucket_start,
        COALESCE(vpb.votes_in_hour, 0) AS votes_in_hour,
        SUM(COALESCE(vpb.votes_in_hour, 0)) OVER (
            ORDER BY b.bucket_start
            ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
        ) AS cumulative_votes
    FROM buckets b
    LEFT JOIN votes_per_bucket vpb
           ON vpb.bucket_start = b.bucket_start
)
SELECT
    bwc.bucket_start,
    bwc.votes_in_hour,
    bwc.cumulative_votes,
    ROUND(
        bwc.cumulative_votes::NUMERIC / NULLIF(te.total, 0) * 100,
        2
    ) AS cumulative_turnout_percent
FROM bucket_with_cum bwc
CROSS JOIN total_eligible te
ORDER BY bwc.bucket_start;
