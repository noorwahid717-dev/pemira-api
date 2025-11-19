package announcement

import (
	"context"
	"time"
	
	"pemira-api/internal/shared"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int64) (*Announcement, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) ListPublished(ctx context.Context, electionID int64, params shared.PaginationParams) ([]*Announcement, int64, error) {
	return s.repo.List(ctx, params, electionID, true)
}

func (s *Service) ListAll(ctx context.Context, electionID int64, params shared.PaginationParams) ([]*Announcement, int64, error) {
	return s.repo.List(ctx, params, electionID, false)
}

func (s *Service) Create(ctx context.Context, announcement *Announcement) error {
	announcement.CreatedAt = time.Now()
	announcement.UpdatedAt = time.Now()
	
	if announcement.IsPublished && announcement.PublishedAt == nil {
		now := time.Now()
		announcement.PublishedAt = &now
	}
	
	return s.repo.Create(ctx, announcement)
}

func (s *Service) Update(ctx context.Context, announcement *Announcement) error {
	announcement.UpdatedAt = time.Now()
	return s.repo.Update(ctx, announcement)
}

func (s *Service) Publish(ctx context.Context, id int64) error {
	announcement, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	
	now := time.Now()
	announcement.IsPublished = true
	announcement.PublishedAt = &now
	announcement.UpdatedAt = now
	
	return s.repo.Update(ctx, announcement)
}
