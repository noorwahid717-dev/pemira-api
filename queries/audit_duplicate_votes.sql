-- Audit: Cek apakah ada token_hash yang dipakai lebih dari 1x
-- Parameter: $1 = election_id

SELECT
    token_hash,
    COUNT(*) AS usage_count
FROM votes
WHERE election_id = $1
GROUP BY token_hash
HAVING COUNT(*) > 1;
