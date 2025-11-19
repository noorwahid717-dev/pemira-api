package candidate

import (
"context"
"errors"
"math"
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
ShortBio         string         `json:"short_bio"`
Tagline          string         `json:"tagline"`
FacultyName      string         `json:"faculty_name"`
StudyProgramName string         `json:"study_program_name"`
Status           string         `json:"status"`
Stats            CandidateStats `json:"stats"`
}

// CandidateDetailDTO represents a candidate in detail view
type CandidateDetailDTO struct {
ID               int64          `json:"id"`
ElectionID       int64          `json:"election_id"`
Number           int            `json:"number"`
Name             string         `json:"name"`
PhotoURL         string         `json:"photo_url"`
ShortBio         string         `json:"short_bio"`
LongBio          string         `json:"long_bio"`
Tagline          string         `json:"tagline"`
FacultyName      string         `json:"faculty_name"`
StudyProgramName string         `json:"study_program_name"`
CohortYear       *int           `json:"cohort_year,omitempty"`
Vision           string         `json:"vision"`
Missions         []string       `json:"missions"`
MainPrograms     []MainProgram  `json:"main_programs"`
Media            Media          `json:"media"`
SocialLinks      []SocialLink   `json:"social_links"`
Status           string         `json:"status"`
Stats            CandidateStats `json:"stats"`
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

// ListPublicCandidates returns published candidates for student view
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
Status: ptrStatus(CandidateStatusPublished),
Search: search,
Limit:  limit,
Offset: (page - 1) * limit,
}

candidates, total, err := s.repo.ListByElection(ctx, electionID, filter)
if err != nil {
return nil, Pagination{}, err
}

// Get stats for all candidates
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
if c.Status != CandidateStatusPublished {
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
