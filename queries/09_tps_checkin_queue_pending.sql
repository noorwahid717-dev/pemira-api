-- Antrian PENDING untuk satu TPS, urut dari paling awal
-- Parameter: $1 = tps_id (bukan election_id)

SELECT
    tc.id                   AS checkin_id,
    tc.scan_at,
    v.id                    AS voter_id,
    v.nim,
    v.name,
    v.faculty_name,
    v.study_program_name,
    v.cohort_year
FROM tps_checkins tc
JOIN voters v
  ON v.id = tc.voter_id
WHERE tc.tps_id = $1
  AND tc.status = 'PENDING'
ORDER BY tc.scan_at ASC
LIMIT 100;
