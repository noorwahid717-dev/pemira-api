-- Turnout per fakultas
-- Parameter: $1 = election_id

WITH base AS (
    SELECT
        vs.election_id,
        v.faculty_code,
        v.faculty_name,
        vs.has_voted
    FROM voter_status vs
    JOIN voters v
      ON v.id = vs.voter_id
    WHERE vs.election_id = $1
      AND vs.is_eligible = TRUE
)
SELECT
    faculty_code,
    faculty_name,
    COUNT(*)                                 AS total_eligible,
    SUM(CASE WHEN has_voted THEN 1 ELSE 0 END) AS total_voted,
    ROUND(
        SUM(CASE WHEN has_voted THEN 1 ELSE 0 END)::NUMERIC
        / NULLIF(COUNT(*), 0) * 100,
        2
    ) AS turnout_percent
FROM base
GROUP BY faculty_code, faculty_name
ORDER BY faculty_name;
