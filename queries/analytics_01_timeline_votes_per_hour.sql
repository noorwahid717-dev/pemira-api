-- Timeline: jumlah suara per jam (semua channel)
-- Parameter: $1 = election_id
-- Output: time-series data untuk grafik line/bar

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
        COUNT(*) AS total_votes
    FROM votes v
    WHERE v.election_id = $1
    GROUP BY date_trunc('hour', v.cast_at)
)
SELECT
    b.bucket_start,
    COALESCE(vpb.total_votes, 0) AS total_votes
FROM buckets b
LEFT JOIN votes_per_bucket vpb
    ON vpb.bucket_start = b.bucket_start
ORDER BY b.bucket_start;
