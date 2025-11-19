-- Heatmap: fakultas × kandidat (jumlah suara)
-- Parameter: $1 = election_id
-- Output: matrix data untuk heatmap visualization

SELECT
    v.faculty_code,
    v.faculty_name,
    c.id               AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name    AS candidate_name,
    COUNT(*)           AS total_votes
FROM votes vts
JOIN voter_status vs
  ON vs.election_id = vts.election_id
 AND vs.vote_token_hash = vts.token_hash
JOIN voters v
  ON v.id = vs.voter_id
JOIN candidates c
  ON c.id = vts.candidate_id
 AND c.election_id = vts.election_id
WHERE vts.election_id = $1
GROUP BY v.faculty_code, v.faculty_name, c.id, c.candidate_number, c.chairman_name
ORDER BY v.faculty_name, c.candidate_number;
