package voting

import (
	"context"
	"fmt"
	
	"github.com/jackc/pgx/v5"
	"pemira-api/internal/candidate"
	"pemira-api/internal/shared"
)

type candidateRepository struct{}

func NewCandidateRepository() CandidateRepository {
	return &candidateRepository{}
}

func (r *candidateRepository) GetByIDWithTx(ctx context.Context, tx pgx.Tx, candidateID int64) (*candidate.Candidate, error) {
	query := `
		SELECT id, election_id, order_number, name, vision_mission, photo_url, is_active, created_at, updated_at
		FROM candidates
		WHERE id = $1
	`
	
	var c candidate.Candidate
	
	err := tx.QueryRow(ctx, query, candidateID).Scan(
		&c.ID,
		&c.ElectionID,
		&c.OrderNumber,
		&c.Name,
		&c.VisionMission,
		&c.PhotoURL,
		&c.IsActive,
		&c.CreatedAt,
		&c.UpdatedAt,
	)
	
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, shared.ErrNotFound
		}
		return nil, fmt.Errorf("get candidate: %w", err)
	}
	
	return &c, nil
}
