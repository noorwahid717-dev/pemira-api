package election

import "time"

type VoteMethod string

const (
	VoteMethodNone   VoteMethod = "NONE"
	VoteMethodOnline VoteMethod = "ONLINE"
	VoteMethodTPS    VoteMethod = "TPS"
)

type MeStatusDTO struct {
	ElectionID    int64      `json:"election_id"`
	VoterID       int64      `json:"voter_id"`
	Eligible      bool       `json:"eligible"`
	HasVoted      bool       `json:"has_voted"`
	Method        VoteMethod `json:"method"`
	TPSID         *int64     `json:"tps_id,omitempty"`
	LastVoteAt    *time.Time `json:"last_vote_at,omitempty"`
	OnlineAllowed bool       `json:"online_allowed"`
	TPSAllowed    bool       `json:"tps_allowed"`
}

// MeStatusRow mirrors voter_status joined with elections.
type MeStatusRow struct {
	ElectionID      int64
	VoterID         int64
	IsEligible      bool
	HasVoted        bool
	LastVoteAt      *time.Time
	LastVoteChannel *string
	LastTPSID       *int64
	OnlineEnabled   bool
	TPSEnabled      bool
}
