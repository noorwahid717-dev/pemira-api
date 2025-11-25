package tps_test

import (
	"context"
	"testing"
	"time"

	"pemira-api/internal/tps"
)

// Mock repository for testing
type mockRepository struct {
	tpsList     []*tps.TPS
	qrList      []*tps.TPSQR
	checkinList []*tps.TPSCheckin
}

func (m *mockRepository) GetByID(ctx context.Context, id int64) (*tps.TPS, error) {
	for _, t := range m.tpsList {
		if t.ID == id {
			return t, nil
		}
	}
	return nil, tps.ErrTPSNotFound
}

func (m *mockRepository) GetByCode(ctx context.Context, code string) (*tps.TPS, error) {
	for _, t := range m.tpsList {
		if t.Code == code {
			return t, nil
		}
	}
	return nil, tps.ErrTPSNotFound
}

func (m *mockRepository) GetByIDElection(ctx context.Context, electionID, id int64) (*tps.TPS, error) {
	return m.GetByID(ctx, id)
}

func (m *mockRepository) List(ctx context.Context, filter tps.ListFilter) ([]*tps.TPS, int, error) {
	return m.tpsList, len(m.tpsList), nil
}

func (m *mockRepository) Create(ctx context.Context, tpsRow *tps.TPS) error      { return nil }
func (m *mockRepository) Update(ctx context.Context, tpsRow *tps.TPS) error      { return nil }
func (m *mockRepository) Delete(ctx context.Context, electionID, id int64) error { return nil }
func (m *mockRepository) GetStats(ctx context.Context, tpsID int64) (*tps.TPSStats, error) {
	return &tps.TPSStats{}, nil
}
func (m *mockRepository) CreateQR(ctx context.Context, qr *tps.TPSQR) error { return nil }
func (m *mockRepository) GetActiveQR(ctx context.Context, tpsID int64) (*tps.TPSQR, error) {
	for _, qr := range m.qrList {
		if qr.TPSID == tpsID && qr.IsActive {
			return qr, nil
		}
	}
	return nil, tps.ErrQRInvalid
}
func (m *mockRepository) GetQRBySecret(ctx context.Context, tpsCode, secret string) (*tps.TPSQR, error) {
	for _, qr := range m.qrList {
		if qr.QRToken == secret {
			return qr, nil
		}
	}
	return nil, tps.ErrQRInvalid
}
func (m *mockRepository) RevokeQR(ctx context.Context, qrID int64) error { return nil }
func (m *mockRepository) GetQRMetadata(ctx context.Context, tpsID int64) (*tps.QRInfo, error) {
	return &tps.QRInfo{ID: 1, QRToken: "token", IsActive: true, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}
func (m *mockRepository) RotateQR(ctx context.Context, tpsID int64) (*tps.QRInfo, error) {
	return &tps.QRInfo{ID: 2, QRToken: "token2", IsActive: true, CreatedAt: time.Now().Format(time.RFC3339)}, nil
}
func (m *mockRepository) GetQRPrintPayload(ctx context.Context, tpsID int64) (string, error) {
	return "PEMIRA|TPS01|token", nil
}
func (m *mockRepository) AssignPanitia(ctx context.Context, tpsID int64, members []tps.TPSPanitia) error {
	return nil
}
func (m *mockRepository) GetPanitia(ctx context.Context, tpsID int64) ([]*tps.TPSPanitia, error) {
	return nil, nil
}
func (m *mockRepository) IsPanitiaAssigned(ctx context.Context, tpsID, userID int64) (bool, error) {
	return true, nil
}
func (m *mockRepository) ClearPanitia(ctx context.Context, tpsID int64) error { return nil }
func (m *mockRepository) CreateCheckin(ctx context.Context, checkin *tps.TPSCheckin) error {
	checkin.ID = int64(len(m.checkinList) + 1)
	m.checkinList = append(m.checkinList, checkin)
	return nil
}
func (m *mockRepository) GetCheckin(ctx context.Context, id int64) (*tps.TPSCheckin, error) {
	for _, c := range m.checkinList {
		if c.ID == id {
			return c, nil
		}
	}
	return nil, tps.ErrCheckinNotFound
}
func (m *mockRepository) GetCheckinByVoter(ctx context.Context, voterID, electionID int64) (*tps.TPSCheckin, error) {
	for _, c := range m.checkinList {
		if c.VoterID == voterID && c.ElectionID == electionID {
			return c, nil
		}
	}
	return nil, nil
}
func (m *mockRepository) ListCheckins(ctx context.Context, tpsID int64, status string, page, limit int) ([]*tps.TPSCheckin, error) {
	return m.checkinList, nil
}
func (m *mockRepository) UpdateCheckin(ctx context.Context, checkin *tps.TPSCheckin) error {
	for i, c := range m.checkinList {
		if c.ID == checkin.ID {
			m.checkinList[i] = checkin
		}
	}
	return nil
}
func (m *mockRepository) CountCheckins(ctx context.Context, tpsID int64, status string) (int, error) {
	return len(m.checkinList), nil
}
func (m *mockRepository) IsVoterEligible(ctx context.Context, voterID, electionID int64) (bool, error) {
	return true, nil
}
func (m *mockRepository) HasVoterVoted(ctx context.Context, voterID, electionID int64) (bool, error) {
	return false, nil
}
func (m *mockRepository) GetVoterInfo(ctx context.Context, voterID int64) (*tps.VoterInfo, error) {
	return &tps.VoterInfo{ID: voterID, Name: "Test", NIM: "123"}, nil
}
func (m *mockRepository) PanelDashboardStats(ctx context.Context, tpsID, electionID int64) (*tps.PanelDashboardStatsRow, error) {
	return &tps.PanelDashboardStatsRow{}, nil
}
func (m *mockRepository) PanelListCheckins(ctx context.Context, tpsID int64, status, search string, limit, offset int) ([]tps.PanelCheckinRow, int, error) {
	return []tps.PanelCheckinRow{}, 0, nil
}
func (m *mockRepository) PanelGetCheckin(ctx context.Context, checkinID int64) (*tps.PanelCheckinRow, error) {
	return nil, tps.ErrCheckinNotFound
}
func (m *mockRepository) PanelTimeline(ctx context.Context, tpsID int64) ([]tps.PanelTimelineRow, error) {
	return []tps.PanelTimelineRow{}, nil
}
func (m *mockRepository) GetOperatorInfo(ctx context.Context, userID int64) (*tps.OperatorInfo, error) {
	return &tps.OperatorInfo{ID: userID}, nil
}
func (m *mockRepository) ParseRegistrationCode(ctx context.Context, raw string) (*tps.PanelRegistrationCode, error) {
	return &tps.PanelRegistrationCode{ElectionID: 1, VoterID: 1, TPSID: ptrInt64(1)}, nil
}
func (m *mockRepository) CreatePanelCheckin(ctx context.Context, reg tps.PanelRegistrationCode) (*tps.PanelCheckinRow, error) {
	return &tps.PanelCheckinRow{ID: 1, TPSID: *reg.TPSID, ElectionID: reg.ElectionID, VoterID: reg.VoterID, Status: tps.CheckinStatusApproved, ScanAt: time.Now()}, nil
}
func (m *mockRepository) ListOperators(ctx context.Context, tpsID int64) ([]tps.OperatorInfo, error) {
	return []tps.OperatorInfo{}, nil
}
func (m *mockRepository) CreateOperator(ctx context.Context, tpsID int64, op tps.OperatorCreate) (*tps.OperatorInfo, error) {
	return &tps.OperatorInfo{ID: 1, Username: op.Username, Name: op.Name, Email: op.Email, TPSID: &tpsID}, nil
}
func (m *mockRepository) DeleteOperator(ctx context.Context, tpsID, userID int64) error { return nil }

func ptrInt64(v int64) *int64 { return &v }

// ... implement other Repository methods as needed for tests

func TestService_ScanQR_ValidQR(t *testing.T) {
	// Setup
	votingDate := time.Now().Add(24 * time.Hour)
	mockRepo := &mockRepository{
		tpsList: []*tps.TPS{
			{
				ID:         1,
				ElectionID: 1,
				Code:       "TPS01",
				Name:       "TPS 1",
				Status:     tps.StatusActive,
				VotingDate: &votingDate,
			},
		},
		qrList: []*tps.TPSQR{
			{
				ID:       1,
				TPSID:    1,
				QRToken:  "abc123",
				IsActive: true,
			},
		},
	}

	service := tps.NewService(mockRepo)
	ctx := context.Background()

	// Test
	req := &tps.ScanQRRequest{
		QRPayload: "PEMIRA|TPS01|abc123",
	}

	result, err := service.ScanQR(ctx, 100, req)

	// Assert
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Status != tps.CheckinStatusPending {
		t.Errorf("Expected status PENDING, got: %s", result.Status)
	}

	if result.TPS.Code != "TPS01" {
		t.Errorf("Expected TPS code TPS01, got: %s", result.TPS.Code)
	}
}

func TestService_ScanQR_InvalidQRFormat(t *testing.T) {
	mockRepo := &mockRepository{}
	service := tps.NewService(mockRepo)
	ctx := context.Background()

	req := &tps.ScanQRRequest{
		QRPayload: "INVALID|FORMAT",
	}

	_, err := service.ScanQR(ctx, 100, req)

	if err != tps.ErrQRInvalid {
		t.Errorf("Expected ErrQRInvalid, got: %v", err)
	}
}

func TestService_ApproveCheckin_Success(t *testing.T) {
	mockRepo := &mockRepository{
		checkinList: []*tps.TPSCheckin{
			{
				ID:         1,
				TPSID:      1,
				VoterID:    100,
				ElectionID: 1,
				Status:     tps.CheckinStatusPending,
			},
		},
	}

	service := tps.NewService(mockRepo)
	ctx := context.Background()

	result, err := service.ApproveCheckin(ctx, 1, 1, 200)

	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}

	if result.Status != tps.CheckinStatusApproved {
		t.Errorf("Expected status APPROVED, got: %s", result.Status)
	}

	if result.ApprovedAt.IsZero() {
		t.Error("Expected ApprovedAt to be set")
	}
}

func TestService_ApproveCheckin_NotPending(t *testing.T) {
	mockRepo := &mockRepository{
		checkinList: []*tps.TPSCheckin{
			{
				ID:         1,
				TPSID:      1,
				VoterID:    100,
				ElectionID: 1,
				Status:     tps.CheckinStatusApproved, // Already approved
			},
		},
	}

	service := tps.NewService(mockRepo)
	ctx := context.Background()

	_, err := service.ApproveCheckin(ctx, 1, 1, 200)

	if err != tps.ErrCheckinNotPending {
		t.Errorf("Expected ErrCheckinNotPending, got: %v", err)
	}
}

func TestService_GenerateQRSecret(t *testing.T) {
	service := tps.NewService(nil)

	secret := service.GenerateQRSecret()

	if len(secret) != 12 {
		t.Errorf("Expected secret length 12, got: %d", len(secret))
	}

	// Generate multiple secrets and ensure they're unique
	secrets := make(map[string]bool)
	for i := 0; i < 100; i++ {
		s := service.GenerateQRSecret()
		if secrets[s] {
			t.Error("Generated duplicate secret")
		}
		secrets[s] = true
	}
}

// Example benchmark test
func BenchmarkService_GenerateQRSecret(b *testing.B) {
	service := tps.NewService(nil)

	for i := 0; i < b.N; i++ {
		service.GenerateQRSecret()
	}
}

// Example table-driven test
func TestService_ValidateQRPayload(t *testing.T) {
	tests := []struct {
		name      string
		payload   string
		wantError error
	}{
		{
			name:      "Valid QR",
			payload:   "PEMIRA|TPS01|abc123",
			wantError: nil,
		},
		{
			name:      "Invalid prefix",
			payload:   "INVALID|TPS01|abc123",
			wantError: tps.ErrQRInvalid,
		},
		{
			name:      "Missing parts",
			payload:   "PEMIRA|TPS01",
			wantError: tps.ErrQRInvalid,
		},
		{
			name:      "Too many parts",
			payload:   "PEMIRA|TPS01|abc123|extra",
			wantError: tps.ErrQRInvalid,
		},
		{
			name:      "Empty payload",
			payload:   "",
			wantError: tps.ErrQRInvalid,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test logic here
			// This is a simplified example - actual implementation would test ScanQR
		})
	}
}
