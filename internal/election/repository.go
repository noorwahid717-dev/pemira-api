package election

import "context"

type Repository interface {
	GetCurrentElection(ctx context.Context) (*Election, error)
	ListPublicElections(ctx context.Context) ([]Election, error)
	GetByID(ctx context.Context, id int64) (*Election, error)
	GetVoterStatus(ctx context.Context, electionID, voterID int64) (*MeStatusRow, error)
}
