package tps

import "context"

type Repository interface {
	// TPS Management
	GetByID(ctx context.Context, id int64) (*TPS, error)
	GetByIDElection(ctx context.Context, electionID, id int64) (*TPS, error)
	GetByCode(ctx context.Context, code string) (*TPS, error)
	List(ctx context.Context, filter ListFilter) ([]*TPS, int, error)
	Create(ctx context.Context, tps *TPS) error
	Update(ctx context.Context, tps *TPS) error
	Delete(ctx context.Context, electionID, id int64) error
	GetStats(ctx context.Context, tpsID int64) (*TPSStats, error)

	// QR Management
	CreateQR(ctx context.Context, qr *TPSQR) error
	GetActiveQR(ctx context.Context, tpsID int64) (*TPSQR, error)
	GetQRBySecret(ctx context.Context, tpsCode, secret string) (*TPSQR, error)
	RevokeQR(ctx context.Context, qrID int64) error
	GetQRMetadata(ctx context.Context, tpsID int64) (*QRInfo, error)
	RotateQR(ctx context.Context, tpsID int64) (*QRInfo, error)
	GetQRPrintPayload(ctx context.Context, tpsID int64) (string, error)

	// Panitia Management
	AssignPanitia(ctx context.Context, tpsID int64, members []TPSPanitia) error
	GetPanitia(ctx context.Context, tpsID int64) ([]*TPSPanitia, error)
	IsPanitiaAssigned(ctx context.Context, tpsID, userID int64) (bool, error)
	ClearPanitia(ctx context.Context, tpsID int64) error
	ListOperators(ctx context.Context, tpsID int64) ([]OperatorInfo, error)
	CreateOperator(ctx context.Context, tpsID int64, op OperatorCreate) (*OperatorInfo, error)
	DeleteOperator(ctx context.Context, tpsID, userID int64) error

	// Check-in Management
	CreateCheckin(ctx context.Context, checkin *TPSCheckin) error
	GetCheckin(ctx context.Context, id int64) (*TPSCheckin, error)
	GetCheckinByVoter(ctx context.Context, voterID, electionID int64) (*TPSCheckin, error)
	ListCheckins(ctx context.Context, tpsID int64, status string, page, limit int) ([]*TPSCheckin, error)
	UpdateCheckin(ctx context.Context, checkin *TPSCheckin) error
	CountCheckins(ctx context.Context, tpsID int64, status string) (int, error)

	// Voter validation
	IsVoterEligible(ctx context.Context, voterID, electionID int64) (bool, error)
	HasVoterVoted(ctx context.Context, voterID, electionID int64) (bool, error)
	GetVoterInfo(ctx context.Context, voterID int64) (*VoterInfo, error)

	// Panel helpers
	PanelDashboardStats(ctx context.Context, tpsID, electionID int64) (*PanelDashboardStatsRow, error)
	PanelListCheckins(ctx context.Context, tpsID int64, status, search string, limit, offset int) ([]PanelCheckinRow, int, error)
	PanelGetCheckin(ctx context.Context, checkinID int64) (*PanelCheckinRow, error)
	PanelTimeline(ctx context.Context, tpsID int64) ([]PanelTimelineRow, error)
	PanelListTPSByElection(ctx context.Context, electionID int64) ([]PanelTPSListItem, error)
	GetOperatorInfo(ctx context.Context, userID int64) (*OperatorInfo, error)
	ParseRegistrationCode(ctx context.Context, raw string) (*PanelRegistrationCode, error)
	CreatePanelCheckin(ctx context.Context, reg PanelRegistrationCode) (*PanelCheckinRow, error)
}

type ListFilter struct {
	Status     string
	ElectionID int64
	Page       int
	Limit      int
}
