package voting

import (
	"context"
	"fmt"
	
	"github.com/jackc/pgx/v5"
)

type voteStatsRepository struct{}

func NewVoteStatsRepository() VoteStatsRepository {
	return &voteStatsRepository{}
}

func (r *voteStatsRepository) IncrementCandidateCount(ctx context.Context, tx pgx.Tx, electionID, candidateID int64, channel string, tpsID *int64) error {
	query := `
		INSERT INTO vote_stats (election_id, candidate_id, total_votes, updated_at)
		VALUES ($1, $2, 1, NOW())
		ON CONFLICT (election_id, candidate_id)
		DO UPDATE SET
			total_votes = vote_stats.total_votes + 1,
			updated_at = NOW()
	`
	
	_, err := tx.Exec(ctx, query, electionID, candidateID)
	if err != nil {
		return fmt.Errorf("increment candidate count: %w", err)
	}
	
	return nil
}
