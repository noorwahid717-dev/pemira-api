-- Hasil per TPS: berapa suara tiap kandidat di TPS tersebut
-- Parameter: $1 = election_id

SELECT
    t.id               AS tps_id,
    t.code             AS tps_code,
    t.name             AS tps_name,
    c.id               AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name    AS chairman_name,
    COUNT(v.id)        AS total_votes
FROM tps t
JOIN votes v
    ON v.tps_id = t.id
   AND v.election_id = $1
JOIN candidates c
    ON c.id = v.candidate_id
   AND c.election_id = $1
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name, c.id, c.candidate_number, c.chairman_name
ORDER BY t.code, c.candidate_number;
