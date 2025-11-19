package announcement

import (
	"context"
	"pemira-api/internal/shared"
)

type Repository interface {
	GetByID(ctx context.Context, id int64) (*Announcement, error)
	List(ctx context.Context, params shared.PaginationParams, electionID int64, onlyPublished bool) ([]*Announcement, int64, error)
	Create(ctx context.Context, announcement *Announcement) error
	Update(ctx context.Context, announcement *Announcement) error
	Delete(ctx context.Context, id int64) error
}
