package voting

import "time"

type Vote struct {
	ID            int64     `json:"id"`
	ElectionID    int64     `json:"election_id"`
	CandidateID   int64     `json:"candidate_id"`
	TokenHash     string    `json:"token_hash"`
	Channel       string    `json:"channel"` // "ONLINE" | "TPS"
	TPSID         *int64    `json:"tps_id"`
	CandidateQRID *int64    `json:"candidate_qr_id,omitempty"`
	BallotScanID  *int64    `json:"ballot_scan_id,omitempty"`
	CastAt        time.Time `json:"cast_at"`
}

type VoteToken struct {
	ID         int64      `json:"id"`
	ElectionID int64      `json:"election_id"`
	VoterID    int64      `json:"voter_id"`
	TokenHash  string     `json:"token_hash"`
	IssuedAt   time.Time  `json:"issued_at"`
	UsedAt     *time.Time `json:"used_at"`
	Method     string     `json:"method"` // "ONLINE" | "TPS"
	TPSID      *int64     `json:"tps_id"`
}

type VoterStatusEntity struct {
	ID              int64      `json:"id"`
	ElectionID      int64      `json:"election_id"`
	VoterID         int64      `json:"voter_id"`
	IsEligible      bool       `json:"is_eligible"`
	HasVoted        bool       `json:"has_voted"`
	VotingMethod    *string    `json:"voting_method"` // "ONLINE" | "TPS"
	TPSID           *int64     `json:"tps_id"`
	VotedAt         *time.Time `json:"voted_at"`
	TokenHash       *string    `json:"token_hash"`
	PreferredMethod *string    `json:"preferred_method,omitempty"`
	OnlineAllowed   bool       `json:"online_allowed"`
	TPSAllowed      bool       `json:"tps_allowed"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type VoteResultEntity struct {
	ElectionID int64         `json:"election_id"`
	VoterID    int64         `json:"voter_id"`
	Method     string        `json:"method"`
	VotedAt    time.Time     `json:"voted_at"`
	TPS        *TPSInfo      `json:"tps,omitempty"`
	Receipt    ReceiptDetail `json:"receipt"`
}

type CandidateQR struct {
	ID          int64  `json:"id"`
	ElectionID  int64  `json:"election_id"`
	CandidateID int64  `json:"candidate_id"`
	Version     int    `json:"version"`
	QRToken     string `json:"qr_token"`
	IsActive    bool   `json:"is_active"`
}

type BallotScan struct {
	ID              int64     `json:"id"`
	ElectionID      int64     `json:"election_id"`
	TPSID           int64     `json:"tps_id"`
	CheckinID       int64     `json:"checkin_id"`
	VoterID         int64     `json:"voter_id"`
	CandidateID     *int64    `json:"candidate_id,omitempty"`
	CandidateQRID   *int64    `json:"candidate_qr_id,omitempty"`
	RawPayload      string    `json:"raw_payload"`
	PayloadValid    bool      `json:"payload_valid"`
	Status          string    `json:"status"`
	RejectedReason  *string   `json:"rejected_reason,omitempty"`
	ScannedByUserID int64     `json:"scanned_by_user_id"`
	ScannedAt       time.Time `json:"scanned_at"`
}

type VoterTPSQR struct {
	ID         int64      `json:"id"`
	VoterID    int64      `json:"voter_id"`
	ElectionID int64      `json:"election_id"`
	QRToken    string     `json:"qr_token"`
	IsActive   bool       `json:"is_active"`
	RotatedAt  *time.Time `json:"rotated_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
