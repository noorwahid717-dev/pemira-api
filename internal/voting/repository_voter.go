package voting

import (
	"context"
	"fmt"
	
	"github.com/jackc/pgx/v5"
	"pemira-api/internal/shared"
	"pemira-api/internal/shared/constants"
)

type voterRepository struct{}

func NewVoterRepository() VoterRepository {
	return &voterRepository{}
}

func (r *voterRepository) GetStatusForUpdate(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatusEntity, error) {
	query := `
		SELECT id, election_id, voter_id, has_voted, status, voted_via, tps_id, voted_at, token_hash
		FROM voter_election_status
		WHERE election_id = $1 AND voter_id = $2
		FOR UPDATE
	`
	
	var vs VoterStatusEntity
	var status string
	
	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(
		&vs.ID,
		&vs.ElectionID,
		&vs.VoterID,
		&vs.HasVoted,
		&status,
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
	
	vs.Status = status
	vs.IsEligible = (status == string(constants.VoterStatusEligible))
	
	return &vs, nil
}

func (r *voterRepository) UpdateStatus(ctx context.Context, tx pgx.Tx, status *VoterStatusEntity) error {
	query := `
		UPDATE voter_election_status
		SET has_voted = $1,
		    voted_via = $2,
		    tps_id = $3,
		    voted_at = $4,
		    token_hash = $5,
		    status = $6,
		    updated_at = NOW()
		WHERE id = $7
	`
	
	newStatus := status.Status
	if status.HasVoted {
		newStatus = string(constants.VoterStatusVoted)
	}
	
	_, err := tx.Exec(ctx, query,
		status.HasVoted,
		status.VotingMethod,
		status.TPSID,
		status.VotedAt,
		status.TokenHash,
		newStatus,
		status.ID,
	)
	
	if err != nil {
		return fmt.Errorf("update voter status: %w", err)
	}
	
	return nil
}
