-- List pemilih yang belum vote untuk reminder campaign
-- Parameter: $1 = election_id

SELECT
    v.id,
    v.nim,
    v.name,
    v.email,
    v.faculty_name,
    v.study_program_name
FROM voter_status vs
JOIN voters v ON v.id = vs.voter_id
WHERE vs.election_id = $1
  AND vs.is_eligible = TRUE
  AND vs.has_voted = FALSE
ORDER BY v.faculty_name, v.name
LIMIT 100;
