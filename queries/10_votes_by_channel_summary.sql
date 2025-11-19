-- Berapa suara via ONLINE vs TPS untuk satu election
-- Parameter: $1 = election_id

SELECT
    SUM(CASE WHEN channel = 'ONLINE' THEN 1 ELSE 0 END) AS total_online,
    SUM(CASE WHEN channel = 'TPS'    THEN 1 ELSE 0 END) AS total_tps,
    COUNT(*)                                            AS total_votes
FROM votes
WHERE election_id = $1;
