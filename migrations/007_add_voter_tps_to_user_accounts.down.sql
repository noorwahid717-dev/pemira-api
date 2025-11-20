-- +goose Down
-- Rollback voter_id and tps_id from user_accounts

ALTER TABLE user_accounts
    DROP COLUMN IF EXISTS tps_id,
    DROP COLUMN IF EXISTS voter_id;
