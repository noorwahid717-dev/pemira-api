package candidate

import (
"context"
)

// AnalyticsRepository defines the interface for analytics data needed by candidate service
// This should match methods from internal/analytics package
type AnalyticsRepository interface {
// GetCandidateVoteStats returns vote counts and percentages per candidate
// Implementation should use query: queries/01_total_votes_per_candidate.sql
GetCandidateVoteStats(ctx context.Context, electionID int64) (map[int64]CandidateStats, error)
}

// AnalyticsStatsAdapter adapts AnalyticsRepository to StatsProvider interface
type AnalyticsStatsAdapter struct {
analytics AnalyticsRepository
}

// NewAnalyticsStatsAdapter creates a new stats adapter
func NewAnalyticsStatsAdapter(analytics AnalyticsRepository) *AnalyticsStatsAdapter {
return &AnalyticsStatsAdapter{
analytics: analytics,
}
}

// GetCandidateStats implements StatsProvider interface
func (a *AnalyticsStatsAdapter) GetCandidateStats(ctx context.Context, electionID int64) (CandidateStatsMap, error) {
return a.analytics.GetCandidateVoteStats(ctx, electionID)
}
