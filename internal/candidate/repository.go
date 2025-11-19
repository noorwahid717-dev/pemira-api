package candidate

import "context"

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Candidate, error)
	ListByElection(ctx context.Context, electionID int64) ([]*Candidate, error)
	Create(ctx context.Context, candidate *Candidate) error
	Update(ctx context.Context, candidate *Candidate) error
	Delete(ctx context.Context, id int64) error
}
