-- Partisipasi per TPS dengan usage percentage
-- Parameter: $1 = election_id

SELECT
    t.id              AS tps_id,
    t.code            AS tps_code,
    t.name            AS tps_name,
    t.capacity_estimate,
    COUNT(vs.id)      AS total_voted_tps,
    ROUND(
        COUNT(vs.id)::NUMERIC / NULLIF(t.capacity_estimate, 0) * 100,
        2
    ) AS usage_percent
FROM tps t
LEFT JOIN voter_status vs
    ON vs.tps_id = t.id
   AND vs.election_id = t.election_id
   AND vs.has_voted = TRUE
   AND vs.voting_method = 'TPS'
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name, t.capacity_estimate
ORDER BY t.code;
