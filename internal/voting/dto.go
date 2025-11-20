package voting

import "time"

// Request DTOs
type CastVoteRequest struct {
	CandidateID int64 `json:"candidate_id" validate:"required,min=1"`
}

type CastOnlineVoteRequest struct {
	ElectionID  int64 `json:"election_id"`
	CandidateID int64 `json:"candidate_id"`
}

type CastTPSVoteRequest struct {
	ElectionID  int64 `json:"election_id"`
	CandidateID int64 `json:"candidate_id"`
	TPSID       int64 `json:"tps_id"`
}

// Response DTOs
type VotingConfigResponse struct {
	Election ElectionInfo `json:"election"`
	Voter    VoterInfo    `json:"voter"`
	Mode     VotingMode   `json:"mode"`
}

type ElectionInfo struct {
	ID             int64     `json:"id"`
	Code           string    `json:"code"`
	Name           string    `json:"name"`
	Status         string    `json:"status"`
	VotingStartAt  time.Time `json:"voting_start_at"`
	VotingEndAt    time.Time `json:"voting_end_at"`
}

type VoterInfo struct {
	ID            int64      `json:"id"`
	NIM           string     `json:"nim"`
	Name          string     `json:"name"`
	IsEligible    bool       `json:"is_eligible"`
	HasVoted      bool       `json:"has_voted"`
	VotingMethod  *string    `json:"voting_method"`
	VotedAt       *time.Time `json:"voted_at"`
}

type VotingMode struct {
	OnlineEnabled bool `json:"online_enabled"`
	TPSEnabled    bool `json:"tps_enabled"`
}

type VoteReceipt struct {
	ElectionID int64         `json:"election_id"`
	VoterID    int64         `json:"voter_id"`
	Method     string        `json:"method"`
	VotedAt    time.Time     `json:"voted_at"`
	Receipt    ReceiptDetail `json:"receipt"`
	TPS        *TPSInfo      `json:"tps,omitempty"`
}

type ReceiptDetail struct {
	TokenHash string `json:"token_hash"`
	Note      string `json:"note"`
}

type TPSInfo struct {
	ID   int64  `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type TPSVotingStatus struct {
	Eligible  bool       `json:"eligible"`
	TPS       *TPSInfo   `json:"tps,omitempty"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
	Reason    *string    `json:"reason,omitempty"`
}

type ReceiptResponse struct {
	HasVoted     bool       `json:"has_voted"`
	ElectionID   *int64     `json:"election_id,omitempty"`
	Method       *string    `json:"method,omitempty"`
	TPS          *TPSInfo   `json:"tps,omitempty"`
	VotedAt      *time.Time `json:"voted_at,omitempty"`
	Receipt      *ReceiptDetail `json:"receipt,omitempty"`
}

type LiveCountResponse struct {
	ElectionID int64            `json:"election_id"`
	Counts     map[int64]int64  `json:"counts"`
}
