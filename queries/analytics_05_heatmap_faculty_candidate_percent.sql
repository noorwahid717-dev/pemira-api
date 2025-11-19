-- Heatmap: fakultas × kandidat dengan persentase di dalam fakultas
-- Parameter: $1 = election_id
-- Output: matrix dengan preferensi relatif per fakultas

WITH faculty_totals AS (
    SELECT
        v.faculty_code,
        v.faculty_name,
        COUNT(*) AS total_votes_faculty
    FROM votes vts
    JOIN voter_status vs
      ON vs.election_id = vts.election_id
     AND vs.vote_token_hash = vts.token_hash
    JOIN voters v
      ON v.id = vs.voter_id
    WHERE vts.election_id = $1
    GROUP BY v.faculty_code, v.faculty_name
),
faculty_candidate AS (
    SELECT
        v.faculty_code,
        v.faculty_name,
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
    GROUP BY v.faculty_code, v.faculty_name, c.id, c.candidate_number, c.chairman_name
)
SELECT
    fc.faculty_code,
    fc.faculty_name,
    fc.candidate_id,
    fc.candidate_number,
    fc.candidate_name,
    fc.total_votes,
    ROUND(
        fc.total_votes::NUMERIC / NULLIF(ft.total_votes_faculty, 0) * 100,
        2
    ) AS percent_in_faculty
FROM faculty_candidate fc
JOIN faculty_totals ft
  ON ft.faculty_code = fc.faculty_code
ORDER BY fc.faculty_name, fc.candidate_number;
