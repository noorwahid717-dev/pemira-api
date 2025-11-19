package analytics

import "time"

// HourlyVotes represents votes aggregated by hour with channel breakdown
type HourlyVotes struct {
BucketStart time.Time `json:"bucket_start"`
TotalVotes  int64     `json:"total_votes"`
VotesOnline int64     `json:"votes_online"`
VotesTPS    int64     `json:"votes_tps"`
}

// HourlyCandidateVotes represents votes per candidate per hour
type HourlyCandidateVotes struct {
BucketStart     time.Time `json:"bucket_start"`
CandidateID     int64     `json:"candidate_id"`
CandidateNumber int       `json:"candidate_number"`
CandidateName   string    `json:"candidate_name"`
TotalVotes      int64     `json:"total_votes"`
}

// FacultyCandidateHeatmapRow represents faculty preference for candidates
type FacultyCandidateHeatmapRow struct {
FacultyCode      string  `json:"faculty_code"`
FacultyName      string  `json:"faculty_name"`
CandidateID      int64   `json:"candidate_id"`
CandidateNumber  int     `json:"candidate_number"`
CandidateName    string  `json:"candidate_name"`
TotalVotes       int64   `json:"total_votes"`
PercentInFaculty float64 `json:"percent_in_faculty"`
}

// TurnoutPoint represents cumulative turnout at a specific time
type TurnoutPoint struct {
BucketStart              time.Time `json:"bucket_start"`
VotesInHour              int64     `json:"votes_in_hour"`
CumulativeVotes          int64     `json:"cumulative_votes"`
CumulativeTurnoutPercent float64   `json:"cumulative_turnout_percent"`
}

// CohortCandidateVotes represents votes breakdown by cohort year
type CohortCandidateVotes struct {
CohortYear      int    `json:"cohort_year"`
CandidateID     int64  `json:"candidate_id"`
CandidateNumber int    `json:"candidate_number"`
CandidateName   string `json:"candidate_name"`
TotalVotes      int64  `json:"total_votes"`
}

// PeakHour represents voting activity at peak hours
type PeakHour struct {
VoteHour    time.Time `json:"vote_hour"`
HourOfDay   int       `json:"hour_of_day"`
DayName     string    `json:"day_name"`
TotalVotes  int64     `json:"total_votes"`
VotesOnline int64     `json:"votes_online"`
VotesTPS    int64     `json:"votes_tps"`
RankByTotal int       `json:"rank_by_total"`
}

// VotingVelocity represents statistical metrics of voting speed
type VotingVelocity struct {
TotalIntervals   int64   `json:"total_intervals"`
AvgGapMinutes    float64 `json:"avg_gap_minutes"`
MinGapMinutes    float64 `json:"min_gap_minutes"`
MaxGapMinutes    float64 `json:"max_gap_minutes"`
MedianGapMinutes float64 `json:"median_gap_minutes"`
P95GapMinutes    float64 `json:"p95_gap_minutes"`
}
