package voter

import (
	"context"
	
	"pemira-api/internal/shared"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByNIM(ctx context.Context, nim string) (*Voter, error) {
	return s.repo.GetByNIM(ctx, nim)
}

func (s *Service) List(ctx context.Context, params shared.PaginationParams) ([]*Voter, int64, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) GetVoterStatus(ctx context.Context, voterID, electionID int64) (*VoterElectionStatus, error) {
	return s.repo.GetElectionStatus(ctx, voterID, electionID)
}
