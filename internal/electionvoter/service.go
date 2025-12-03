package electionvoter

import (
	"context"
	"fmt"
	"strings"

	"pemira-api/internal/shared"
)

var (
	allowedStatuses       = map[string]struct{}{"PENDING": {}, "VERIFIED": {}, "REJECTED": {}, "VOTED": {}, "BLOCKED": {}}
	allowedVotingMethods  = map[string]struct{}{"ONLINE": {}, "TPS": {}}
	allowedAcademicStatus = map[string]struct{}{"ACTIVE": {}, "GRADUATED": {}, "ON_LEAVE": {}, "DROPPED": {}, "INACTIVE": {}}
)

const defaultAcademicStatus = "ACTIVE"

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) LookupByNIM(ctx context.Context, electionID int64, nim string) (*LookupResult, error) {
	nim = strings.TrimSpace(nim)
	if nim == "" {
		return nil, shared.ErrBadRequest
	}
	return s.repo.LookupByNIM(ctx, electionID, nim)
}

func (s *Service) UpsertAndEnroll(ctx context.Context, electionID int64, in UpsertAndEnrollInput) (*UpsertAndEnrollResult, error) {
	if strings.TrimSpace(in.NIM) == "" {
		return nil, shared.ErrBadRequest
	}
	if strings.TrimSpace(in.Name) == "" {
		return nil, shared.ErrBadRequest
	}

	in.VoterType = strings.ToUpper(strings.TrimSpace(in.VoterType))
	if in.VoterType == "" {
		in.VoterType = "STUDENT"
	}

	in.Status = strings.ToUpper(strings.TrimSpace(in.Status))
	if in.Status == "" {
		in.Status = "PENDING"
	}

	if _, ok := allowedStatuses[in.Status]; !ok {
		return nil, shared.ErrBadRequest
	}

	in.VotingMethod = strings.ToUpper(strings.TrimSpace(in.VotingMethod))
	if in.VotingMethod == "" {
		in.VotingMethod = "ONLINE"
	}
	if _, ok := allowedVotingMethods[in.VotingMethod]; !ok {
		return nil, shared.ErrBadRequest
	}

	normalizedStatus, err := normalizeAcademicStatus(in.AcademicStatus)
	if err != nil {
		return nil, err
	}
	in.AcademicStatus = &normalizedStatus

	// Jika semester diberikan, validasi dan hitung cohort_year jika belum ada
	if in.Semester != nil {
		if *in.Semester < 1 || *in.Semester > 20 {
			return nil, shared.ErrBadRequest
		}
		
		// Jika cohort_year tidak diberikan, hitung dari semester
		// Formula: cohort_year = current_year - ((semester - 1) / 2)
		if in.CohortYear == nil && in.VoterType == "STUDENT" {
			currentYear := 2025 // TODO: gunakan tahun sekarang dari context atau config
			cohortYear := currentYear - ((*in.Semester - 1) / 2)
			in.CohortYear = &cohortYear
		}
	}

	return s.repo.UpsertAndEnroll(ctx, electionID, in)
}

func (s *Service) List(ctx context.Context, electionID int64, filter ListFilter, page, limit int) ([]ElectionVoter, shared.PaginationMeta, error) {
	pag := shared.NewPaginationParams(page, limit)
	items, total, err := s.repo.List(ctx, electionID, filter, pag)
	if err != nil {
		return nil, shared.PaginationMeta{}, err
	}
	meta := shared.PaginationMeta{
		CurrentPage: pag.Page,
		PerPage:     pag.PerPage,
		Total:       total,
		TotalPages:  shared.NewPaginatedResponse(nil, pag, total).Meta.TotalPages,
	}
	return items, meta, nil
}

func (s *Service) UpdateEnrollment(ctx context.Context, electionID int64, enrollmentID int64, in UpdateInput) (*ElectionVoter, error) {
	if in.Status != nil {
		status := strings.ToUpper(strings.TrimSpace(*in.Status))
		if status == "" {
			return nil, shared.ErrBadRequest
		}
		if _, ok := allowedStatuses[status]; !ok {
			return nil, shared.ErrBadRequest
		}
		in.Status = &status
	}

	if in.VotingMethod != nil {
		method := strings.ToUpper(strings.TrimSpace(*in.VotingMethod))
		if _, ok := allowedVotingMethods[method]; !ok {
			return nil, shared.ErrBadRequest
		}
		in.VotingMethod = &method
	}

	if in.Semester != nil {
		if *in.Semester < 1 || *in.Semester > 20 {
			return nil, shared.ErrBadRequest
		}
	}

	return s.repo.UpdateEnrollment(ctx, electionID, enrollmentID, in)
}

func (s *Service) SelfRegister(ctx context.Context, electionID, voterID int64, in SelfRegisterInput) (*ElectionVoter, error) {
	in.VotingMethod = strings.ToUpper(strings.TrimSpace(in.VotingMethod))
	if in.VotingMethod == "" {
		in.VotingMethod = "ONLINE"
	}
	if _, ok := allowedVotingMethods[in.VotingMethod]; !ok {
		return nil, shared.ErrBadRequest
	}
	return s.repo.SelfRegister(ctx, electionID, voterID, in)
}

func (s *Service) GetStatus(ctx context.Context, electionID, voterID int64) (*ElectionVoter, error) {
	return s.repo.GetStatus(ctx, electionID, voterID)
}

// ValidateFilter normalizes and validates filter values.
func ValidateFilter(filter ListFilter) (ListFilter, error) {
	filter.VoterType = strings.ToUpper(strings.TrimSpace(filter.VoterType))
	filter.Status = strings.ToUpper(strings.TrimSpace(filter.Status))
	filter.VotingMethod = strings.ToUpper(strings.TrimSpace(filter.VotingMethod))
	if filter.Status != "" {
		if _, ok := allowedStatuses[filter.Status]; !ok {
			return filter, fmt.Errorf("invalid status")
		}
	}
	if filter.VotingMethod != "" {
		if _, ok := allowedVotingMethods[filter.VotingMethod]; !ok {
			return filter, fmt.Errorf("invalid voting_method")
		}
	}
	return filter, nil
}

func normalizeAcademicStatus(raw *string) (string, error) {
	status := defaultAcademicStatus
	if raw != nil {
		normalized := strings.ToUpper(strings.TrimSpace(*raw))
		if normalized != "" {
			status = normalized
		}
	}
	if _, ok := allowedAcademicStatus[status]; !ok {
		return "", shared.ErrBadRequest
	}
	return status, nil
}
