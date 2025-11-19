-- Timeline: suara per jam per kandidat
-- Parameter: $1 = election_id
-- Output: multi-line chart (satu series per kandidat)

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
        c.id              AS candidate_id,
        c.candidate_number AS candidate_number,
        c.chairman_name   AS candidate_name,
        COUNT(*) AS total_votes
    FROM votes v
    JOIN candidates c
      ON c.id = v.candidate_id
     AND c.election_id = v.election_id
    WHERE v.election_id = $1
    GROUP BY date_trunc('hour', v.cast_at), c.id, c.candidate_number, c.chairman_name
)
SELECT
    b.bucket_start,
    c.id              AS candidate_id,
    c.candidate_number,
    c.chairman_name   AS candidate_name,
    COALESCE(vpb.total_votes, 0) AS total_votes
FROM buckets b
CROSS JOIN (
    SELECT id, candidate_number, chairman_name 
    FROM candidates 
    WHERE election_id = $1
) c
LEFT JOIN votes_per_bucket vpb
       ON vpb.bucket_start = b.bucket_start
      AND vpb.candidate_id = c.id
ORDER BY b.bucket_start, c.candidate_number;
