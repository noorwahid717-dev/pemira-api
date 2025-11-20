package tps

import "context"

type AdminRepository interface {
	// TPS CRUD
	List(ctx context.Context) ([]TPSDTO, error)
	GetByID(ctx context.Context, id int64) (*TPSDTO, error)
	Create(ctx context.Context, req TPSCreateRequest) (*TPSDTO, error)
	Update(ctx context.Context, id int64, req TPSUpdateRequest) (*TPSDTO, error)
	Delete(ctx context.Context, id int64) error

	// QR management
	GetTPSQRMetadata(ctx context.Context, tpsID int64) (*TPSQRMetadataResponse, error)
	RotateTPSQR(ctx context.Context, tpsID int64) (*TPSQRRotateResponse, error)
	GetTPSQRForPrint(ctx context.Context, tpsID int64) (*TPSQRPrintResponse, error)

	// Operator management
	ListOperators(ctx context.Context, tpsID int64) ([]TPSOperatorDTO, error)
	CreateOperator(ctx context.Context, tpsID int64, username, password, name, email string) (*TPSOperatorDTO, error)
	RemoveOperator(ctx context.Context, tpsID, userID int64) error

	// Monitoring
	ListMonitorForElection(ctx context.Context, electionID int64) ([]TPSMonitorDTO, error)
}
