-- +goose Down
-- Non-destructive down migration: keep data, just drop added objects (optional)
DROP TABLE IF EXISTS tps_ballot_scans;
DROP TABLE IF EXISTS candidate_qr_codes;
-- Note: columns/enum values added are left in place to avoid data loss.
