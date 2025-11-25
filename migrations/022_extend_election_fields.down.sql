-- +goose Down
-- Revert extended election fields (enum values remain for compatibility)

ALTER TABLE elections
    DROP COLUMN IF EXISTS tps_max,
    DROP COLUMN IF EXISTS tps_require_ballot_qr,
    DROP COLUMN IF EXISTS tps_require_checkin,
    DROP COLUMN IF EXISTS online_max_sessions_per_voter,
    DROP COLUMN IF EXISTS online_login_url,
    DROP COLUMN IF EXISTS current_phase,
    DROP COLUMN IF EXISTS academic_year;
