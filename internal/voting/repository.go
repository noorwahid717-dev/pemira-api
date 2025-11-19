package voting

import "context"

type Repository interface {
	CreateVote(ctx context.Context, vote *Vote) error
	CreateToken(ctx context.Context, token *VoteToken) error
	GetVoteCount(ctx context.Context, electionID int64) (map[int64]int64, error)
}
