-- Update election schedule dengan semua tahapan
UPDATE elections 
SET 
    registration_start_at = '2025-11-01 00:00:00+07',
    registration_end_at = '2025-11-30 23:59:59+07',
    verification_start_at = '2025-12-01 00:00:00+07',
    verification_end_at = '2025-12-07 23:59:59+07',
    campaign_start_at = '2025-12-08 00:00:00+07',
    campaign_end_at = '2025-12-10 23:59:59+07',
    quiet_start_at = '2025-12-11 00:00:00+07',
    quiet_end_at = '2025-12-14 23:59:59+07',
    voting_start_at = '2025-12-15 00:00:00+07',
    voting_end_at = '2025-12-17 23:59:59+07',
    recap_start_at = '2025-12-21 00:00:00+07',
    recap_end_at = '2025-12-22 23:59:59+07',
    status = 'REGISTRATION',
    online_enabled = true,
    tps_enabled = true,
    updated_at = NOW()
WHERE id = 1;

-- Tampilkan hasil
SELECT 
    id, 
    code, 
    name, 
    status,
    registration_start_at,
    registration_end_at,
    verification_start_at,
    verification_end_at,
    campaign_start_at,
    campaign_end_at,
    quiet_start_at,
    quiet_end_at,
    voting_start_at,
    voting_end_at,
    recap_start_at,
    recap_end_at
FROM elections WHERE id = 1;
