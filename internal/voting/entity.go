package voting

import "time"

type Vote struct {
	ID          int64      `json:"id"`
	ElectionID  int64      `json:"election_id"`
	CandidateID int64      `json:"candidate_id"`
	TokenHash   string     `json:"token_hash"`
	Channel     string     `json:"channel"`      // "ONLINE" | "TPS"
	TPSID       *int64     `json:"tps_id"`
	CastAt      time.Time  `json:"cast_at"`
}

type VoteToken struct {
	ID         int64      `json:"id"`
	ElectionID int64      `json:"election_id"`
	VoterID    int64      `json:"voter_id"`
	TokenHash  string     `json:"token_hash"`
	IssuedAt   time.Time  `json:"issued_at"`
	UsedAt     *time.Time `json:"used_at"`
	Method     string     `json:"method"`       // "ONLINE" | "TPS"
	TPSID      *int64     `json:"tps_id"`
}

type VoterStatusEntity struct {
	ID           int64      `json:"id"`
	ElectionID   int64      `json:"election_id"`
	VoterID      int64      `json:"voter_id"`
	IsEligible   bool       `json:"is_eligible"`
	HasVoted     bool       `json:"has_voted"`
	VotingMethod *string    `json:"voting_method"`   // "ONLINE" | "TPS"
	TPSID        *int64     `json:"tps_id"`
	VotedAt      *time.Time `json:"voted_at"`
	TokenHash    *string    `json:"token_hash"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

type VoteResultEntity struct {
	ElectionID int64    `json:"election_id"`
	VoterID    int64    `json:"voter_id"`
	Method     string   `json:"method"`
	VotedAt    time.Time `json:"voted_at"`
	TPS        *TPSInfo  `json:"tps,omitempty"`
	Receipt    ReceiptDetail `json:"receipt"`
}
