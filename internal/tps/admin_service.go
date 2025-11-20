package tps

import "context"

type AdminService struct {
	repo AdminRepository
}

func NewAdminService(repo AdminRepository) *AdminService {
	return &AdminService{repo: repo}
}

// CRUD TPS
func (s *AdminService) List(ctx context.Context) ([]TPSDTO, error) {
	return s.repo.List(ctx)
}

func (s *AdminService) Get(ctx context.Context, id int64) (*TPSDTO, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *AdminService) Create(ctx context.Context, req TPSCreateRequest) (*TPSDTO, error) {
	return s.repo.Create(ctx, req)
}

func (s *AdminService) Update(ctx context.Context, id int64, req TPSUpdateRequest) (*TPSDTO, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *AdminService) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// Operators
func (s *AdminService) ListOperators(ctx context.Context, tpsID int64) ([]TPSOperatorDTO, error) {
	return s.repo.ListOperators(ctx, tpsID)
}

func (s *AdminService) CreateOperator(
	ctx context.Context,
	tpsID int64,
	username, password, name, email string,
) (*TPSOperatorDTO, error) {
	return s.repo.CreateOperator(ctx, tpsID, username, password, name, email)
}

func (s *AdminService) RemoveOperator(ctx context.Context, tpsID, userID int64) error {
	return s.repo.RemoveOperator(ctx, tpsID, userID)
}

// Monitoring
func (s *AdminService) Monitor(ctx context.Context, electionID int64) ([]TPSMonitorDTO, error) {
	return s.repo.ListMonitorForElection(ctx, electionID)
}
