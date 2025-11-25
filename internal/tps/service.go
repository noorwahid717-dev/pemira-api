package tps

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) ensureTPSElection(ctx context.Context, electionID, tpsID int64) (*TPS, error) {
	if electionID > 0 {
		return s.repo.GetByIDElection(ctx, electionID, tpsID)
	}
	return s.repo.GetByID(ctx, tpsID)
}

// Admin TPS Management
func (s *Service) GetByID(ctx context.Context, id int64) (*TPSDetailResponse, error) {
	tps, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, ErrTPSNotFound
	}
	return s.buildDetail(ctx, tps)
}

func (s *Service) GetByIDElection(ctx context.Context, electionID, id int64) (*TPSDetailResponse, error) {
	tps, err := s.repo.GetByIDElection(ctx, electionID, id)
	if err != nil {
		return nil, ErrTPSNotFound
	}
	return s.buildDetail(ctx, tps)
}

func (s *Service) buildDetail(ctx context.Context, tps *TPS) (*TPSDetailResponse, error) {
	stats, _ := s.repo.GetStats(ctx, tps.ID)
	if stats == nil {
		stats = &TPSStats{}
	}

	qr, _ := s.repo.GetActiveQR(ctx, tps.ID)
	panitia, _ := s.repo.GetPanitia(ctx, tps.ID)

	response := &TPSDetailResponse{
		ID:               tps.ID,
		ElectionID:       tps.ElectionID,
		Code:             tps.Code,
		Name:             tps.Name,
		Location:         tps.Location,
		Status:           tps.Status,
		OpenTime:         tps.OpenTime,
		CloseTime:        tps.CloseTime,
		CapacityEstimate: tps.CapacityEstimate,
		PICName:          tps.PICName,
		PICPhone:         tps.PICPhone,
		Notes:            tps.Notes,
		AreaFaculty:      &FacultyInfo{ID: tps.AreaFacultyID},
		Stats:            *stats,
		Panitia:          []PanitiaInfo{},
	}

	if tps.VotingDate != nil {
		response.VotingDate = tps.VotingDate.Format("2006-01-02")
	}

	if qr != nil {
		response.QR = &QRInfo{
			ID:        qr.ID,
			QRToken:   qr.QRToken,
			IsActive:  qr.IsActive,
			CreatedAt: qr.CreatedAt.Format(time.RFC3339),
		}
	}

	for _, p := range panitia {
		response.Panitia = append(response.Panitia, PanitiaInfo{
			UserID: p.UserID,
			Role:   p.Role,
		})
	}

	return response, nil
}

func (s *Service) List(ctx context.Context, filter ListFilter) (*TPSListResponse, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Limit < 1 {
		filter.Limit = 20
	}

	tpsList, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, err
	}

	items := make([]TPSListItem, 0, len(tpsList))
	for _, t := range tpsList {
		stats, _ := s.repo.GetStats(ctx, t.ID)
		activeQR, _ := s.repo.GetActiveQR(ctx, t.ID)

		item := TPSListItem{
			ID:          t.ID,
			Code:        t.Code,
			Name:        t.Name,
			Location:    t.Location,
			Status:      t.Status,
			OpenTime:    t.OpenTime,
			CloseTime:   t.CloseTime,
			PICName:     t.PICName,
			PICPhone:    t.PICPhone,
			HasActiveQR: activeQR != nil,
		}

		if t.VotingDate != nil {
			item.VotingDate = t.VotingDate.Format("2006-01-02")
		}

		if stats != nil {
			item.TotalVotes = stats.TotalVotes
			item.TotalCheckins = stats.TotalCheckins
		}

		items = append(items, item)
	}

	totalPages := (total + filter.Limit - 1) / filter.Limit

	return &TPSListResponse{
		Items: items,
		Pagination: &PaginationInfo{
			Page:       filter.Page,
			Limit:      filter.Limit,
			TotalItems: total,
			TotalPages: totalPages,
		},
	}, nil
}

func (s *Service) Create(ctx context.Context, req *CreateTPSRequest) (int64, error) {
	votingDate, err := time.Parse("2006-01-02", req.VotingDate)
	if err != nil {
		return 0, ErrInvalidTimeFormat
	}

	tps := &TPS{
		ElectionID:       req.ElectionID,
		Code:             req.Code,
		Name:             req.Name,
		Location:         req.Location,
		Status:           req.Status,
		VotingDate:       &votingDate,
		OpenTime:         req.OpenTime,
		CloseTime:        req.CloseTime,
		CapacityEstimate: req.CapacityEstimate,
	}

	if err := s.repo.Create(ctx, tps); err != nil {
		if strings.Contains(err.Error(), "duplicate") {
			return 0, ErrTPSCodeDuplicate
		}
		return 0, err
	}

	// Auto-generate QR
	if tps.ID > 0 {
		qrSecret := s.GenerateQRSecret()
		qr := &TPSQR{
			TPSID:    tps.ID,
			QRToken:  qrSecret,
			IsActive: true,
		}
		_ = s.repo.CreateQR(ctx, qr)
	}

	return tps.ID, nil
}

func (s *Service) Update(ctx context.Context, id int64, req *UpdateTPSRequest) error {
	tps, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return ErrTPSNotFound
	}

	votingDate, err := time.Parse("2006-01-02", req.VotingDate)
	if err != nil {
		return ErrInvalidTimeFormat
	}

	tps.Name = req.Name
	tps.Location = req.Location
	tps.Status = req.Status
	tps.VotingDate = &votingDate
	tps.OpenTime = req.OpenTime
	tps.CloseTime = req.CloseTime
	tps.CapacityEstimate = req.CapacityEstimate
	tps.PICName = req.PICName
	tps.PICPhone = req.PICPhone
	tps.Notes = req.Notes

	return s.repo.Update(ctx, tps)
}

func (s *Service) UpdateWithElection(ctx context.Context, electionID, id int64, req *UpdateTPSRequest) error {
	tpsRow, err := s.repo.GetByIDElection(ctx, electionID, id)
	if err != nil {
		return ErrTPSNotFound
	}
	if err := s.Update(ctx, tpsRow.ID, req); err != nil {
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, electionID, id int64) error {
	if _, err := s.repo.GetByIDElection(ctx, electionID, id); err != nil {
		return ErrTPSNotFound
	}
	return s.repo.Delete(ctx, electionID, id)
}

func (s *Service) GetQRMetadata(ctx context.Context, electionID, tpsID int64) (*QRInfo, error) {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return nil, err
	}
	return s.repo.GetQRMetadata(ctx, tpsID)
}

func (s *Service) RotateQR(ctx context.Context, electionID, tpsID int64) (*QRInfo, error) {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return nil, err
	}
	return s.repo.RotateQR(ctx, tpsID)
}

func (s *Service) GetQRPrintPayload(ctx context.Context, electionID, tpsID int64) (string, error) {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return "", err
	}
	return s.repo.GetQRPrintPayload(ctx, tpsID)
}

func (s *Service) ListOperators(ctx context.Context, electionID, tpsID int64) ([]OperatorInfo, error) {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return nil, err
	}
	return s.repo.ListOperators(ctx, tpsID)
}

func (s *Service) CreateOperator(ctx context.Context, electionID, tpsID int64, op OperatorCreate) (*OperatorInfo, error) {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return nil, err
	}
	return s.repo.CreateOperator(ctx, tpsID, op)
}

func (s *Service) DeleteOperator(ctx context.Context, electionID, tpsID, userID int64) error {
	if _, err := s.repo.GetByIDElection(ctx, electionID, tpsID); err != nil {
		return err
	}
	return s.repo.DeleteOperator(ctx, tpsID, userID)
}

func (s *Service) AssignPanitia(ctx context.Context, tpsID int64, req *AssignPanitiaRequest) error {
	if _, err := s.repo.GetByID(ctx, tpsID); err != nil {
		return ErrTPSNotFound
	}

	if err := s.repo.ClearPanitia(ctx, tpsID); err != nil {
		return err
	}

	members := make([]TPSPanitia, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, TPSPanitia{
			TPSID:  tpsID,
			UserID: m.UserID,
			Role:   m.Role,
		})
	}

	return s.repo.AssignPanitia(ctx, tpsID, members)
}

func (s *Service) RegenerateQR(ctx context.Context, tpsID int64) (*RegenerateQRResponse, error) {
	tps, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, ErrTPSNotFound
	}

	// Revoke old QR
	oldQR, _ := s.repo.GetActiveQR(ctx, tpsID)
	if oldQR != nil {
		_ = s.repo.RevokeQR(ctx, oldQR.ID)
	}

	// Generate new QR
	qrSecret := s.GenerateQRSecret()
	newQR := &TPSQR{
		TPSID:    tpsID,
		QRToken:  qrSecret,
		IsActive: true,
	}

	if err := s.repo.CreateQR(ctx, newQR); err != nil {
		return nil, err
	}

	payload := fmt.Sprintf("PEMIRA|%s|%s", tps.Code, qrSecret)

	return &RegenerateQRResponse{
		TPSID: tpsID,
		QR: QRPayload{
			ID:        newQR.ID,
			Payload:   payload,
			CreatedAt: newQR.CreatedAt.Format(time.RFC3339),
		},
	}, nil
}

// Student Check-in
func (s *Service) ScanQR(ctx context.Context, voterID int64, req *ScanQRRequest) (*ScanQRResponse, error) {
	// Parse QR payload: PEMIRA|TPS01|c9423e5f97d4
	parts := strings.Split(req.QRPayload, "|")
	if len(parts) != 3 || parts[0] != "PEMIRA" {
		return nil, ErrQRInvalid
	}

	tpsCode := parts[1]
	qrSecret := parts[2]

	// Validate QR
	qr, err := s.repo.GetQRBySecret(ctx, tpsCode, qrSecret)
	if err != nil {
		return nil, ErrQRInvalid
	}

	if !qr.IsActive {
		return nil, ErrQRRevoked
	}

	// Validate TPS
	tps, err := s.repo.GetByCode(ctx, tpsCode)
	if err != nil {
		return nil, ErrTPSNotFound
	}

	if tps.Status != StatusActive {
		return nil, ErrTPSInactive
	}

	// Check voter eligibility
	eligible, err := s.repo.IsVoterEligible(ctx, voterID, tps.ElectionID)
	if err != nil || !eligible {
		return nil, ErrNotEligible
	}

	// Check if already voted
	hasVoted, err := s.repo.HasVoterVoted(ctx, voterID, tps.ElectionID)
	if err == nil && hasVoted {
		return nil, ErrAlreadyVoted
	}

	// Check existing pending checkin
	existingCheckin, _ := s.repo.GetCheckinByVoter(ctx, voterID, tps.ElectionID)
	if existingCheckin != nil && existingCheckin.Status == CheckinStatusPending {
		return &ScanQRResponse{
			CheckinID: existingCheckin.ID,
			TPS: TPSInfo{
				ID:   tps.ID,
				Code: tps.Code,
				Name: tps.Name,
			},
			Status:  CheckinStatusPending,
			Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
		}, nil
	}

	// Create new checkin
	checkin := &TPSCheckin{
		TPSID:      tps.ID,
		VoterID:    voterID,
		ElectionID: tps.ElectionID,
		Status:     CheckinStatusPending,
		ScanAt:     time.Now(),
	}

	if err := s.repo.CreateCheckin(ctx, checkin); err != nil {
		return nil, err
	}

	return &ScanQRResponse{
		CheckinID: checkin.ID,
		TPS: TPSInfo{
			ID:   tps.ID,
			Code: tps.Code,
			Name: tps.Name,
		},
		Status:  CheckinStatusPending,
		Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
		ScanAt:  checkin.ScanAt,
	}, nil
}

func (s *Service) GetCheckinStatus(ctx context.Context, voterID, electionID int64) (*CheckinStatusResponse, error) {
	checkin, err := s.repo.GetCheckinByVoter(ctx, voterID, electionID)
	if err != nil || checkin == nil {
		return &CheckinStatusResponse{
			HasActiveCheckin: false,
		}, nil
	}

	if checkin.Status == CheckinStatusUsed || checkin.Status == CheckinStatusExpired {
		return &CheckinStatusResponse{
			HasActiveCheckin: false,
		}, nil
	}

	tps, _ := s.repo.GetByID(ctx, checkin.TPSID)
	tpsInfo := &TPSInfo{}
	if tps != nil {
		tpsInfo.ID = tps.ID
		tpsInfo.Code = tps.Code
		tpsInfo.Name = tps.Name
	}

	return &CheckinStatusResponse{
		HasActiveCheckin: true,
		Status:           &checkin.Status,
		TPS:              tpsInfo,
		ScanAt:           &checkin.ScanAt,
		ApprovedAt:       checkin.ApprovedAt,
		ExpiresAt:        checkin.ExpiresAt,
	}, nil
}

// TPS Panel Operations
func (s *Service) ListCheckinQueue(ctx context.Context, tpsID int64, status string, page, limit int, userRole string, userID int64) (*CheckinQueueResponse, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}

	if !hasPanelAccess(userRole) {
		// Non-admin/operator must be assigned panitia
		hasAccess, _ := s.repo.IsPanitiaAssigned(ctx, tpsID, userID)
		if !hasAccess {
			return nil, ErrTPSAccessDenied
		}
	}

	checkins, err := s.repo.ListCheckins(ctx, tpsID, status, page, limit)
	if err != nil {
		return nil, err
	}

	items := make([]CheckinQueueItem, 0, len(checkins))
	for _, c := range checkins {
		voterInfo, _ := s.repo.GetVoterInfo(ctx, c.VoterID)
		if voterInfo == nil {
			voterInfo = &VoterInfo{ID: c.VoterID}
		}

		hasVoted, _ := s.repo.HasVoterVoted(ctx, c.VoterID, c.ElectionID)

		items = append(items, CheckinQueueItem{
			ID:       c.ID,
			Voter:    *voterInfo,
			Status:   c.Status,
			ScanAt:   c.ScanAt,
			HasVoted: hasVoted,
		})
	}

	return &CheckinQueueResponse{
		Items: items,
	}, nil
}

func (s *Service) ApproveCheckin(ctx context.Context, tpsID, checkinID, approverID int64) (*ApproveCheckinResponse, error) {
	// Access check handled at handler level for admin/operator.
	checkin, err := s.repo.GetCheckin(ctx, checkinID)
	if err != nil {
		return nil, ErrCheckinNotFound
	}

	if checkin.TPSID != tpsID {
		return nil, ErrTPSAccessDenied
	}

	if checkin.Status != CheckinStatusPending {
		return nil, ErrCheckinNotPending
	}

	// Check if voter already voted
	hasVoted, _ := s.repo.HasVoterVoted(ctx, checkin.VoterID, checkin.ElectionID)
	if hasVoted {
		return nil, ErrAlreadyVoted
	}

	now := time.Now()
	expiresAt := now.Add(15 * time.Minute)

	checkin.Status = CheckinStatusApproved
	checkin.ApprovedByID = &approverID
	checkin.ApprovedAt = &now
	checkin.ExpiresAt = &expiresAt

	if err := s.repo.UpdateCheckin(ctx, checkin); err != nil {
		return nil, err
	}

	voterInfo, _ := s.repo.GetVoterInfo(ctx, checkin.VoterID)
	if voterInfo == nil {
		voterInfo = &VoterInfo{ID: checkin.VoterID}
	}

	tps, _ := s.repo.GetByID(ctx, tpsID)
	tpsInfo := TPSInfo{}
	if tps != nil {
		tpsInfo.ID = tps.ID
		tpsInfo.Code = tps.Code
		tpsInfo.Name = tps.Name
	}

	return &ApproveCheckinResponse{
		CheckinID:  checkin.ID,
		Status:     checkin.Status,
		Voter:      *voterInfo,
		TPS:        tpsInfo,
		ApprovedAt: now,
	}, nil
}

func (s *Service) RejectCheckin(ctx context.Context, tpsID, checkinID, approverID int64, reason string) (*RejectCheckinResponse, error) {
	// Access check handled at handler level for admin/operator.
	checkin, err := s.repo.GetCheckin(ctx, checkinID)
	if err != nil {
		return nil, ErrCheckinNotFound
	}

	if checkin.TPSID != tpsID {
		return nil, ErrTPSAccessDenied
	}

	if checkin.Status != CheckinStatusPending {
		return nil, ErrCheckinNotPending
	}

	checkin.Status = CheckinStatusRejected
	checkin.ApprovedByID = &approverID
	checkin.RejectionReason = &reason

	if err := s.repo.UpdateCheckin(ctx, checkin); err != nil {
		return nil, err
	}

	return &RejectCheckinResponse{
		CheckinID: checkin.ID,
		Status:    checkin.Status,
		Reason:    reason,
	}, nil
}

func (s *Service) GetTPSSummary(ctx context.Context, tpsID, userID int64, userRole string) (*TPSSummaryResponse, error) {
	if !hasPanelAccess(userRole) {
		hasAccess, _ := s.repo.IsPanitiaAssigned(ctx, tpsID, userID)
		if !hasAccess {
			return nil, ErrTPSAccessDenied
		}
	}

	tps, err := s.repo.GetByID(ctx, tpsID)
	if err != nil {
		return nil, ErrTPSNotFound
	}

	stats, _ := s.repo.GetStats(ctx, tpsID)
	if stats == nil {
		stats = &TPSStats{}
	}

	response := &TPSSummaryResponse{
		ID:        tps.ID,
		Code:      tps.Code,
		Name:      tps.Name,
		Location:  tps.Location,
		Status:    tps.Status,
		OpenTime:  tps.OpenTime,
		CloseTime: tps.CloseTime,
		Stats:     *stats,
	}

	if tps.VotingDate != nil {
		response.VotingDate = tps.VotingDate.Format("2006-01-02")
	}

	return response, nil
}

// Utility
func (s *Service) GenerateQRSecret() string {
	bytes := make([]byte, 6)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}
