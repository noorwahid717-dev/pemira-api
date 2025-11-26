package candidate

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
)

// CandidateStatsMap maps candidate ID to their voting statistics
type CandidateStatsMap map[int64]CandidateStats

// StatsProvider provides voting statistics for candidates
type StatsProvider interface {
	// GetCandidateStats returns total votes & percentage per candidate for an election
	GetCandidateStats(ctx context.Context, electionID int64) (CandidateStatsMap, error)
}

// Service provides business logic for candidate operations
type Service struct {
	repo  CandidateRepository
	stats StatsProvider
}

// NewService creates a new candidate service
func NewService(repo CandidateRepository, stats StatsProvider) *Service {
	return &Service{
		repo:  repo,
		stats: stats,
	}
}

// CandidateListItemDTO represents a candidate in list view
type CandidateListItemDTO struct {
	ID               int64          `json:"id"`
	ElectionID       int64          `json:"election_id"`
	Number           int            `json:"number"`
	Name             string         `json:"name"`
	PhotoURL         string         `json:"photo_url"`
	PhotoMediaID     *string        `json:"photo_media_id,omitempty"`
	ShortBio         string         `json:"short_bio"`
	Tagline          string         `json:"tagline"`
	FacultyName      string         `json:"faculty_name"`
	StudyProgramName string         `json:"study_program_name"`
	Status           string         `json:"status"`
	Stats            CandidateStats `json:"stats"`
}

// CandidateDetailDTO represents a candidate in detail view
type CandidateDetailDTO struct {
	ID               int64                `json:"id"`
	ElectionID       int64                `json:"election_id"`
	Number           int                  `json:"number"`
	Name             string               `json:"name"`
	PhotoURL         string               `json:"photo_url"`
	PhotoMediaID     *string              `json:"photo_media_id,omitempty"`
	ShortBio         string               `json:"short_bio"`
	LongBio          string               `json:"long_bio"`
	Tagline          string               `json:"tagline"`
	FacultyName      string               `json:"faculty_name"`
	StudyProgramName string               `json:"study_program_name"`
	CohortYear       *int                 `json:"cohort_year,omitempty"`
	Vision           string               `json:"vision"`
	Missions         []string             `json:"missions"`
	MainPrograms     []MainProgram        `json:"main_programs"`
	Media            Media                `json:"media"`
	MediaFiles       []CandidateMediaMeta `json:"media_files,omitempty"`
	SocialLinks      []SocialLink         `json:"social_links"`
	Status           string               `json:"status"`
	Stats            CandidateStats       `json:"stats"`
}

// Pagination represents pagination metadata
type Pagination struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalItems int64 `json:"total_items"`
	TotalPages int64 `json:"total_pages"`
}

// ErrCandidateNotPublished is returned when trying to access unpublished candidate
var ErrCandidateNotPublished = errors.New("candidate not published")

// ListPublicCandidates returns approved candidates for student view
func (s *Service) ListPublicCandidates(
	ctx context.Context,
	electionID int64,
	search string,
	page, limit int,
) ([]CandidateListItemDTO, Pagination, error) {
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 10
	}

	filter := Filter{
		Status: ptrStatus(CandidateStatusApproved),
		Search: search,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	candidates, total, err := s.repo.ListByElection(ctx, electionID, filter)
	if err != nil {
		return nil, Pagination{}, err
	}

	statsMap, err := s.stats.GetCandidateStats(ctx, electionID)
	if err != nil {
		// Fallback to empty stats if stats service fails
		statsMap = CandidateStatsMap{}
	}

	dtos := make([]CandidateListItemDTO, 0, len(candidates))
	for _, c := range candidates {
		stats := statsMap[c.ID]
		dtos = append(dtos, CandidateListItemDTO{
			ID:               c.ID,
			ElectionID:       c.ElectionID,
			Number:           c.Number,
			Name:             c.Name,
			PhotoURL:         c.PhotoURL,
			PhotoMediaID:     c.PhotoMediaID,
			ShortBio:         c.ShortBio,
			Tagline:          c.Tagline,
			FacultyName:      c.FacultyName,
			StudyProgramName: c.StudyProgramName,
			Status:           string(c.Status),
			Stats:            stats,
		})
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}

	pag := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	return dtos, pag, nil
}

// GetPublicCandidateDetail returns detailed candidate info for student view
func (s *Service) GetPublicCandidateDetail(
	ctx context.Context,
	electionID, candidateID int64,
) (*CandidateDetailDTO, error) {
	c, err := s.repo.GetByID(ctx, electionID, candidateID)
	if err != nil {
		return nil, err
	}

	// Only published candidates are accessible to students
	if c.Status != CandidateStatusApproved {
		return nil, ErrCandidateNotPublished
	}

	// Get stats for this candidate
	statsMap, err := s.stats.GetCandidateStats(ctx, electionID)
	if err != nil {
		statsMap = CandidateStatsMap{}
	}
	stats := statsMap[c.ID]

	dto := &CandidateDetailDTO{
		ID:               c.ID,
		ElectionID:       c.ElectionID,
		Number:           c.Number,
		Name:             c.Name,
		PhotoURL:         c.PhotoURL,
		PhotoMediaID:     c.PhotoMediaID,
		ShortBio:         c.ShortBio,
		LongBio:          c.LongBio,
		Tagline:          c.Tagline,
		FacultyName:      c.FacultyName,
		StudyProgramName: c.StudyProgramName,
		CohortYear:       c.CohortYear,
		Vision:           c.Vision,
		Missions:         c.Missions,
		MainPrograms:     c.MainPrograms,
		Media:            c.Media,
		MediaFiles:       c.MediaFiles,
		SocialLinks:      c.SocialLinks,
		Status:           string(c.Status),
		Stats:            stats,
	}

	return dto, nil
}

// ptrStatus creates a pointer to CandidateStatus
func ptrStatus(s CandidateStatus) *CandidateStatus {
	return &s
}

// AdminListCandidates returns all candidates for admin view (no status filter by default)
func (s *Service) AdminListCandidates(
	ctx context.Context,
	electionID int64,
	search string,
	status *CandidateStatus,
	page, limit int,
) ([]CandidateDetailDTO, Pagination, error) {
	if status != nil && !status.IsValid() {
		return nil, Pagination{}, ErrCandidateStatusInvalid
	}
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	filter := Filter{
		Status: status,
		Search: search,
		Limit:  limit,
		Offset: (page - 1) * limit,
	}

	candidates, total, err := s.repo.ListByElection(ctx, electionID, filter)
	if err != nil {
		// Log error to file
		if f, ferr := os.OpenFile("/tmp/service_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); ferr == nil {
			defer f.Close()
			f.WriteString(fmt.Sprintf("AdminListCandidates - ListByElection error: %v\n", err))
		}
		return nil, Pagination{}, err
	}

	// Get stats for all candidates
	statsMap, err := s.stats.GetCandidateStats(ctx, electionID)
	if err != nil {
		statsMap = CandidateStatsMap{}
	}

	dtos := make([]CandidateDetailDTO, 0, len(candidates))
	for _, c := range candidates {
		stats := statsMap[c.ID]
		dtos = append(dtos, CandidateDetailDTO{
			ID:               c.ID,
			ElectionID:       c.ElectionID,
			Number:           c.Number,
			Name:             c.Name,
			PhotoURL:         c.PhotoURL,
			PhotoMediaID:     c.PhotoMediaID,
			ShortBio:         c.ShortBio,
			LongBio:          c.LongBio,
			Tagline:          c.Tagline,
			FacultyName:      c.FacultyName,
			StudyProgramName: c.StudyProgramName,
			CohortYear:       c.CohortYear,
			Vision:           c.Vision,
			Missions:         c.Missions,
			MainPrograms:     c.MainPrograms,
			Media:            c.Media,
			MediaFiles:       []CandidateMediaMeta{},
			SocialLinks:      c.SocialLinks,
			Status:           string(c.Status),
			Stats:            stats,
		})
	}

	totalPages := int64(0)
	if limit > 0 {
		totalPages = int64(math.Ceil(float64(total) / float64(limit)))
	}

	pag := Pagination{
		Page:       page,
		Limit:      limit,
		TotalItems: total,
		TotalPages: totalPages,
	}

	return dtos, pag, nil
}

// AdminCreateCandidate creates a new candidate
func (s *Service) AdminCreateCandidate(
	ctx context.Context,
	electionID int64,
	req AdminCreateCandidateRequest,
) (*CandidateDetailDTO, error) {
	status := req.Status
	if status == "" {
		status = CandidateStatusPending
	}
	if !status.IsValid() {
		return nil, ErrCandidateStatusInvalid
	}

	// Check if number is already taken
	exists, err := s.repo.CheckNumberExists(ctx, electionID, req.Number, nil)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrCandidateNumberTaken
	}

	// Create candidate entity
	candidate := &Candidate{
		ElectionID:       electionID,
		Number:           req.Number,
		Name:             req.Name,
		PhotoURL:         req.PhotoURL,
		ShortBio:         req.ShortBio,
		LongBio:          req.LongBio,
		Tagline:          req.Tagline,
		FacultyName:      req.FacultyName,
		StudyProgramName: req.StudyProgramName,
		CohortYear:       req.CohortYear,
		Vision:           req.Vision,
		Missions:         req.Missions,
		MainPrograms:     req.MainPrograms,
		Media:            req.Media,
		SocialLinks:      req.SocialLinks,
		Status:           status,
	}

	created, err := s.repo.Create(ctx, candidate)
	if err != nil {
		return nil, err
	}

	// Get stats
	statsMap, _ := s.stats.GetCandidateStats(ctx, electionID)
	stats := statsMap[created.ID]

	return &CandidateDetailDTO{
		ID:               created.ID,
		ElectionID:       created.ElectionID,
		Number:           created.Number,
		Name:             created.Name,
		PhotoURL:         created.PhotoURL,
		PhotoMediaID:     created.PhotoMediaID,
		ShortBio:         created.ShortBio,
		LongBio:          created.LongBio,
		Tagline:          created.Tagline,
		FacultyName:      created.FacultyName,
		StudyProgramName: created.StudyProgramName,
		CohortYear:       created.CohortYear,
		Vision:           created.Vision,
		Missions:         created.Missions,
		MainPrograms:     created.MainPrograms,
		Media:            created.Media,
		MediaFiles:       created.MediaFiles,
		SocialLinks:      created.SocialLinks,
		Status:           string(created.Status),
		Stats:            stats,
	}, nil
}

// AdminGetCandidate returns candidate detail for admin (no status restriction)
func (s *Service) AdminGetCandidate(
	ctx context.Context,
	electionID, candidateID int64,
) (*CandidateDetailDTO, error) {
	c, err := s.repo.GetByID(ctx, electionID, candidateID)
	if err != nil {
		return nil, err
	}

	if mediaFiles, err := s.repo.ListMediaMeta(ctx, candidateID); err == nil {
		c.MediaFiles = mediaFiles
	}

	statsMap, _ := s.stats.GetCandidateStats(ctx, electionID)
	stats := statsMap[c.ID]

	return &CandidateDetailDTO{
		ID:               c.ID,
		ElectionID:       c.ElectionID,
		Number:           c.Number,
		Name:             c.Name,
		PhotoURL:         c.PhotoURL,
		PhotoMediaID:     c.PhotoMediaID,
		ShortBio:         c.ShortBio,
		LongBio:          c.LongBio,
		Tagline:          c.Tagline,
		FacultyName:      c.FacultyName,
		StudyProgramName: c.StudyProgramName,
		CohortYear:       c.CohortYear,
		Vision:           c.Vision,
		Missions:         c.Missions,
		MainPrograms:     c.MainPrograms,
		Media:            c.Media,
		MediaFiles:       c.MediaFiles,
		SocialLinks:      c.SocialLinks,
		Status:           string(c.Status),
		Stats:            stats,
	}, nil
}

// AdminUpdateCandidate updates an existing candidate
func (s *Service) AdminUpdateCandidate(
	ctx context.Context,
	electionID, candidateID int64,
	req AdminUpdateCandidateRequest,
) (*CandidateDetailDTO, error) {
	// Get existing candidate
	existing, err := s.repo.GetByID(ctx, electionID, candidateID)
	if err != nil {
		return nil, err
	}

	// Check if number is being changed and if it's already taken
	if req.Number != nil && *req.Number != existing.Number {
		exists, err := s.repo.CheckNumberExists(ctx, electionID, *req.Number, &candidateID)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, ErrCandidateNumberTaken
		}
	}

	// Apply updates
	if req.Number != nil {
		existing.Number = *req.Number
	}
	if req.Name != nil {
		existing.Name = *req.Name
	}
	if req.PhotoURL != nil {
		existing.PhotoURL = *req.PhotoURL
	}
	if req.ShortBio != nil {
		existing.ShortBio = *req.ShortBio
	}
	if req.LongBio != nil {
		existing.LongBio = *req.LongBio
	}
	if req.Tagline != nil {
		existing.Tagline = *req.Tagline
	}
	if req.FacultyName != nil {
		existing.FacultyName = *req.FacultyName
	}
	if req.StudyProgramName != nil {
		existing.StudyProgramName = *req.StudyProgramName
	}
	if req.CohortYear != nil {
		existing.CohortYear = req.CohortYear
	}
	if req.Vision != nil {
		existing.Vision = *req.Vision
	}
	if req.Missions != nil {
		existing.Missions = *req.Missions
	}
	if req.MainPrograms != nil {
		existing.MainPrograms = *req.MainPrograms
	}
	if req.Media != nil {
		existing.Media = *req.Media
	}
	if req.SocialLinks != nil {
		existing.SocialLinks = *req.SocialLinks
	}
	if req.Status != nil {
		newStatus := *req.Status
		if !newStatus.IsValid() {
			return nil, ErrCandidateStatusInvalid
		}
		existing.Status = newStatus
	}

	updated, err := s.repo.Update(ctx, electionID, candidateID, existing)
	if err != nil {
		return nil, err
	}

	statsMap, _ := s.stats.GetCandidateStats(ctx, electionID)
	stats := statsMap[updated.ID]

	return &CandidateDetailDTO{
		ID:               updated.ID,
		ElectionID:       updated.ElectionID,
		Number:           updated.Number,
		Name:             updated.Name,
		PhotoURL:         updated.PhotoURL,
		PhotoMediaID:     updated.PhotoMediaID,
		ShortBio:         updated.ShortBio,
		LongBio:          updated.LongBio,
		Tagline:          updated.Tagline,
		FacultyName:      updated.FacultyName,
		StudyProgramName: updated.StudyProgramName,
		CohortYear:       updated.CohortYear,
		Vision:           updated.Vision,
		Missions:         updated.Missions,
		MainPrograms:     updated.MainPrograms,
		Media:            updated.Media,
		MediaFiles:       updated.MediaFiles,
		SocialLinks:      updated.SocialLinks,
		Status:           string(updated.Status),
		Stats:            stats,
	}, nil
}

// AdminDeleteCandidate deletes a candidate
func (s *Service) AdminDeleteCandidate(
	ctx context.Context,
	electionID, candidateID int64,
) error {
	return s.repo.Delete(ctx, electionID, candidateID)
}

// AdminPublishCandidate publishes a candidate
func (s *Service) AdminPublishCandidate(
	ctx context.Context,
	electionID, candidateID int64,
) (*CandidateDetailDTO, error) {
	err := s.repo.UpdateStatus(ctx, electionID, candidateID, CandidateStatusApproved)
	if err != nil {
		return nil, err
	}

	return s.AdminGetCandidate(ctx, electionID, candidateID)
}

// AdminUnpublishCandidate unpublishes a candidate
func (s *Service) AdminUnpublishCandidate(
	ctx context.Context,
	electionID, candidateID int64,
) (*CandidateDetailDTO, error) {
	err := s.repo.UpdateStatus(ctx, electionID, candidateID, CandidateStatusPending)
	if err != nil {
		return nil, err
	}

	return s.AdminGetCandidate(ctx, electionID, candidateID)
}

func (s *Service) UploadProfileMedia(ctx context.Context, candidateID int64, media CandidateMediaCreate) (*CandidateMedia, error) {
	return s.repo.SaveProfileMedia(ctx, candidateID, media)
}

func (s *Service) GetProfileMedia(ctx context.Context, candidateID int64) (*CandidateMedia, error) {
	return s.repo.GetProfileMedia(ctx, candidateID)
}

func (s *Service) DeleteProfileMedia(ctx context.Context, candidateID, adminID int64) error {
	return s.repo.DeleteProfileMedia(ctx, candidateID, adminID)
}

func (s *Service) UploadMedia(ctx context.Context, candidateID int64, media CandidateMediaCreate) (*CandidateMedia, error) {
	return s.repo.AddMedia(ctx, candidateID, media)
}

func (s *Service) GetMedia(ctx context.Context, candidateID int64, mediaID string) (*CandidateMedia, error) {
	return s.repo.GetMedia(ctx, candidateID, mediaID)
}

func (s *Service) DeleteMedia(ctx context.Context, candidateID int64, mediaID string) error {
	return s.repo.DeleteMedia(ctx, candidateID, mediaID)
}
