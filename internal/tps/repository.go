package tps

import "context"

type Repository interface {
	GetByID(ctx context.Context, id int64) (*TPS, error)
	List(ctx context.Context, electionID int64) ([]*TPS, error)
	Create(ctx context.Context, tps *TPS) error
	Update(ctx context.Context, tps *TPS) error
	
	CreateCheckin(ctx context.Context, checkin *TPSCheckin) error
	GetCheckin(ctx context.Context, id int64) (*TPSCheckin, error)
	ListCheckins(ctx context.Context, tpsID int64) ([]*TPSCheckin, error)
	UpdateCheckin(ctx context.Context, checkin *TPSCheckin) error
}
