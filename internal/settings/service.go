package settings

import (
	"context"
	"fmt"
	"strconv"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetAll(ctx context.Context) (SettingsResponse, error) {
	activeID, err := s.repo.GetActiveElectionID(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	
	defaultID, err := s.repo.GetDefaultElectionID(ctx)
	if err != nil {
		return SettingsResponse{}, err
	}
	
	return SettingsResponse{
		ActiveElectionID:  activeID,
		DefaultElectionID: defaultID,
	}, nil
}

func (s *Service) GetActiveElectionID(ctx context.Context) (int, error) {
	return s.repo.GetActiveElectionID(ctx)
}

func (s *Service) UpdateActiveElectionID(ctx context.Context, electionID int, updatedBy int64) error {
	if electionID <= 0 {
		return fmt.Errorf("invalid election ID")
	}
	
	value := strconv.Itoa(electionID)
	return s.repo.Update(ctx, "active_election_id", value, updatedBy)
}

func (s *Service) UpdateDefaultElectionID(ctx context.Context, electionID int, updatedBy int64) error {
	if electionID <= 0 {
		return fmt.Errorf("invalid election ID")
	}
	
	value := strconv.Itoa(electionID)
	return s.repo.Update(ctx, "default_election_id", value, updatedBy)
}
