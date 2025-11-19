package analytics

import (
"context"
"fmt"

"golang.org/x/sync/errgroup"
)

// Service handles analytics business logic
type Service struct {
repo AnalyticsRepository
}

// NewService creates a new analytics service
func NewService(repo AnalyticsRepository) *Service {
return &Service{repo: repo}
}

// DashboardCharts contains all chart data for analytics dashboard
type DashboardCharts struct {
HourlyVotes           []HourlyVotes                 `json:"hourly_votes"`
HourlyByCandidate     []HourlyCandidateVotes        `json:"hourly_by_candidate"`
FacultyHeatmap        []FacultyCandidateHeatmapRow  `json:"faculty_heatmap"`
TurnoutTimeline       []TurnoutPoint                `json:"turnout_timeline"`
CohortBreakdown       []CohortCandidateVotes        `json:"cohort_breakdown"`
PeakHours             []PeakHour                    `json:"peak_hours"`
VotingVelocity        *VotingVelocity               `json:"voting_velocity"`
}

// GetDashboardCharts fetches all analytics data in parallel
func (s *Service) GetDashboardCharts(ctx context.Context, electionID int64) (*DashboardCharts, error) {
var result DashboardCharts
g, ctx := errgroup.WithContext(ctx)

// Fetch hourly votes by channel
g.Go(func() error {
data, err := s.repo.GetHourlyVotesByChannel(ctx, electionID)
if err != nil {
return fmt.Errorf("hourly votes: %w", err)
}
result.HourlyVotes = data
return nil
})

// Fetch hourly votes by candidate
g.Go(func() error {
data, err := s.repo.GetHourlyVotesByCandidate(ctx, electionID)
if err != nil {
return fmt.Errorf("hourly by candidate: %w", err)
}
result.HourlyByCandidate = data
return nil
})

// Fetch faculty heatmap
g.Go(func() error {
data, err := s.repo.GetFacultyCandidateHeatmap(ctx, electionID)
if err != nil {
return fmt.Errorf("faculty heatmap: %w", err)
}
result.FacultyHeatmap = data
return nil
})

// Fetch turnout timeline
g.Go(func() error {
data, err := s.repo.GetTurnoutTimeline(ctx, electionID)
if err != nil {
return fmt.Errorf("turnout timeline: %w", err)
}
result.TurnoutTimeline = data
return nil
})

// Fetch cohort breakdown
g.Go(func() error {
data, err := s.repo.GetCohortCandidateVotes(ctx, electionID)
if err != nil {
return fmt.Errorf("cohort breakdown: %w", err)
}
result.CohortBreakdown = data
return nil
})

// Fetch peak hours
g.Go(func() error {
data, err := s.repo.GetPeakHours(ctx, electionID)
if err != nil {
return fmt.Errorf("peak hours: %w", err)
}
result.PeakHours = data
return nil
})

// Fetch voting velocity
g.Go(func() error {
data, err := s.repo.GetVotingVelocity(ctx, electionID)
if err != nil {
return fmt.Errorf("voting velocity: %w", err)
}
result.VotingVelocity = data
return nil
})

if err := g.Wait(); err != nil {
return nil, err
}

return &result, nil
}

// GetHourlyVotesByChannel wraps repository method
func (s *Service) GetHourlyVotesByChannel(ctx context.Context, electionID int64) ([]HourlyVotes, error) {
return s.repo.GetHourlyVotesByChannel(ctx, electionID)
}

// GetHourlyVotesByCandidate wraps repository method
func (s *Service) GetHourlyVotesByCandidate(ctx context.Context, electionID int64) ([]HourlyCandidateVotes, error) {
return s.repo.GetHourlyVotesByCandidate(ctx, electionID)
}

// GetFacultyCandidateHeatmap wraps repository method
func (s *Service) GetFacultyCandidateHeatmap(ctx context.Context, electionID int64) ([]FacultyCandidateHeatmapRow, error) {
return s.repo.GetFacultyCandidateHeatmap(ctx, electionID)
}

// GetTurnoutTimeline wraps repository method
func (s *Service) GetTurnoutTimeline(ctx context.Context, electionID int64) ([]TurnoutPoint, error) {
return s.repo.GetTurnoutTimeline(ctx, electionID)
}

// GetCohortCandidateVotes wraps repository method
func (s *Service) GetCohortCandidateVotes(ctx context.Context, electionID int64) ([]CohortCandidateVotes, error) {
return s.repo.GetCohortCandidateVotes(ctx, electionID)
}

// GetPeakHours wraps repository method
func (s *Service) GetPeakHours(ctx context.Context, electionID int64) ([]PeakHour, error) {
return s.repo.GetPeakHours(ctx, electionID)
}

// GetVotingVelocity wraps repository method
func (s *Service) GetVotingVelocity(ctx context.Context, electionID int64) (*VotingVelocity, error) {
return s.repo.GetVotingVelocity(ctx, electionID)
}
