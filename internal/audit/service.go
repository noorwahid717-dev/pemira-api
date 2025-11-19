package audit

import (
	"context"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Log(
	ctx context.Context,
	actorID int64,
	action AuditAction,
	entityType string,
	entityID int64,
	metadata map[string]interface{},
	ipAddress, userAgent string,
) error {
	log := &AuditLog{
		ActorID:    actorID,
		Action:     string(action),
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
		CreatedAt:  time.Now(),
	}

	return s.repo.Create(ctx, log)
}

func (s *Service) GetLogs(ctx context.Context, filters map[string]interface{}) ([]*AuditLog, error) {
	logs, _, err := s.repo.List(ctx, shared.PaginationParams{Page: 1, PerPage: 100}, filters)
	return logs, err
}
