package analytics

import (
"context"
_ "embed"

"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed ../../queries/analytics_02_timeline_votes_by_channel.sql
var qGetHourlyVotesByChannel string

//go:embed ../../queries/analytics_03_timeline_votes_per_candidate.sql
var qGetHourlyVotesByCandidate string

//go:embed ../../queries/analytics_05_heatmap_faculty_candidate_percent.sql
var qFacultyCandidateHeatmap string

//go:embed ../../queries/analytics_06_turnout_cumulative_timeline.sql
var qTurnoutTimeline string

//go:embed ../../queries/analytics_07_votes_by_cohort_candidate.sql
var qCohortCandidateVotes string

//go:embed ../../queries/analytics_09_peak_hours_analysis.sql
var qPeakHoursAnalysis string

//go:embed ../../queries/analytics_10_voting_velocity.sql
var qVotingVelocity string

// AnalyticsRepository defines the interface for analytics data access
type AnalyticsRepository interface {
GetHourlyVotesByChannel(ctx context.Context, electionID int64) ([]HourlyVotes, error)
GetHourlyVotesByCandidate(ctx context.Context, electionID int64) ([]HourlyCandidateVotes, error)
GetFacultyCandidateHeatmap(ctx context.Context, electionID int64) ([]FacultyCandidateHeatmapRow, error)
GetTurnoutTimeline(ctx context.Context, electionID int64) ([]TurnoutPoint, error)
GetCohortCandidateVotes(ctx context.Context, electionID int64) ([]CohortCandidateVotes, error)
GetPeakHours(ctx context.Context, electionID int64) ([]PeakHour, error)
GetVotingVelocity(ctx context.Context, electionID int64) (*VotingVelocity, error)
}

// AnalyticsRepo implements AnalyticsRepository using pgxpool
type AnalyticsRepo struct {
db *pgxpool.Pool
}

// NewAnalyticsRepo creates a new analytics repository
func NewAnalyticsRepo(db *pgxpool.Pool) *AnalyticsRepo {
return &AnalyticsRepo{db: db}
}

// GetHourlyVotesByChannel returns hourly vote counts with channel breakdown
func (r *AnalyticsRepo) GetHourlyVotesByChannel(
ctx context.Context,
electionID int64,
) ([]HourlyVotes, error) {
rows, err := r.db.Query(ctx, qGetHourlyVotesByChannel, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []HourlyVotes
for rows.Next() {
var hv HourlyVotes
if err := rows.Scan(
&hv.BucketStart,
&hv.VotesOnline,
&hv.VotesTPS,
&hv.TotalVotes,
); err != nil {
return nil, err
}
result = append(result, hv)
}
if rows.Err() != nil {
return nil, rows.Err()
}

return result, nil
}

// GetHourlyVotesByCandidate returns hourly vote counts per candidate
func (r *AnalyticsRepo) GetHourlyVotesByCandidate(
ctx context.Context,
electionID int64,
) ([]HourlyCandidateVotes, error) {
rows, err := r.db.Query(ctx, qGetHourlyVotesByCandidate, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []HourlyCandidateVotes
for rows.Next() {
var row HourlyCandidateVotes
if err := rows.Scan(
&row.BucketStart,
&row.CandidateID,
&row.CandidateNumber,
&row.CandidateName,
&row.TotalVotes,
); err != nil {
return nil, err
}
result = append(result, row)
}
if rows.Err() != nil {
return nil, rows.Err()
}
return result, nil
}

// GetFacultyCandidateHeatmap returns faculty-candidate preference matrix with percentages
func (r *AnalyticsRepo) GetFacultyCandidateHeatmap(
ctx context.Context,
electionID int64,
) ([]FacultyCandidateHeatmapRow, error) {
rows, err := r.db.Query(ctx, qFacultyCandidateHeatmap, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []FacultyCandidateHeatmapRow
for rows.Next() {
var row FacultyCandidateHeatmapRow
if err := rows.Scan(
&row.FacultyCode,
&row.FacultyName,
&row.CandidateID,
&row.CandidateNumber,
&row.CandidateName,
&row.TotalVotes,
&row.PercentInFaculty,
); err != nil {
return nil, err
}
result = append(result, row)
}
if rows.Err() != nil {
return nil, rows.Err()
}
return result, nil
}

// GetTurnoutTimeline returns cumulative turnout progression over time
func (r *AnalyticsRepo) GetTurnoutTimeline(
ctx context.Context,
electionID int64,
) ([]TurnoutPoint, error) {
rows, err := r.db.Query(ctx, qTurnoutTimeline, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []TurnoutPoint
for rows.Next() {
var row TurnoutPoint
if err := rows.Scan(
&row.BucketStart,
&row.VotesInHour,
&row.CumulativeVotes,
&row.CumulativeTurnoutPercent,
); err != nil {
return nil, err
}
result = append(result, row)
}
if rows.Err() != nil {
return nil, rows.Err()
}
return result, nil
}

// GetCohortCandidateVotes returns vote breakdown by cohort year and candidate
func (r *AnalyticsRepo) GetCohortCandidateVotes(
ctx context.Context,
electionID int64,
) ([]CohortCandidateVotes, error) {
rows, err := r.db.Query(ctx, qCohortCandidateVotes, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []CohortCandidateVotes
for rows.Next() {
var row CohortCandidateVotes
if err := rows.Scan(
&row.CohortYear,
&row.CandidateID,
&row.CandidateNumber,
&row.CandidateName,
&row.TotalVotes,
); err != nil {
return nil, err
}
result = append(result, row)
}
if rows.Err() != nil {
return nil, rows.Err()
}
return result, nil
}

// GetPeakHours returns ranking of busiest voting hours
func (r *AnalyticsRepo) GetPeakHours(
ctx context.Context,
electionID int64,
) ([]PeakHour, error) {
rows, err := r.db.Query(ctx, qPeakHoursAnalysis, electionID)
if err != nil {
return nil, err
}
defer rows.Close()

var result []PeakHour
for rows.Next() {
var row PeakHour
if err := rows.Scan(
&row.VoteHour,
&row.HourOfDay,
&row.DayName,
&row.TotalVotes,
&row.VotesOnline,
&row.VotesTPS,
&row.RankByTotal,
); err != nil {
return nil, err
}
result = append(result, row)
}
if rows.Err() != nil {
return nil, rows.Err()
}
return result, nil
}

// GetVotingVelocity returns statistical metrics of voting speed
func (r *AnalyticsRepo) GetVotingVelocity(
ctx context.Context,
electionID int64,
) (*VotingVelocity, error) {
var result VotingVelocity
err := r.db.QueryRow(ctx, qVotingVelocity, electionID).Scan(
&result.TotalIntervals,
&result.AvgGapMinutes,
&result.MinGapMinutes,
&result.MaxGapMinutes,
&result.MedianGapMinutes,
&result.P95GapMinutes,
)
if err != nil {
return nil, err
}
return &result, nil
}
