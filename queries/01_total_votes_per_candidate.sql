-- Total suara per kandidat di satu election
-- Parameter: $1 = election_id

SELECT
    c.id              AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name   AS chairman_name,
    c.vice_chairman_name AS vice_chairman_name,
    COUNT(v.id)       AS total_votes
FROM candidates c
LEFT JOIN votes v
    ON v.candidate_id = c.id
   AND v.election_id = $1
WHERE c.election_id = $1
GROUP BY c.id, c.candidate_number, c.chairman_name, c.vice_chairman_name
ORDER BY c.candidate_number;
