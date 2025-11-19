package monitoring

import "time"

type VoteStats struct {
	ElectionID      int64     `json:"election_id"`
	CandidateID     int64     `json:"candidate_id"`
	TotalVotes      int64     `json:"total_votes"`
	TotalVotesOnline int64    `json:"total_votes_online"`
	TotalVotesTPS   int64     `json:"total_votes_tps"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type ParticipationStats struct {
	ElectionID      int64   `json:"election_id"`
	TotalEligible   int64   `json:"total_eligible"`
	TotalVoted      int64   `json:"total_voted"`
	ParticipationPct float64 `json:"participation_pct"`
}

type TPSStats struct {
	TPSID         int64  `json:"tps_id"`
	TPSName       string `json:"tps_name"`
	TotalVotes    int64  `json:"total_votes"`
	PendingCheckins int64 `json:"pending_checkins"`
}

type LiveCountSnapshot struct {
	ElectionID    int64                  `json:"election_id"`
	Timestamp     time.Time              `json:"timestamp"`
	TotalVotes    int64                  `json:"total_votes"`
	Participation ParticipationStats     `json:"participation"`
	CandidateVotes map[int64]int64       `json:"candidate_votes"`
	TPSStats      []TPSStats             `json:"tps_stats"`
}
