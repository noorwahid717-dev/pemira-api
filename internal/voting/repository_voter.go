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
		       voting_method, tps_id, voted_at, vote_token_hash
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
		    updated_at = NOW()
		WHERE id = $6
	`
	
	_, err := tx.Exec(ctx, query,
		status.HasVoted,
		status.VotingMethod,
		status.TPSID,
		status.VotedAt,
		status.TokenHash,
		status.ID,
	)
	
	if err != nil {
		return fmt.Errorf("update voter status: %w", err)
	}
	
	return nil
}
