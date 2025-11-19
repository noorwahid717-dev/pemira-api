package candidate

import "context"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Candidate, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListByElection(ctx context.Context, electionID int64) ([]*Candidate, error) {
	return s.repo.ListByElection(ctx, electionID)
}
