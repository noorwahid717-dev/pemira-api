package tps

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

type PostgresRepository struct {
	db *sql.DB
}

func NewPostgresRepository(db *sql.DB) Repository {
	return &PostgresRepository{db: db}
}

// TPS Management
func (r *PostgresRepository) GetByID(ctx context.Context, id int64) (*TPS, error) {
	query := `
		SELECT id, election_id, code, name, location, status, voting_date,
		       open_time, close_time, capacity_estimate, area_faculty_id,
		       created_at, updated_at
		FROM tps
		WHERE id = $1
	`
	
	var tps TPS
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
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
		       created_at, updated_at
		FROM tps
		WHERE code = $1
	`
	
	var tps TPS
	err := r.db.QueryRowContext(ctx, query, code).Scan(
		&tps.ID, &tps.ElectionID, &tps.Code, &tps.Name, &tps.Location,
		&tps.Status, &tps.VotingDate, &tps.OpenTime, &tps.CloseTime,
		&tps.CapacityEstimate, &tps.AreaFacultyID, &tps.CreatedAt, &tps.UpdatedAt,
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
		       created_at, updated_at
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
		                 open_time, close_time, capacity_estimate, area_faculty_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at
	`
	
	return r.db.QueryRowContext(ctx, query,
		tps.ElectionID, tps.Code, tps.Name, tps.Location, tps.Status,
		tps.VotingDate, tps.OpenTime, tps.CloseTime, tps.CapacityEstimate,
		tps.AreaFacultyID,
	).Scan(&tps.ID, &tps.CreatedAt, &tps.UpdatedAt)
}

func (r *PostgresRepository) Update(ctx context.Context, tps *TPS) error {
	query := `
		UPDATE tps
		SET name = $1, location = $2, status = $3, voting_date = $4,
		    open_time = $5, close_time = $6, capacity_estimate = $7
		WHERE id = $8
	`
	
	result, err := r.db.ExecContext(ctx, query,
		tps.Name, tps.Location, tps.Status, tps.VotingDate,
		tps.OpenTime, tps.CloseTime, tps.CapacityEstimate, tps.ID,
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
		INSERT INTO tps_qr (tps_id, qr_secret_suffix, is_active)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	
	return r.db.QueryRowContext(ctx, query,
		qr.TPSID, qr.QRSecretSuffix, qr.IsActive,
	).Scan(&qr.ID, &qr.CreatedAt)
}

func (r *PostgresRepository) GetActiveQR(ctx context.Context, tpsID int64) (*TPSQR, error) {
	query := `
		SELECT id, tps_id, qr_secret_suffix, is_active, revoked_at, created_at
		FROM tps_qr
		WHERE tps_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT 1
	`
	
	var qr TPSQR
	err := r.db.QueryRowContext(ctx, query, tpsID).Scan(
		&qr.ID, &qr.TPSID, &qr.QRSecretSuffix, &qr.IsActive,
		&qr.RevokedAt, &qr.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, nil
	}
	
	return &qr, err
}

func (r *PostgresRepository) GetQRBySecret(ctx context.Context, tpsCode, secret string) (*TPSQR, error) {
	query := `
		SELECT q.id, q.tps_id, q.qr_secret_suffix, q.is_active, q.revoked_at, q.created_at
		FROM tps_qr q
		JOIN tps t ON t.id = q.tps_id
		WHERE t.code = $1 AND q.qr_secret_suffix = $2
		ORDER BY q.created_at DESC
		LIMIT 1
	`
	
	var qr TPSQR
	err := r.db.QueryRowContext(ctx, query, tpsCode, secret).Scan(
		&qr.ID, &qr.TPSID, &qr.QRSecretSuffix, &qr.IsActive,
		&qr.RevokedAt, &qr.CreatedAt,
	)
	
	if err == sql.ErrNoRows {
		return nil, ErrQRInvalid
	}
	
	return &qr, err
}

func (r *PostgresRepository) RevokeQR(ctx context.Context, qrID int64) error {
	query := `
		UPDATE tps_qr
		SET is_active = FALSE, revoked_at = $1
		WHERE id = $2
	`
	
	_, err := r.db.ExecContext(ctx, query, time.Now(), qrID)
	return err
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
		FROM voter_eligibility 
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
			SELECT 1 FROM votes 
			WHERE voter_id = $1 AND election_id = $2
		)
	`
	
	var hasVoted bool
	err := r.db.QueryRowContext(ctx, query, voterID, electionID).Scan(&hasVoted)
	return hasVoted, err
}

func (r *PostgresRepository) GetVoterInfo(ctx context.Context, voterID int64) (*VoterInfo, error) {
	query := `
		SELECT id, nim, name, faculty, study_program, cohort_year, academic_status
		FROM users
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
