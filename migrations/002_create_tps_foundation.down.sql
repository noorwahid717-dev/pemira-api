-- +goose Down
-- Drop triggers
DROP TRIGGER IF EXISTS update_voter_status_updated_at ON voter_status;
DROP TRIGGER IF EXISTS update_tps_updated_at ON tps;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop tables (reverse order)
DROP TABLE IF EXISTS votes;
DROP TABLE IF EXISTS voter_status;
DROP TABLE IF EXISTS tps_checkins;
DROP TABLE IF EXISTS tps_qr;
DROP TABLE IF EXISTS tps;

-- Drop enums
DROP TYPE IF EXISTS vote_channel;
DROP TYPE IF EXISTS voting_method;
DROP TYPE IF EXISTS tps_checkin_status;
DROP TYPE IF EXISTS tps_status;
