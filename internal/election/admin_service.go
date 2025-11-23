package election

import (
	"context"
	"errors"
	"time"
)

type AdminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

var (
	ErrElectionAlreadyOpen    = errors.New("election already open for voting")
	ErrElectionNotInOpenState = errors.New("election is not in voting-open state")
	ErrInvalidStatusChange    = errors.New("invalid election status change")
)

func (s *AdminService) List(
	ctx context.Context,
	filter AdminElectionListFilter,
	page, limit int,
) ([]AdminElectionDTO, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	items, total, err := s.repo.ListElections(ctx, filter)
	if err != nil {
		return nil, Pagination{}, err
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = (total + int64(limit) - 1) / int64(limit)
	}

	p := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}
	return items, p, nil
}

func (s *AdminService) Create(ctx context.Context, req AdminElectionCreateRequest) (*AdminElectionDTO, error) {
	return s.repo.CreateElection(ctx, req)
}

func (s *AdminService) Get(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	return s.repo.GetElectionByID(ctx, id)
}

func (s *AdminService) Update(ctx context.Context, id int64, req AdminElectionUpdateRequest) (*AdminElectionDTO, error) {
	return s.repo.UpdateElection(ctx, id, req)
}

func (s *AdminService) OpenVoting(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Status == ElectionStatusVotingOpen {
		return nil, ErrElectionAlreadyOpen
	}
	if e.Status == ElectionStatusArchived {
		return nil, ErrInvalidStatusChange
	}

	now := time.Now().UTC()
	startAt := e.VotingStartAt
	if startAt == nil {
		startAt = &now
	}

	return s.repo.SetVotingStatus(ctx, id, ElectionStatusVotingOpen, startAt, nil)
}

func (s *AdminService) CloseVoting(ctx context.Context, id int64) (*AdminElectionDTO, error) {
	e, err := s.repo.GetElectionByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if e.Status != ElectionStatusVotingOpen {
		return nil, ErrElectionNotInOpenState
	}

	now := time.Now().UTC()
	return s.repo.SetVotingStatus(ctx, id, ElectionStatusVotingClosed, nil, &now)
}

func (s *AdminService) GetBranding(ctx context.Context, electionID int64) (*BrandingSettings, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.GetBranding(ctx, electionID)
}

func (s *AdminService) GetBrandingLogo(ctx context.Context, electionID int64, slot BrandingSlot) (*BrandingFile, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.GetBrandingFile(ctx, electionID, slot)
}

func (s *AdminService) UploadBrandingLogo(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	file BrandingFileCreate,
) (*BrandingFile, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.SaveBrandingFile(ctx, electionID, slot, file)
}

func (s *AdminService) DeleteBrandingLogo(
	ctx context.Context,
	electionID int64,
	slot BrandingSlot,
	adminID int64,
) (*BrandingSettings, error) {
	if _, err := s.repo.GetElectionByID(ctx, electionID); err != nil {
		return nil, err
	}
	return s.repo.DeleteBrandingFile(ctx, electionID, slot, adminID)
}
