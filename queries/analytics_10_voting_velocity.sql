-- Analisis kecepatan voting (velocity): berapa menit rata-rata antar suara
-- Parameter: $1 = election_id
-- Output: velocity metrics untuk monitoring

WITH vote_times AS (
    SELECT
        cast_at,
        LAG(cast_at) OVER (ORDER BY cast_at) AS prev_cast_at
    FROM votes
    WHERE election_id = $1
),
gaps AS (
    SELECT
        EXTRACT(EPOCH FROM (cast_at - prev_cast_at)) / 60 AS gap_minutes
    FROM vote_times
    WHERE prev_cast_at IS NOT NULL
)
SELECT
    COUNT(*) AS total_intervals,
    ROUND(AVG(gap_minutes)::NUMERIC, 2) AS avg_gap_minutes,
    ROUND(MIN(gap_minutes)::NUMERIC, 2) AS min_gap_minutes,
    ROUND(MAX(gap_minutes)::NUMERIC, 2) AS max_gap_minutes,
    ROUND(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY gap_minutes)::NUMERIC, 2) AS median_gap_minutes,
    ROUND(PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY gap_minutes)::NUMERIC, 2) AS p95_gap_minutes
FROM gaps;
