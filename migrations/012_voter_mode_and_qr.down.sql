-- +goose Down
-- Drop only the QR table; keep added columns to avoid data loss
DROP TABLE IF EXISTS voter_tps_qr;
