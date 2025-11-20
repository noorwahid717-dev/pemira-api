-- +goose Down
-- =========================================================
-- Revert tps_qr table structure
-- =========================================================

DROP INDEX IF EXISTS ux_tps_active_qr;
DROP INDEX IF EXISTS ux_tps_qr_token;
CREATE UNIQUE INDEX ux_tps_qr_secret ON tps_qr (qr_token);

ALTER TABLE tps_qr RENAME COLUMN rotated_at TO revoked_at;
ALTER TABLE tps_qr RENAME COLUMN qr_token TO qr_secret;

-- =========================================================
-- Remove PIC fields from TPS table
-- =========================================================

ALTER TABLE tps
    DROP COLUMN IF EXISTS notes,
    DROP COLUMN IF EXISTS pic_phone,
    DROP COLUMN IF EXISTS pic_name;
