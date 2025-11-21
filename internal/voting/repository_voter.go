package voting

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"pemira-api/internal/shared"
)

type voterRepository struct{}

func NewVoterRepository() VoterRepository {
	return &voterRepository{}
}

func (r *voterRepository) GetStatusForUpdate(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatusEntity, error) {
	query := `
		SELECT id, election_id, voter_id, is_eligible, has_voted, 
		       voting_method, tps_id, voted_at, vote_token_hash,
		       preferred_method, online_allowed, tps_allowed
		FROM voter_status
		WHERE election_id = $1 AND voter_id = $2
		FOR UPDATE
	`

	var vs VoterStatusEntity

	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(
		&vs.ID,
		&vs.ElectionID,
		&vs.VoterID,
		&vs.IsEligible,
		&vs.HasVoted,
		&vs.VotingMethod,
		&vs.TPSID,
		&vs.VotedAt,
		&vs.TokenHash,
		&vs.PreferredMethod,
		&vs.OnlineAllowed,
		&vs.TPSAllowed,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrVoterNotEligible
		}
		return nil, fmt.Errorf("get voter status: %w", err)
	}

	return &vs, nil
}

func (r *voterRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, status *VoterStatusEntity) error {
	query := `
		UPDATE voter_status
		SET has_voted = $1,
		    voting_method = $2,
		    tps_id = $3,
		    voted_at = $4,
		    vote_token_hash = $5,
		    preferred_method = COALESCE($6, preferred_method),
		    online_allowed = $7,
		    tps_allowed = $8,
		    updated_at = NOW()
		WHERE id = $9
	`

	_, err := tx.Exec(ctx, query,
		status.HasVoted,
		status.VotingMethod,
		status.TPSID,
		status.VotedAt,
		status.TokenHash,
		status.PreferredMethod,
		status.OnlineAllowed,
		status.TPSAllowed,
		status.ID,
	)

	if err != nil {
		return fmt.Errorf("update voter status: %w", err)
	}

	return nil
}

// EnsureStatus ensures a voter_status row exists and updates preferred/allowed flags.
func (r *voterRepository) EnsureStatus(ctx context.Context, tx pgx.Tx, electionID, voterID int64, preferred string, onlineAllowed, tpsAllowed bool) (*VoterStatusEntity, error) {
	query := `
		INSERT INTO voter_status (election_id, voter_id, is_eligible, has_voted, preferred_method, online_allowed, tps_allowed)
		VALUES ($1,$2,TRUE,FALSE,$3,$4,$5)
		ON CONFLICT (election_id, voter_id)
		DO UPDATE SET preferred_method = EXCLUDED.preferred_method,
		              online_allowed = EXCLUDED.online_allowed,
		              tps_allowed = EXCLUDED.tps_allowed,
		              updated_at = NOW()
		RETURNING id, election_id, voter_id, is_eligible, has_voted, voting_method, tps_id, voted_at, vote_token_hash, preferred_method, online_allowed, tps_allowed, created_at, updated_at
	`

	var vs VoterStatusEntity
	err := tx.QueryRow(ctx, query, electionID, voterID, preferred, onlineAllowed, tpsAllowed).Scan(
		&vs.ID,
		&vs.ElectionID,
		&vs.VoterID,
		&vs.IsEligible,
		&vs.HasVoted,
		&vs.VotingMethod,
		&vs.TPSID,
		&vs.VotedAt,
		&vs.TokenHash,
		&vs.PreferredMethod,
		&vs.OnlineAllowed,
		&vs.TPSAllowed,
		&vs.CreatedAt,
		&vs.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("ensure voter_status: %w", err)
	}
	return &vs, nil
}
