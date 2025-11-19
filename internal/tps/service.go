package tps

import (
	"context"
	"crypto/rand"
	"encoding/hex"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id int64) (*TPS, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, electionID int64) ([]*TPS, error) {
	return s.repo.List(ctx, electionID)
}

func (s *Service) GenerateQRCode() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *Service) CreateCheckin(ctx context.Context, checkin *TPSCheckin) error {
	return s.repo.CreateCheckin(ctx, checkin)
}

func (s *Service) ApproveCheckin(ctx context.Context, checkinID, approverID int64) error {
	checkin, err := s.repo.GetCheckin(ctx, checkinID)
	if err != nil {
		return err
	}
	
	checkin.Status = "APPROVED"
	checkin.ApprovedBy = &approverID
	
	return s.repo.UpdateCheckin(ctx, checkin)
}

func (s *Service) ListCheckins(ctx context.Context, tpsID int64) ([]*TPSCheckin, error) {
	return s.repo.ListCheckins(ctx, tpsID)
}
