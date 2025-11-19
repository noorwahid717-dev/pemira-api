package audit

import (
	"context"
	"pemira-api/internal/shared"
)

type Repository interface {
	Create(ctx context.Context, log *AuditLog) error
	List(ctx context.Context, params shared.PaginationParams, filters map[string]interface{}) ([]*AuditLog, int64, error)
	GetByID(ctx context.Context, id int64) (*AuditLog, error)
}
