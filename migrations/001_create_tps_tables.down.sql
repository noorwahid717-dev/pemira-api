-- Drop triggers
DROP TRIGGER IF EXISTS update_tps_checkins_updated_at ON tps_checkins;
DROP TRIGGER IF EXISTS update_tps_updated_at ON tps;

-- Drop function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_tps_checkins_scan_at;
DROP INDEX IF EXISTS idx_tps_checkins_status;
DROP INDEX IF EXISTS idx_tps_checkins_election_id;
DROP INDEX IF EXISTS idx_tps_checkins_voter_id;
DROP INDEX IF EXISTS idx_tps_checkins_tps_id;

DROP INDEX IF EXISTS idx_tps_panitia_user_id;
DROP INDEX IF EXISTS idx_tps_panitia_tps_id;

DROP INDEX IF EXISTS idx_tps_qr_active;
DROP INDEX IF EXISTS idx_tps_qr_tps_id;

DROP INDEX IF EXISTS idx_tps_code;
DROP INDEX IF EXISTS idx_tps_status;
DROP INDEX IF EXISTS idx_tps_election_id;

-- Drop tables
DROP TABLE IF EXISTS tps_checkins;
DROP TABLE IF EXISTS tps_panitia;
DROP TABLE IF EXISTS tps_qr;
DROP TABLE IF EXISTS tps;
