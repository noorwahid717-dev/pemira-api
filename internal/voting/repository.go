package voting

import (
	"context"
	"time"
	
	"github.com/jackc/pgx/v5"
	"pemira-api/internal/candidate"
	"pemira-api/internal/tps"
)

// VoterRepository handles voter status operations
type VoterRepository interface {
	// GetStatusForUpdate gets voter status with row-level lock (FOR UPDATE)
	GetStatusForUpdate(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatusEntity, error)
	
	// UpdateStatus updates voter status
	UpdateStatus(ctx context.Context, tx pgx.Tx, status *VoterStatusEntity) error
}

// CandidateRepository handles candidate operations within transaction
type CandidateRepository interface {
	// GetByIDWithTx gets candidate by ID within transaction
	GetByIDWithTx(ctx context.Context, tx pgx.Tx, candidateID int64) (*candidate.Candidate, error)
}

// VoteRepository handles vote and token operations
type VoteRepository interface {
	// InsertToken inserts a new vote token
	InsertToken(ctx context.Context, tx pgx.Tx, token *VoteToken) error
	
	// InsertVote inserts a new vote
	InsertVote(ctx context.Context, tx pgx.Tx, vote *Vote) error
	
	// MarkTokenUsed marks a token as used
	MarkTokenUsed(ctx context.Context, tx pgx.Tx, electionID int64, tokenHash string, usedAt time.Time) error
	
	// GetLatestApprovedCheckin gets the latest approved TPS check-in for a voter
	GetLatestApprovedCheckin(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*tps.TPSCheckin, error)
	
	// GetTPSByID gets TPS information by ID
	GetTPSByID(ctx context.Context, tx pgx.Tx, tpsID int64) (*tps.TPS, error)
	
	// MarkCheckinUsed marks a check-in as used
	MarkCheckinUsed(ctx context.Context, tx pgx.Tx, checkinID int64, usedAt time.Time) error
	
	// GetCheckinByID returns checkin by ID with lock
	GetCheckinByID(ctx context.Context, tx pgx.Tx, checkinID int64) (*tps.TPSCheckin, error)
}

// VoteStatsRepository handles vote statistics (optional, for read-model)
type VoteStatsRepository interface {
	// IncrementCandidateCount increments vote count for a candidate
	IncrementCandidateCount(ctx context.Context, tx pgx.Tx, electionID, candidateID int64, channel string, tpsID *int64) error
}

// Repository is the legacy interface (kept for backward compatibility)
type Repository interface {
	CreateVote(ctx context.Context, vote *Vote) error
	CreateToken(ctx context.Context, token *VoteToken) error
	GetVoteCount(ctx context.Context, electionID int64) (map[int64]int64, error)
}
