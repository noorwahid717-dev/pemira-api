-- Turnout total satu pemilu
-- Parameter: $1 = election_id

SELECT
    COUNT(*) FILTER (WHERE is_eligible)                      AS total_eligible,
    COUNT(*) FILTER (WHERE is_eligible AND has_voted)        AS total_voted,
    ROUND(
        COUNT(*) FILTER (WHERE is_eligible AND has_voted)::NUMERIC
        / NULLIF(COUNT(*) FILTER (WHERE is_eligible), 0) * 100,
        2
    ) AS turnout_percent
FROM voter_status
WHERE election_id = $1;
