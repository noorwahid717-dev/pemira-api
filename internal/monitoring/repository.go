package monitoring

import "context"

type Repository interface {
	GetVoteStats(ctx context.Context, electionID int64) ([]*VoteStats, error)
	GetParticipationStats(ctx context.Context, electionID int64) (*ParticipationStats, error)
	GetTPSStats(ctx context.Context, electionID int64) ([]*TPSStats, error)
	GetLiveCount(ctx context.Context, electionID int64) (map[int64]int64, error)
}
