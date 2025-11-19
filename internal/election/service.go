package election

import (
	"context"
	"time"
	
	"pemira-api/internal/shared"
	"pemira-api/internal/shared/constants"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetCurrent(ctx context.Context) (*Election, error) {
	return s.repo.GetCurrent(ctx)
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Election, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) GetCurrentPhase(ctx context.Context, electionID int64) (*constants.ElectionPhase, error) {
	phases, err := s.repo.GetPhases(ctx, electionID)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	for _, phase := range phases {
		if now.After(phase.StartDate) && now.Before(phase.EndDate) {
			return &phase.Phase, nil
		}
	}

	return nil, shared.ErrNotFound
}

func (s *Service) CanVote(ctx context.Context, electionID int64) (bool, error) {
	phase, err := s.GetCurrentPhase(ctx, electionID)
	if err != nil {
		return false, err
	}

	return *phase == constants.PhaseVoting, nil
}
