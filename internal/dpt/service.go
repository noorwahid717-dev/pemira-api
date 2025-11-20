package dpt

import (
	"context"
	"math"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Import(ctx context.Context, electionID int64, rows []ImportRow) (*ImportResult, error) {
	return s.repo.ImportVotersForElection(ctx, electionID, rows)
}

func (s *Service) List(
	ctx context.Context,
	electionID int64,
	filter ListFilter,
	page, limit int,
) ([]VoterWithStatusDTO, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 50
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	items, total, err := s.repo.ListVotersForElection(ctx, electionID, filter)
	if err != nil {
		return nil, Pagination{}, err
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}
	p := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}
	return items, p, nil
}

func (s *Service) ExportStream(
	ctx context.Context,
	electionID int64,
	filter ListFilter,
	fn func(VoterWithStatusDTO) error,
) error {
	return s.repo.StreamVotersForElection(ctx, electionID, filter, fn)
}
