-- +goose Down
-- Revert voting_method changes

ALTER TABLE voter_status
    DROP CONSTRAINT IF EXISTS chk_voter_status_method_has_voted;

ALTER TABLE voter_status
    ADD CONSTRAINT chk_voter_status_method_has_voted
    CHECK (
        (has_voted = FALSE AND voting_method IS NULL AND tps_id IS NULL AND voted_at IS NULL)
     OR (has_voted = TRUE AND voting_method IS NOT NULL AND voted_at IS NOT NULL)
    );

ALTER TABLE voters
    DROP COLUMN IF EXISTS voting_method;
