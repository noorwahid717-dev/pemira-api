package voting

import "time"

type Vote struct {
	ID          int64     `json:"id"`
	ElectionID  int64     `json:"election_id"`
	CandidateID int64     `json:"candidate_id"`
	TokenHash   string    `json:"token_hash"`
	VotedAt     time.Time `json:"voted_at"`
	VotedVia    string    `json:"voted_via"`
}

type VoteToken struct {
	ID         int64     `json:"id"`
	VoterID    int64     `json:"voter_id"`
	ElectionID int64     `json:"election_id"`
	Token      string    `json:"token"`
	UsedAt     *time.Time `json:"used_at"`
	CreatedAt  time.Time `json:"created_at"`
}
