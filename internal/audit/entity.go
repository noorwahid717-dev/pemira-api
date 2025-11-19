package audit

import "time"

type AuditLog struct {
	ID         int64                  `json:"id"`
	ActorID    int64                  `json:"actor_id"`
	Action     string                 `json:"action"`
	EntityType string                 `json:"entity_type"`
	EntityID   int64                  `json:"entity_id"`
	Metadata   map[string]interface{} `json:"metadata"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	CreatedAt  time.Time              `json:"created_at"`
}

type AuditAction string

const (
	ActionVoteCast          AuditAction = "VOTE_CAST"
	ActionElectionCreated   AuditAction = "ELECTION_CREATED"
	ActionElectionUpdated   AuditAction = "ELECTION_UPDATED"
	ActionCandidateCreated  AuditAction = "CANDIDATE_CREATED"
	ActionCandidateUpdated  AuditAction = "CANDIDATE_UPDATED"
	ActionTPSCreated        AuditAction = "TPS_CREATED"
	ActionTPSQRRegenerated  AuditAction = "TPS_QR_REGENERATED"
	ActionDPTImported       AuditAction = "DPT_IMPORTED"
	ActionVoterStatusReset  AuditAction = "VOTER_STATUS_RESET"
	ActionCheckinApproved   AuditAction = "CHECKIN_APPROVED"
	ActionCheckinRejected   AuditAction = "CHECKIN_REJECTED"
)
