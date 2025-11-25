-- Seed Election Schedule for Pemira 2025
-- Mengatur jadwal tahapan pemilihan raya

-- Election ID 1: Pemilihan Raya BEM 2024
UPDATE election_phases SET 
  start_at = '2025-11-01 00:00:00+07',
  end_at = '2025-11-30 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'REGISTRATION';

UPDATE election_phases SET 
  start_at = '2025-12-01 00:00:00+07',
  end_at = '2025-12-07 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'VERIFICATION';

UPDATE election_phases SET 
  start_at = '2025-12-08 00:00:00+07',
  end_at = '2025-12-10 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'CAMPAIGN';

UPDATE election_phases SET 
  start_at = '2025-12-11 00:00:00+07',
  end_at = '2025-12-14 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'QUIET_PERIOD';

UPDATE election_phases SET 
  start_at = '2025-12-15 00:00:00+07',
  end_at = '2025-12-17 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'VOTING';

UPDATE election_phases SET 
  start_at = '2025-12-21 00:00:00+07',
  end_at = '2025-12-22 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 1 AND phase_key = 'RECAP';

-- Election ID 2: Pemilihan Raya BEM 2025
UPDATE election_phases SET 
  start_at = '2025-11-01 00:00:00+07',
  end_at = '2025-11-30 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'REGISTRATION';

UPDATE election_phases SET 
  start_at = '2025-12-01 00:00:00+07',
  end_at = '2025-12-07 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'VERIFICATION';

UPDATE election_phases SET 
  start_at = '2025-12-08 00:00:00+07',
  end_at = '2025-12-10 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'CAMPAIGN';

UPDATE election_phases SET 
  start_at = '2025-12-11 00:00:00+07',
  end_at = '2025-12-14 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'QUIET_PERIOD';

UPDATE election_phases SET 
  start_at = '2025-12-15 00:00:00+07',
  end_at = '2025-12-17 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'VOTING';

UPDATE election_phases SET 
  start_at = '2025-12-21 00:00:00+07',
  end_at = '2025-12-22 23:59:59+07',
  updated_at = NOW()
WHERE election_id = 2 AND phase_key = 'RECAP';

-- Verify the changes
SELECT 
  e.id AS election_id,
  e.name AS election_name,
  ep.phase_key,
  ep.phase_label,
  ep.start_at,
  ep.end_at
FROM election_phases ep
JOIN elections e ON e.id = ep.election_id
ORDER BY e.id, ep.phase_order;
