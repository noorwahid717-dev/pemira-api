package election

import "context"

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Election, error)
	GetCurrent(ctx context.Context) (*Election, error)
	List(ctx context.Context) ([]*Election, error)
	Create(ctx context.Context, election *Election) error
	Update(ctx context.Context, election *Election) error
	
	GetPhases(ctx context.Context, electionID int64) ([]*ElectionPhaseSchedule, error)
	CreatePhase(ctx context.Context, phase *ElectionPhaseSchedule) error
	UpdatePhase(ctx context.Context, phase *ElectionPhaseSchedule) error
}
