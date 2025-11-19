package voting

import (
	"context"
	"time"
)

// AuditEntry represents an audit log entry
type AuditEntry struct {
	ActorVoterID *int64         `json:"actor_voter_id"`
	ActorUserID  *int64         `json:"actor_user_id"`
	Action       string         `json:"action"`
	EntityType   string         `json:"entity_type"`
	EntityID     int64          `json:"entity_id"`
	Metadata     map[string]any `json:"metadata"`
	CreatedAt    time.Time      `json:"created_at"`
}

// AuditService handles audit logging
type AuditService interface {
	Log(ctx context.Context, entry AuditEntry) error
}

// Simple audit service implementation
type auditService struct{}

func NewAuditService() AuditService {
	return &auditService{}
}

func (s *auditService) Log(ctx context.Context, entry AuditEntry) error {
	// For now, this is a no-op or could log to stdout
	// In production, this would write to audit_logs table or send to queue
	return nil
}
