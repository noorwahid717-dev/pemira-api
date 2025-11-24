package dpt

import (
	"context"
	"math"
	"strings"
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

	// Post-process: detect voter_type for empty ones
	for i := range items {
		if items[i].VoterType == "" {
			items[i].VoterType = detectVoterType(&items[i])
		}
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

// detectVoterType determines voter type based on available data
func detectVoterType(voter *VoterWithStatusDTO) string {
	// Check if semester is valid (not "tidak diisi" or "belum")
	semester := strings.ToLower(strings.TrimSpace(voter.Semester))
	if semester != "" && 
		!strings.Contains(semester, "tidak diisi") && 
		!strings.Contains(semester, "belum") {
		return "STUDENT"
	}
	
	// Check NIM length for lecturer/staff
	nimLen := len(voter.NIM)
	if nimLen >= 18 {
		return "LECTURER"
	}
	if nimLen >= 16 {
		return "STAFF"
	}
	
	// Default to STUDENT
	return "STUDENT"
}
