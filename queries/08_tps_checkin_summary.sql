-- Ringkasan check-in per TPS untuk dashboard TPS / admin
-- Parameter: $1 = election_id

SELECT
    t.id   AS tps_id,
    t.code AS tps_code,
    t.name AS tps_name,

    COUNT(tc.id)                                         AS total_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'PENDING')    AS pending_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'APPROVED')   AS approved_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'REJECTED')   AS rejected_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'USED')       AS used_checkins,
    COUNT(tc.id) FILTER (WHERE tc.status = 'EXPIRED')    AS expired_checkins

FROM tps t
LEFT JOIN tps_checkins tc
       ON tc.tps_id = t.id
      AND tc.election_id = t.election_id
WHERE t.election_id = $1
GROUP BY t.id, t.code, t.name
ORDER BY t.code;
