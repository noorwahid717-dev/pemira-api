-- Partisipasi per angkatan (cohort) per kandidat
-- Parameter: $1 = election_id
-- Output: data untuk clustered bar chart atau grouped analysis

SELECT
    v.cohort_year,
    c.id               AS candidate_id,
    c.candidate_number AS candidate_number,
    c.chairman_name    AS candidate_name,
    COUNT(*) AS total_votes
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
GROUP BY v.cohort_year, c.id, c.candidate_number, c.chairman_name
ORDER BY v.cohort_year, c.candidate_number;
