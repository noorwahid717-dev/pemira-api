-- +goose Up
-- Add voter_id and tps_id to user_accounts for linking to voters and TPS

ALTER TABLE user_accounts 
ADD COLUMN voter_id BIGINT NULL,
ADD COLUMN tps_id BIGINT NULL;

-- Add foreign key constraints
ALTER TABLE user_accounts
ADD CONSTRAINT fk_user_accounts_voter
FOREIGN KEY (voter_id) REFERENCES voters(id) ON DELETE SET NULL;

-- Note: TPS foreign key will be added when we have proper TPS table structure
-- For now, just add the column

-- Add indexes
CREATE INDEX idx_user_accounts_voter_id ON user_accounts(voter_id) WHERE voter_id IS NOT NULL;
CREATE INDEX idx_user_accounts_tps_id ON user_accounts(tps_id) WHERE tps_id IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_user_accounts_tps_id;
DROP INDEX IF EXISTS idx_user_accounts_voter_id;

ALTER TABLE user_accounts
DROP CONSTRAINT IF EXISTS fk_user_accounts_voter;

ALTER TABLE user_accounts 
DROP COLUMN IF EXISTS tps_id,
DROP COLUMN IF EXISTS voter_id;
