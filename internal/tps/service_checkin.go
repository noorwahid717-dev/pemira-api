package tps

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CheckinService struct {
	db *pgxpool.Pool
}

func NewCheckinService(db *pgxpool.Pool) *CheckinService {
	return &CheckinService{
		db: db,
	}
}

func (s *CheckinService) withTx(ctx context.Context, fn func(tx pgx.Tx) error) error {
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()
	if err := fn(tx); err != nil {
		_ = tx.Rollback(ctx)
		return err
	}
	return tx.Commit(ctx)
}

// CheckinScan - Mahasiswa scan QR di TPS
func (s *CheckinService) CheckinScan(
	ctx context.Context,
	voterID int64,
	qrPayload string,
) (*ScanQRResponse, error) {
	// Parse QR payload
	tpsCode, qrSecret, err := s.parseQRPayload(qrPayload)
	if err != nil {
		return nil, ErrQRInvalid
	}

	var result *ScanQRResponse

	err = s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Load QR entry & TPS
		qr, err := s.findActiveQRByCodeAndSecret(ctx, tx, tpsCode, qrSecret)
		if err != nil {
			return err
		}

		tpsEntry, err := s.getTPSByID(ctx, tx, qr.TPSID)
		if err != nil {
			return err
		}

		if tpsEntry.Status != StatusActive {
			return ErrTPSInactive
		}

		// 2. Cek election & fase
		election, err := s.getElectionByID(ctx, tx, tpsEntry.ElectionID)
		if err != nil {
			return err
		}

		if election.Status != "VOTING_OPEN" {
			return ErrElectionNotOpen
		}

		// 3. Load voter + status
		voter, err := s.getVoterByID(ctx, tx, election.ID, voterID)
		if err != nil {
			return err
		}

		if !voter.IsEligible {
			return ErrNotEligible
		}

		status, err := s.getVoterStatus(ctx, tx, election.ID, voterID)
		if err != nil {
			return err
		}

		if status.HasVoted {
			return ErrAlreadyVoted
		}

		// 4. Cek apakah sudah ada checkin pending
		existingCheckin, _ := s.getPendingCheckin(ctx, tx, voterID, election.ID)
		if existingCheckin != nil {
			// Return existing pending checkin
			result = &ScanQRResponse{
				CheckinID: existingCheckin.ID,
				TPS: TPSInfo{
					ID:   tpsEntry.ID,
					Code: tpsEntry.Code,
					Name: tpsEntry.Name,
				},
				Status:  CheckinStatusPending,
				Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
				ScanAt:  existingCheckin.ScanAt,
			}
			return nil
		}

		// 5. Buat row tps_checkins status PENDING
		now := time.Now().UTC()
		checkinID, err := s.insertCheckin(ctx, tx, &TPSCheckin{
			ElectionID: election.ID,
			TPSID:      tpsEntry.ID,
			VoterID:    voter.ID,
			Status:     CheckinStatusPending,
			ScanAt:     now,
		})
		if err != nil {
			return err
		}

		// 6. (Opsional) Audit log
		_ = s.logAudit(ctx, tx, AuditLog{
			ActorVoterID: &voter.ID,
			Action:       "TPS_CHECKIN_CREATED",
			EntityType:   "TPS_CHECKIN",
			EntityID:     checkinID,
			Metadata: map[string]interface{}{
				"tps_id": tpsEntry.ID,
			},
		})

		// 7. Build result
		result = &ScanQRResponse{
			CheckinID: checkinID,
			TPS: TPSInfo{
				ID:   tpsEntry.ID,
				Code: tpsEntry.Code,
				Name: tpsEntry.Name,
			},
			Status:  CheckinStatusPending,
			Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
			ScanAt:  now,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ApproveCheckin - Panitia TPS menyetujui check-in
func (s *CheckinService) ApproveCheckin(
	ctx context.Context,
	operatorUserID int64,
	tpsID int64,
	checkinID int64,
) (*ApproveCheckinResponse, error) {
	var result *ApproveCheckinResponse

	err := s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Validasi operator punya akses ke TPS
		hasAccess, err := s.operatorHasAccess(ctx, tx, tpsID, operatorUserID)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrTPSAccessDenied
		}

		// 2. Ambil check-in
		checkin, err := s.getCheckinByID(ctx, tx, checkinID)
		if err != nil {
			return err
		}

		if checkin.TPSID != tpsID {
			return ErrTPSAccessDenied
		}

		if checkin.Status != CheckinStatusPending {
			return ErrCheckinNotPending
		}

		// 3. Ambil election + voter status
		election, err := s.getElectionByID(ctx, tx, checkin.ElectionID)
		if err != nil {
			return err
		}

		if election.Status != "VOTING_OPEN" {
			return ErrElectionNotOpen
		}

		status, err := s.getVoterStatus(ctx, tx, election.ID, checkin.VoterID)
		if err != nil {
			return err
		}
		if status.HasVoted {
			return ErrAlreadyVoted
		}

		// 4. Update check-in -> APPROVED
		now := time.Now().UTC()
		expiresAt := now.Add(15 * time.Minute)

		err = s.updateCheckinToApproved(ctx, tx, checkinID, operatorUserID, now, expiresAt)
		if err != nil {
			return err
		}

		// 5. Load voter & TPS info untuk response
		voter, err := s.getVoterByID(ctx, tx, election.ID, checkin.VoterID)
		if err != nil {
			return err
		}

		tpsEntry, err := s.getTPSByID(ctx, tx, tpsID)
		if err != nil {
			return err
		}

		// 6. Audit
		_ = s.logAudit(ctx, tx, AuditLog{
			ActorUserID: &operatorUserID,
			Action:      "TPS_CHECKIN_APPROVED",
			EntityType:  "TPS_CHECKIN",
			EntityID:    checkin.ID,
			Metadata: map[string]interface{}{
				"tps_id":     tpsEntry.ID,
				"voter_id":   voter.ID,
				"expires_at": expiresAt,
			},
		})

		// Build result
		result = &ApproveCheckinResponse{
			CheckinID: checkin.ID,
			Status:    CheckinStatusApproved,
			Voter: VoterInfo{
				ID:   voter.ID,
				NIM:  voter.NIM,
				Name: voter.Name,
			},
			TPS: TPSInfo{
				ID:   tpsEntry.ID,
				Code: tpsEntry.Code,
				Name: tpsEntry.Name,
			},
			ApprovedAt: now,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// RejectCheckin - Panitia TPS menolak check-in
func (s *CheckinService) RejectCheckin(
	ctx context.Context,
	operatorUserID int64,
	tpsID int64,
	checkinID int64,
	reason string,
) (*RejectCheckinResponse, error) {
	var result *RejectCheckinResponse

	err := s.withTx(ctx, func(tx pgx.Tx) error {
		// 1. Validasi operator punya akses ke TPS
		hasAccess, err := s.operatorHasAccess(ctx, tx, tpsID, operatorUserID)
		if err != nil {
			return err
		}
		if !hasAccess {
			return ErrTPSAccessDenied
		}

		// 2. Ambil check-in
		checkin, err := s.getCheckinByID(ctx, tx, checkinID)
		if err != nil {
			return err
		}

		if checkin.TPSID != tpsID {
			return ErrTPSAccessDenied
		}

		if checkin.Status != CheckinStatusPending {
			return ErrCheckinNotPending
		}

		// 3. Update check-in -> REJECTED
		err = s.updateCheckinToRejected(ctx, tx, checkinID, operatorUserID, reason)
		if err != nil {
			return err
		}

		// 4. Audit
		_ = s.logAudit(ctx, tx, AuditLog{
			ActorUserID: &operatorUserID,
			Action:      "TPS_CHECKIN_REJECTED",
			EntityType:  "TPS_CHECKIN",
			EntityID:    checkin.ID,
			Metadata: map[string]interface{}{
				"tps_id": tpsID,
				"reason": reason,
			},
		})

		result = &RejectCheckinResponse{
			CheckinID: checkin.ID,
			Status:    CheckinStatusRejected,
			Reason:    reason,
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// ===== Helper functions (repository methods with tx) =====

func (s *CheckinService) parseQRPayload(payload string) (tpsCode, qrSecret string, err error) {
	// Format: "PEMIRA|TPS01|c9423e5f97d4"
	parts := strings.Split(payload, "|")
	if len(parts) != 3 || parts[0] != "PEMIRA" {
		return "", "", ErrQRInvalid
	}
	return parts[1], parts[2], nil
}

func (s *CheckinService) findActiveQRByCodeAndSecret(ctx context.Context, tx pgx.Tx, tpsCode, secret string) (*TPSQR, error) {
	query := `
		SELECT qr.id, qr.tps_id, qr.qr_secret_suffix, qr.is_active, qr.revoked_at, qr.created_at
		FROM tps_qr qr
		JOIN tps t ON t.id = qr.tps_id
		WHERE t.code = $1 AND qr.qr_secret_suffix = $2 AND qr.is_active = true
	`

	var qr TPSQR
	err := tx.QueryRow(ctx, query, tpsCode, secret).Scan(
		&qr.ID, &qr.TPSID, &qr.QRToken, &qr.IsActive, &qr.RotatedAt, &qr.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrQRInvalid
		}
		return nil, err
	}

	return &qr, nil
}

func (s *CheckinService) getTPSByID(ctx context.Context, tx pgx.Tx, id int64) (*TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at
		FROM tps
		WHERE id = $1
	`

	var tps TPS
	err := tx.QueryRow(ctx, query, id).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrTPSNotFound
		}
		return nil, err
	}

	return &tps, nil
}

type Election struct {
	ID     int64
	Status string
}

func (s *CheckinService) getElectionByID(ctx context.Context, tx pgx.Tx, id int64) (*Election, error) {
	query := `SELECT id, status FROM elections WHERE id = $1`

	var election Election
	err := tx.QueryRow(ctx, query, id).Scan(&election.ID, &election.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrElectionNotOpen
		}
		return nil, err
	}

	return &election, nil
}

type Voter struct {
	ID         int64
	NIM        string
	Name       string
	IsEligible bool
}

func (s *CheckinService) getVoterByID(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*Voter, error) {
	query := `
		SELECT v.id, v.nim, v.name, v.is_eligible
		FROM voters v
		WHERE v.election_id = $1 AND v.id = $2
	`

	var voter Voter
	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(
		&voter.ID, &voter.NIM, &voter.Name, &voter.IsEligible,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotEligible
		}
		return nil, err
	}

	return &voter, nil
}

type VoterStatus struct {
	HasVoted bool
}

func (s *CheckinService) getVoterStatus(ctx context.Context, tx pgx.Tx, electionID, voterID int64) (*VoterStatus, error) {
	query := `
		SELECT has_voted
		FROM voter_status
		WHERE election_id = $1 AND voter_id = $2
	`

	var status VoterStatus
	err := tx.QueryRow(ctx, query, electionID, voterID).Scan(&status.HasVoted)
	if err != nil {
		if err == pgx.ErrNoRows {
			// Jika belum ada status, default false
			return &VoterStatus{HasVoted: false}, nil
		}
		return nil, err
	}

	return &status, nil
}

func (s *CheckinService) getPendingCheckin(ctx context.Context, tx pgx.Tx, voterID, electionID int64) (*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at,
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE voter_id = $1 AND election_id = $2 AND status = $3
		ORDER BY created_at DESC
		LIMIT 1
	`

	var checkin TPSCheckin
	err := tx.QueryRow(ctx, query, voterID, electionID, CheckinStatusPending).Scan(
		&checkin.ID, &checkin.TPSID, &checkin.VoterID, &checkin.ElectionID,
		&checkin.Status, &checkin.ScanAt, &checkin.ApprovedAt, &checkin.ApprovedByID,
		&checkin.RejectionReason, &checkin.ExpiresAt, &checkin.CreatedAt, &checkin.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &checkin, nil
}

func (s *CheckinService) insertCheckin(ctx context.Context, tx pgx.Tx, checkin *TPSCheckin) (int64, error) {
	query := `
		INSERT INTO tps_checkins (tps_id, voter_id, election_id, status, scan_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id
	`

	var id int64
	err := tx.QueryRow(ctx, query,
		checkin.TPSID, checkin.VoterID, checkin.ElectionID, checkin.Status, checkin.ScanAt,
	).Scan(&id)

	return id, err
}

func (s *CheckinService) operatorHasAccess(ctx context.Context, tx pgx.Tx, tpsID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tps_panitia
			WHERE tps_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := tx.QueryRow(ctx, query, tpsID, userID).Scan(&exists)
	return exists, err
}

func (s *CheckinService) getCheckinByID(ctx context.Context, tx pgx.Tx, id int64) (*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at,
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE id = $1
	`

	var checkin TPSCheckin
	err := tx.QueryRow(ctx, query, id).Scan(
		&checkin.ID, &checkin.TPSID, &checkin.VoterID, &checkin.ElectionID,
		&checkin.Status, &checkin.ScanAt, &checkin.ApprovedAt, &checkin.ApprovedByID,
		&checkin.RejectionReason, &checkin.ExpiresAt, &checkin.CreatedAt, &checkin.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrCheckinNotFound
		}
		return nil, err
	}

	return &checkin, nil
}

func (s *CheckinService) updateCheckinToApproved(ctx context.Context, tx pgx.Tx, checkinID, approverID int64, now, expiresAt time.Time) error {
	query := `
		UPDATE tps_checkins
		SET status = $1, approved_by_id = $2, approved_at = $3, expires_at = $4, updated_at = NOW()
		WHERE id = $5
	`

	_, err := tx.Exec(ctx, query, CheckinStatusApproved, approverID, now, expiresAt, checkinID)
	return err
}

func (s *CheckinService) updateCheckinToRejected(ctx context.Context, tx pgx.Tx, checkinID, approverID int64, reason string) error {
	query := `
		UPDATE tps_checkins
		SET status = $1, approved_by_id = $2, rejection_reason = $3, updated_at = NOW()
		WHERE id = $4
	`

	_, err := tx.Exec(ctx, query, CheckinStatusRejected, approverID, reason, checkinID)
	return err
}

type AuditLog struct {
	ActorVoterID *int64
	ActorUserID  *int64
	Action       string
	EntityType   string
	EntityID     int64
	Metadata     map[string]interface{}
}

func (s *CheckinService) logAudit(ctx context.Context, tx pgx.Tx, log AuditLog) error {
	// Simplified audit logging - adjust based on your audit table schema
	query := `
		INSERT INTO audit_logs (actor_voter_id, actor_user_id, action, entity_type, entity_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW())
	`

	metadataJSON := fmt.Sprintf("%v", log.Metadata) // Simple string representation, use json.Marshal in production

	_, err := tx.Exec(ctx, query,
		log.ActorVoterID, log.ActorUserID, log.Action, log.EntityType, log.EntityID, metadataJSON,
	)

	// Don't fail the transaction if audit fails
	if err != nil {
		// Log error but don't return it
		return nil
	}

	return nil
}
