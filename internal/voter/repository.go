package voter

import (
	"context"
	"pemira-api/internal/shared"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Voter, error)
	GetByNIM(ctx context.Context, nim string) (*Voter, error)
	List(ctx context.Context, params shared.PaginationParams) ([]*Voter, int64, error)
	Create(ctx context.Context, voter *Voter) error
	Update(ctx context.Context, voter *Voter) error
	Delete(ctx context.Context, id int64) error
	
	GetElectionStatus(ctx context.Context, voterID, electionID int64) (*VoterElectionStatus, error)
	CreateElectionStatus(ctx context.Context, status *VoterElectionStatus) error
	UpdateElectionStatus(ctx context.Context, status *VoterElectionStatus) error
}
