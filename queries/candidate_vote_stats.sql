-- Candidate vote statistics with percentage
-- Parameter: $1 = election_id
-- Output: candidate_id, candidate_votes, percentage

WITH total_election AS (
    SELECT
        COUNT(*)::NUMERIC AS total_votes
    FROM votes
    WHERE election_id = $1
)
SELECT
    v.candidate_id,
    COUNT(*) AS candidate_votes,
    CASE
        WHEN te.total_votes = 0 THEN 0
        ELSE ROUND(COUNT(*)::NUMERIC / te.total_votes * 100, 2)
    END AS percentage
FROM votes v
CROSS JOIN total_election te
WHERE v.election_id = $1
GROUP BY v.candidate_id, te.total_votes
ORDER BY v.candidate_id;
