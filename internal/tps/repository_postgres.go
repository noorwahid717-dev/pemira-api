package tps

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"pemira-api/internal/auth"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

// NewPostgresRepositoryFromPool adapts a pgxpool.Pool to the existing sql-based repository.
// It uses the stdlib compatibility wrapper to reuse existing implementation without refactor.
func NewPostgresRepositoryFromPool(pool *pgxpool.Pool) Repository {
	sqlDB := stdlib.OpenDBFromPool(pool)
	return &PostgresRepository{db: sqlDB}
}

// TPS Management
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at, pic_name, pic_phone, notes
		FROM tps
		WHERE id = $1
	`

	var tps TPS
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
		&tps.PICName, &tps.PICPhone, &tps.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTPSNotFound
	}

	return &tps, err
}

func (r *PostgresRepository) GetByIDElection(ctx context.Context, electionID, id int64) (*TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at, pic_name, pic_phone, notes
		FROM tps
		WHERE id = $1 AND election_id = $2
	`

	var tps TPS
	err := r.db.QueryRowContext(ctx, query, id, electionID).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
		&tps.PICName, &tps.PICPhone, &tps.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTPSNotFound
	}

	return &tps, err
}

func (r *PostgresRepository) GetByCode(ctx context.Context, code string) (*TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at, pic_name, pic_phone, notes
		FROM tps
		WHERE code = $1
	`

	var tps TPS
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
		&tps.PICName, &tps.PICPhone, &tps.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, ErrTPSNotFound
	}

	return &tps, err
}

func (r *PostgresRepository) List(ctx context.Context, filter ListFilter) ([]*TPS, int, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at, pic_name, pic_phone, notes
		FROM tps
		WHERE 1=1
	`
	countQuery := "SELECT COUNT(*) FROM tps WHERE 1=1"
	args := []interface{}{}
	argPos := 1

	if filter.Status != "" {
		query += " AND status = $" + string(rune('0'+argPos))
		countQuery += " AND status = $" + string(rune('0'+argPos))
		args = append(args, filter.Status)
		argPos++
	}

	if filter.ElectionID > 0 {
		query += " AND election_id = $" + string(rune('0'+argPos))
		countQuery += " AND election_id = $" + string(rune('0'+argPos))
		args = append(args, filter.ElectionID)
		argPos++
	}

	// Count total
	var total int
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	// Pagination
	query += " ORDER BY created_at DESC"
	offset := (filter.Page - 1) * filter.Limit
	query += " LIMIT $" + string(rune('0'+argPos)) + " OFFSET $" + string(rune('0'+argPos+1))
	args = append(args, filter.Limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	tpsList := make([]*TPS, 0)
	for rows.Next() {
		var tps TPS
		err := rows.Scan(
			&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
			&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
			&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
			&tps.PICName, &tps.PICPhone, &tps.Notes,
		)
		if err != nil {
			return nil, 0, err
		}
		tpsList = append(tpsList, &tps)
	}

	return tpsList, total, nil
}

func (r *PostgresRepository) Create(ctx context.Context, tps *TPS) error {
	query := `
		INSERT INTO tps (election_id, code, name, location, status, voting_date,
		                 open_time, close_time, capacity_estimate, area_faculty_id, pic_name, pic_phone, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		tps.ElectionID, tps.Code, tps.Name, tps.Location, tps.Status,
		tps.VotingDate, tps.OpenTime, tps.CloseTime, tps.CapacityEstimate,
		tps.AreaFacultyID, tps.PICName, tps.PICPhone, tps.Notes,
	).Scan(&tps.ID, &tps.CreatedAt, &tps.UpdatedAt)
}

func (r *PostgresRepository) Update(ctx context.Context, tps *TPS) error {
	query := `
		UPDATE tps
		SET name = $1, location = $2, status = $3, voting_date = $4,
		    open_time = $5, close_time = $6, capacity_estimate = $7,
		    pic_name = $8, pic_phone = $9, notes = $10
		WHERE id = $11
	`

	result, err := r.db.ExecContext(ctx, query,
		tps.Name, tps.Location, tps.Status, tps.VotingDate,
		tps.OpenTime, tps.CloseTime, tps.CapacityEstimate,
		tps.PICName, tps.PICPhone, tps.Notes, tps.ID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrTPSNotFound
	}

	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, electionID, id int64) error {
	// Ensure election match
	var exists bool
	if err := r.db.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM tps WHERE id = $1 AND election_id = $2)`, id, electionID).Scan(&exists); err != nil {
		return err
	}
	if !exists {
		return ErrTPSNotFound
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM tps WHERE id = $1`, id)
	return err
}

func (r *PostgresRepository) GetStats(ctx context.Context, tpsID int64) (*TPSStats, error) {
	query := `
		SELECT 
			COUNT(CASE WHEN status IN ('APPROVED', 'USED', 'EXPIRED') THEN 1 END) as total_checkins,
			COUNT(CASE WHEN status = 'PENDING' THEN 1 END) as pending_checkins,
			COUNT(CASE WHEN status = 'APPROVED' THEN 1 END) as approved_checkins,
			COUNT(CASE WHEN status = 'REJECTED' THEN 1 END) as rejected_checkins
		FROM tps_checkins
		WHERE tps_id = $1
	`

	var stats TPSStats
	err := r.db.QueryRowContext(ctx, query, tpsID).Scan(
		&stats.TotalCheckins, &stats.PendingCheckins,
		&stats.ApprovedCheckins, &stats.RejectedCheckins,
	)

	// Get total votes from votes table (if exists)
	// This is a placeholder - adjust based on your voting module schema
	voteQuery := `
		SELECT COUNT(*) 
		FROM votes 
		WHERE tps_id = $1
	`
	_ = r.db.QueryRowContext(ctx, voteQuery, tpsID).Scan(&stats.TotalVotes)

	return &stats, err
}

// QR Management
func (r *PostgresRepository) CreateQR(ctx context.Context, qr *TPSQR) error {
	query := `
		INSERT INTO tps_qr (tps_id, qr_token, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`

	return r.db.QueryRowContext(ctx, query,
		qr.TPSID, qr.QRToken, qr.IsActive,
	).Scan(&qr.ID, &qr.CreatedAt)
}

func (r *PostgresRepository) GetActiveQR(ctx context.Context, tpsID int64) (*TPSQR, error) {
	query := `
		SELECT id, tps_id, qr_token, is_active, rotated_at, created_at
		FROM tps_qr
		WHERE tps_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`

	var qr TPSQR
	err := r.db.QueryRowContext(ctx, query, tpsID).Scan(
		&qr.ID, &qr.TPSID, &qr.QRToken, &qr.IsActive,
		&qr.RotatedAt, &qr.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &qr, err
}

func (r *PostgresRepository) GetQRBySecret(ctx context.Context, tpsCode, secret string) (*TPSQR, error) {
	query := `
		SELECT q.id, q.tps_id, q.qr_token, q.is_active, q.rotated_at, q.created_at
		FROM tps_qr q
		JOIN tps t ON t.id = q.tps_id
		WHERE t.code = $1 AND q.qr_token = $2
		ORDER BY q.created_at DESC
		LIMIT 1
	`

	var qr TPSQR
	err := r.db.QueryRowContext(ctx, query, tpsCode, secret).Scan(
		&qr.ID, &qr.TPSID, &qr.QRToken, &qr.IsActive,
		&qr.RotatedAt, &qr.CreatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrQRInvalid
	}

	return &qr, err
}

func (r *PostgresRepository) RevokeQR(ctx context.Context, qrID int64) error {
	query := `
		UPDATE tps_qr
		SET is_active = FALSE, rotated_at = $1
		WHERE id = $2
	`

	_, err := r.db.ExecContext(ctx, query, time.Now(), qrID)
	return err
}

func (r *PostgresRepository) GetQRMetadata(ctx context.Context, tpsID int64) (*QRInfo, error) {
	qr, err := r.GetActiveQR(ctx, tpsID)
	if err != nil {
		return nil, err
	}
	if qr == nil {
		return nil, ErrQRInvalid
	}
	return &QRInfo{
		ID:        qr.ID,
		QRToken:   qr.QRToken,
		IsActive:  qr.IsActive,
		CreatedAt: qr.CreatedAt.Format(time.RFC3339),
	}, nil
}

// local short token generator for TPS QR (kept simple for panel flow)
func generateQRTokenSimple() string {
	bytes := make([]byte, 6)
	_, _ = rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

func (r *PostgresRepository) RotateQR(ctx context.Context, tpsID int64) (*QRInfo, error) {
	// Revoke existing
	if qr, _ := r.GetActiveQR(ctx, tpsID); qr != nil {
		_ = r.RevokeQR(ctx, qr.ID)
	}
	secret := generateQRTokenSimple()
	qr := &TPSQR{
		TPSID:    tpsID,
		QRToken:  secret,
		IsActive: true,
	}
	if err := r.CreateQR(ctx, qr); err != nil {
		return nil, err
	}
	return &QRInfo{
		ID:        qr.ID,
		QRToken:   qr.QRToken,
		IsActive:  qr.IsActive,
		CreatedAt: qr.CreatedAt.Format(time.RFC3339),
	}, nil
}

func (r *PostgresRepository) GetQRPrintPayload(ctx context.Context, tpsID int64) (string, error) {
	qr, err := r.GetActiveQR(ctx, tpsID)
	if err != nil || qr == nil {
		return "", ErrQRInvalid
	}
	tpsRow, err := r.GetByID(ctx, tpsID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("PEMIRA|%s|%s", tpsRow.Code, qr.QRToken), nil
}

// Panitia Management
func (r *PostgresRepository) AssignPanitia(ctx context.Context, tpsID int64, members []TPSPanitia) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO tps_panitia (tps_id, user_id, role)
		VALUES ($1, $2, $3)
	`

	for _, member := range members {
		_, err := tx.ExecContext(ctx, query, tpsID, member.UserID, member.Role)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PostgresRepository) GetPanitia(ctx context.Context, tpsID int64) ([]*TPSPanitia, error) {
	query := `
		SELECT id, tps_id, user_id, role, created_at
		FROM tps_panitia
		WHERE tps_id = $1
		ORDER BY created_at
	`

	rows, err := r.db.QueryContext(ctx, query, tpsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*TPSPanitia, 0)
	for rows.Next() {
		var p TPSPanitia
		err := rows.Scan(&p.ID, &p.TPSID, &p.UserID, &p.Role, &p.CreatedAt)
		if err != nil {
			return nil, err
		}
		members = append(members, &p)
	}

	return members, nil
}

func (r *PostgresRepository) IsPanitiaAssigned(ctx context.Context, tpsID, userID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM tps_panitia 
			WHERE tps_id = $1 AND user_id = $2
		)
	`

	var exists bool
	err := r.db.QueryRowContext(ctx, query, tpsID, userID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ClearPanitia(ctx context.Context, tpsID int64) error {
	query := "DELETE FROM tps_panitia WHERE tps_id = $1"
	_, err := r.db.ExecContext(ctx, query, tpsID)
	return err
}

func (r *PostgresRepository) ListOperators(ctx context.Context, tpsID int64) ([]OperatorInfo, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, username, full_name, COALESCE(email,'')
		FROM user_accounts
		WHERE tps_id = $1 AND role = 'TPS_OPERATOR' AND is_active = TRUE
		ORDER BY username
	`, tpsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ops []OperatorInfo
	for rows.Next() {
		var op OperatorInfo
		if err := rows.Scan(&op.ID, &op.Username, &op.Name, &op.Email); err != nil {
			return nil, err
		}
		op.TPSID = &tpsID
		ops = append(ops, op)
	}
	return ops, nil
}

func (r *PostgresRepository) CreateOperator(ctx context.Context, tpsID int64, op OperatorCreate) (*OperatorInfo, error) {
	passwordHash, err := auth.HashPassword(op.Password)
	if err != nil {
		return nil, err
	}

	query := `
		INSERT INTO user_accounts (username, email, password_hash, full_name, role, tps_id, is_active)
		VALUES ($1, $2, $3, $4, 'TPS_OPERATOR', $5, TRUE)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRowContext(ctx, query, op.Username, op.Email, passwordHash, op.Name, tpsID).Scan(&id); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "duplicate") {
			return nil, ErrOperatorExists
		}
		return nil, err
	}

	return &OperatorInfo{
		ID:       id,
		Username: op.Username,
		Name:     op.Name,
		Email:    op.Email,
		TPSID:    &tpsID,
	}, nil
}

func (r *PostgresRepository) DeleteOperator(ctx context.Context, tpsID, userID int64) error {
	res, err := r.db.ExecContext(ctx, `
		DELETE FROM user_accounts
		WHERE id = $1 AND tps_id = $2 AND role = 'TPS_OPERATOR'
	`, userID, tpsID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return ErrOperatorNotFound
	}
	return nil
}

// Check-in Management
func (r *PostgresRepository) CreateCheckin(ctx context.Context, checkin *TPSCheckin) error {
	query := `
		INSERT INTO tps_checkins (tps_id, voter_id, election_id, status, scan_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRowContext(ctx, query,
		checkin.TPSID, checkin.VoterID, checkin.ElectionID,
		checkin.Status, checkin.ScanAt,
	).Scan(&checkin.ID, &checkin.CreatedAt, &checkin.UpdatedAt)
}

func (r *PostgresRepository) GetCheckin(ctx context.Context, id int64) (*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at,
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE id = $1
	`

	var c TPSCheckin
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&c.ID, &c.TPSID, &c.VoterID, &c.ElectionID, &c.Status, &c.ScanAt,
		&c.ApprovedAt, &c.ApprovedByID, &c.RejectionReason, &c.ExpiresAt,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCheckinNotFound
	}

	return &c, err
}

func (r *PostgresRepository) GetCheckinByVoter(ctx context.Context, voterID, electionID int64) (*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at,
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE voter_id = $1 AND election_id = $2
		ORDER BY created_at DESC
		LIMIT 1
	`

	var c TPSCheckin
	err := r.db.QueryRowContext(ctx, query, voterID, electionID).Scan(
		&c.ID, &c.TPSID, &c.VoterID, &c.ElectionID, &c.Status, &c.ScanAt,
		&c.ApprovedAt, &c.ApprovedByID, &c.RejectionReason, &c.ExpiresAt,
		&c.CreatedAt, &c.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}

	return &c, err
}

func (r *PostgresRepository) ListCheckins(ctx context.Context, tpsID int64, status string, page, limit int) ([]*TPSCheckin, error) {
	query := `
		SELECT id, tps_id, voter_id, election_id, status, scan_at,
		       approved_at, approved_by_id, rejection_reason, expires_at,
		       created_at, updated_at
		FROM tps_checkins
		WHERE tps_id = $1
	`

	args := []interface{}{tpsID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	query += " ORDER BY scan_at DESC LIMIT $" + string(rune('0'+len(args)+1)) + " OFFSET $" + string(rune('0'+len(args)+2))
	offset := (page - 1) * limit
	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checkins := make([]*TPSCheckin, 0)
	for rows.Next() {
		var c TPSCheckin
		err := rows.Scan(
			&c.ID, &c.TPSID, &c.VoterID, &c.ElectionID, &c.Status, &c.ScanAt,
			&c.ApprovedAt, &c.ApprovedByID, &c.RejectionReason, &c.ExpiresAt,
			&c.CreatedAt, &c.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		checkins = append(checkins, &c)
	}

	return checkins, nil
}

func (r *PostgresRepository) UpdateCheckin(ctx context.Context, checkin *TPSCheckin) error {
	query := `
		UPDATE tps_checkins
		SET status = $1, approved_at = $2, approved_by_id = $3,
		    rejection_reason = $4, expires_at = $5
		WHERE id = $6
	`

	result, err := r.db.ExecContext(ctx, query,
		checkin.Status, checkin.ApprovedAt, checkin.ApprovedByID,
		checkin.RejectionReason, checkin.ExpiresAt, checkin.ID,
	)

	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return ErrCheckinNotFound
	}

	return nil
}

func (r *PostgresRepository) CountCheckins(ctx context.Context, tpsID int64, status string) (int, error) {
	query := "SELECT COUNT(*) FROM tps_checkins WHERE tps_id = $1"
	args := []interface{}{tpsID}

	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}

	var count int
	err := r.db.QueryRowContext(ctx, query, args...).Scan(&count)
	return count, err
}

// Voter validation (placeholder - adjust to your schema)
func (r *PostgresRepository) IsVoterEligible(ctx context.Context, voterID, electionID int64) (bool, error) {
	query := `
		SELECT is_eligible 
		FROM voter_status 
		WHERE voter_id = $1 AND election_id = $2
	`

	var eligible bool
	err := r.db.QueryRowContext(ctx, query, voterID, electionID).Scan(&eligible)

	if err == sql.ErrNoRows {
		return false, nil
	}

	return eligible, err
}

func (r *PostgresRepository) HasVoterVoted(ctx context.Context, voterID, electionID int64) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM voter_status 
			WHERE voter_id = $1 AND election_id = $2 AND has_voted = TRUE
		)
	`

	var hasVoted bool
	err := r.db.QueryRowContext(ctx, query, voterID, electionID).Scan(&hasVoted)
	return hasVoted, err
}

func (r *PostgresRepository) GetVoterInfo(ctx context.Context, voterID int64) (*VoterInfo, error) {
	query := `
		SELECT id, nim, name, faculty_name, study_program_name, cohort_year, academic_status
		FROM voters
		WHERE id = $1
	`

	var v VoterInfo
	err := r.db.QueryRowContext(ctx, query, voterID).Scan(
		&v.ID, &v.NIM, &v.Name, &v.Faculty, &v.StudyProgram,
		&v.CohortYear, &v.AcademicStatus,
	)

	if err == sql.ErrNoRows {
		return nil, errors.New("voter not found")
	}

	return &v, err
}

// Panel helpers
func (r *PostgresRepository) PanelDashboardStats(ctx context.Context, tpsID, electionID int64) (*PanelDashboardStatsRow, error) {
	stats := &PanelDashboardStatsRow{}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM voter_status 
		WHERE election_id = $1 AND COALESCE(tps_allowed, true) = true
	`, electionID).Scan(&stats.TotalRegistered); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM tps_checkins 
		WHERE tps_id = $1 AND status IN ('APPROVED','USED','VOTED')
	`, tpsID).Scan(&stats.TotalCheckedIn); err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) 
		FROM votes 
		WHERE tps_id = $1
	`, tpsID).Scan(&stats.TotalVoted); err != nil {
		return nil, err
	}

	_ = r.db.QueryRowContext(ctx, `
		SELECT MAX(ts) FROM (
			SELECT MAX(scan_at) AS ts FROM tps_checkins WHERE tps_id = $1
			UNION ALL
			SELECT MAX(cast_at) AS ts FROM votes WHERE tps_id = $1
		) t
	`, tpsID).Scan(&stats.LastActivity)

	return stats, nil
}

func (r *PostgresRepository) PanelListCheckins(ctx context.Context, tpsID int64, status, search string, limit, offset int) ([]PanelCheckinRow, int, error) {
	args := []interface{}{tpsID}
	where := "WHERE c.tps_id = $1"
	argPos := 2

	if status != "" && status != "ALL" {
		where += " AND c.status = $" + strconv.Itoa(argPos)
		args = append(args, status)
		argPos++
	}

	if strings.TrimSpace(search) != "" {
		where += " AND (v.nim ILIKE $" + strconv.Itoa(argPos) + " OR v.name ILIKE $" + strconv.Itoa(argPos) + ")"
		args = append(args, "%"+search+"%")
		argPos++
	}

	countQuery := "SELECT COUNT(*) FROM tps_checkins c JOIN voters v ON v.id = c.voter_id " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT c.id, c.tps_id, c.election_id, c.voter_id, v.name, v.nim, 
		       COALESCE(v.faculty_name,''), COALESCE(v.study_program_name,''),
		       c.status, c.scan_at, c.voted_at
		FROM tps_checkins c
		JOIN voters v ON v.id = c.voter_id
	` + where + `
		ORDER BY c.scan_at DESC
		LIMIT $` + strconv.Itoa(argPos) + ` OFFSET $` + strconv.Itoa(argPos+1)

	args = append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	list := []PanelCheckinRow{}
	for rows.Next() {
		var row PanelCheckinRow
		if err := rows.Scan(
			&row.ID, &row.TPSID, &row.ElectionID, &row.VoterID,
			&row.VoterName, &row.VoterNIM, &row.Faculty, &row.Program,
			&row.Status, &row.ScanAt, &row.VotedAt,
		); err != nil {
			return nil, 0, err
		}
		list = append(list, row)
	}

	return list, total, nil
}

func (r *PostgresRepository) PanelGetCheckin(ctx context.Context, checkinID int64) (*PanelCheckinRow, error) {
	query := `
		SELECT c.id, c.tps_id, c.election_id, c.voter_id, v.name, v.nim,
		       COALESCE(v.faculty_name,''), COALESCE(v.study_program_name,''),
		       c.status, c.scan_at, c.voted_at
		FROM tps_checkins c
		JOIN voters v ON v.id = c.voter_id
		WHERE c.id = $1
	`

	var row PanelCheckinRow
	err := r.db.QueryRowContext(ctx, query, checkinID).Scan(
		&row.ID, &row.TPSID, &row.ElectionID, &row.VoterID,
		&row.VoterName, &row.VoterNIM, &row.Faculty, &row.Program,
		&row.Status, &row.ScanAt, &row.VotedAt,
	)
	if err == sql.ErrNoRows {
		return nil, ErrCheckinNotFound
	}
	return &row, err
}

func (r *PostgresRepository) PanelTimeline(ctx context.Context, tpsID int64) ([]PanelTimelineRow, error) {
	query := `
		WITH checkin_hours AS (
			SELECT date_trunc('hour', scan_at) AS hour_ts,
			       COUNT(*) FILTER (WHERE status IN ('APPROVED','USED','VOTED')) AS checked_in
			FROM tps_checkins
			WHERE tps_id = $1
			GROUP BY 1
		),
		vote_hours AS (
			SELECT date_trunc('hour', cast_at) AS hour_ts,
			       COUNT(*) AS voted
			FROM votes
			WHERE tps_id = $1
			GROUP BY 1
		),
		hours AS (
			SELECT hour_ts FROM checkin_hours
			UNION
			SELECT hour_ts FROM vote_hours
		)
		SELECT to_char(hours.hour_ts, 'HH24:MI') AS hour,
		       COALESCE(ch.checked_in, 0) AS checked_in,
		       COALESCE(vh.voted, 0) AS voted
		FROM hours
		LEFT JOIN checkin_hours ch ON ch.hour_ts = hours.hour_ts
		LEFT JOIN vote_hours vh ON vh.hour_ts = hours.hour_ts
		ORDER BY hours.hour_ts
	`

	rows, err := r.db.QueryContext(ctx, query, tpsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []PanelTimelineRow
	for rows.Next() {
		var row PanelTimelineRow
		if err := rows.Scan(&row.BucketStart, &row.CheckedIn, &row.Voted); err != nil {
			return nil, err
		}
		list = append(list, row)
	}
	return list, nil
}

func (r *PostgresRepository) PanelListTPSByElection(ctx context.Context, electionID int64) ([]PanelTPSListItem, error) {
	query := `
		SELECT t.id, t.code, t.name, t.location, t.status, t.open_time, t.close_time, t.capacity_estimate
		FROM tps t
		WHERE t.election_id = $1
		ORDER BY t.code
	`

	rows, err := r.db.QueryContext(ctx, query, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []PanelTPSListItem
	for rows.Next() {
		var tpsRow PanelTPSListItem
		if err := rows.Scan(
			&tpsRow.TPS.ID,
			&tpsRow.TPS.Code,
			&tpsRow.TPS.Name,
			&tpsRow.Location,
			&tpsRow.Status,
			&tpsRow.OpenTime,
			&tpsRow.CloseTime,
			&tpsRow.Capacity,
		); err != nil {
			return nil, err
		}

		stats, err := r.PanelDashboardStats(ctx, tpsRow.TPS.ID, electionID)
		if err != nil {
			return nil, err
		}
		tpsRow.Stats = *stats
		items = append(items, tpsRow)
	}

	return items, nil
}

func (r *PostgresRepository) GetOperatorInfo(ctx context.Context, userID int64) (*OperatorInfo, error) {
	query := `
		SELECT id, username, full_name, tps_id
		FROM user_accounts
		WHERE id = $1
	`

	var info OperatorInfo
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&info.ID, &info.Username, &info.Name, &info.TPSID)
	if err == sql.ErrNoRows {
		return nil, ErrTPSAccessDenied
	}
	return &info, err
}

func (r *PostgresRepository) ParseRegistrationCode(ctx context.Context, raw string) (*PanelRegistrationCode, error) {
	parts := strings.Split(raw, "|")
	var electionID, voterID int64
	var tpsID *int64

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch {
		case strings.HasPrefix(p, "E:"):
			if val, err := strconv.ParseInt(strings.TrimPrefix(p, "E:"), 10, 64); err == nil {
				electionID = val
			}
		case strings.HasPrefix(p, "V:"):
			if val, err := strconv.ParseInt(strings.TrimPrefix(p, "V:"), 10, 64); err == nil {
				voterID = val
			}
		case strings.HasPrefix(p, "T:"):
			if val, err := strconv.ParseInt(strings.TrimPrefix(p, "T:"), 10, 64); err == nil {
				tpsID = &val
			}
		}
	}

	if electionID == 0 || voterID == 0 {
		return nil, ErrQRInvalid
	}

	return &PanelRegistrationCode{
		ElectionID: electionID,
		VoterID:    voterID,
		TPSID:      tpsID,
		Raw:        raw,
	}, nil
}

func (r *PostgresRepository) CreatePanelCheckin(ctx context.Context, reg PanelRegistrationCode) (*PanelCheckinRow, error) {
	if reg.TPSID == nil {
		return nil, ErrTPSMismatch
	}

	tpsRow, err := r.GetByID(ctx, *reg.TPSID)
	if err != nil {
		return nil, err
	}
	if tpsRow.Status != StatusActive {
		return nil, ErrTPSInactive
	}
	if tpsRow.ElectionID != reg.ElectionID {
		return nil, ErrTPSMismatch
	}

	var isEligible, tpsAllowed, hasVoted bool
	err = r.db.QueryRowContext(ctx, `
		SELECT is_eligible, COALESCE(tps_allowed, true), has_voted
		FROM voter_status
		WHERE voter_id = $1 AND election_id = $2
	`, reg.VoterID, reg.ElectionID).Scan(&isEligible, &tpsAllowed, &hasVoted)
	if err == sql.ErrNoRows {
		return nil, ErrNotEligible
	}
	if err != nil {
		return nil, err
	}
	if !isEligible || !tpsAllowed {
		return nil, ErrNotTPSVoter
	}
	if hasVoted {
		return nil, ErrAlreadyVoted
	}

	var existing int
	_ = r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM tps_checkins 
		WHERE voter_id = $1 AND election_id = $2 AND status IN ('APPROVED','USED','VOTED')
	`, reg.VoterID, reg.ElectionID).Scan(&existing)
	if existing > 0 {
		return nil, ErrCheckinAlreadyExists
	}

	var voter PanelCheckinRow
	err = r.db.QueryRowContext(ctx, `
		INSERT INTO tps_checkins (tps_id, voter_id, election_id, status, scan_at)
		VALUES ($1, $2, $3, 'APPROVED', NOW())
		RETURNING id, tps_id, election_id, voter_id, 'APPROVED', NOW(), NULL
	`, reg.TPSID, reg.VoterID, reg.ElectionID).Scan(
		&voter.ID, &voter.TPSID, &voter.ElectionID, &voter.VoterID, &voter.Status, &voter.ScanAt, &voter.VotedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := r.db.QueryRowContext(ctx, `
		SELECT name, nim, COALESCE(faculty_name,''), COALESCE(study_program_name,'')
		FROM voters WHERE id = $1
	`, reg.VoterID).Scan(&voter.VoterName, &voter.VoterNIM, &voter.Faculty, &voter.Program); err != nil {
		return nil, err
	}

	return &voter, nil
}
