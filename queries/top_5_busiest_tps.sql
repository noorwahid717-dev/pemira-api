-- Top 5 TPS dengan jumlah suara terbanyak
-- Parameter: $1 = election_id

SELECT
    t.code,
    t.name,
    COUNT(v.id) AS total_votes
FROM tps t
LEFT JOIN votes v
    ON v.tps_id = t.id
   AND v.election_id = $1
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name
ORDER BY total_votes DESC
LIMIT 5;
