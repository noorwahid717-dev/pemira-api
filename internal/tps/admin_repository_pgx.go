package tps

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type PgAdminRepository struct {
	db *pgxpool.Pool
}

func NewPgAdminRepository(db *pgxpool.Pool) *PgAdminRepository {
	return &PgAdminRepository{db: db}
}

// ErrTPSNotFound is already defined in errors.go

// List returns all TPS
func (r *PgAdminRepository) List(ctx context.Context) ([]TPSDTO, error) {
	const q = `
SELECT 
    t.id,
    t.code,
    t.name,
    t.location,
    t.capacity_estimate,
    CASE WHEN t.status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    t.open_time::TEXT,
    t.close_time::TEXT,
    t.pic_name,
    t.pic_phone,
    t.notes,
    EXISTS(SELECT 1 FROM tps_qr WHERE tps_id = t.id AND is_active = TRUE) as has_active_qr,
    t.created_at,
    t.updated_at
FROM tps t
ORDER BY t.code
`
	rows, err := r.db.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TPSDTO
	for rows.Next() {
		var t TPSDTO
		if err := rows.Scan(
			&t.ID,
			&t.Code,
			&t.Name,
			&t.Location,
			&t.Capacity,
			&t.IsActive,
			&t.OpenTime,
			&t.CloseTime,
			&t.PICName,
			&t.PICPhone,
			&t.Notes,
			&t.HasActiveQR,
			&t.CreatedAt,
			&t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		items = append(items, t)
	}
	return items, rows.Err()
}

// GetByID returns TPS by ID
func (r *PgAdminRepository) GetByID(ctx context.Context, id int64) (*TPSDTO, error) {
	const q = `
SELECT 
    t.id,
    t.code,
    t.name,
    t.location,
    t.capacity_estimate,
    CASE WHEN t.status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    t.open_time::TEXT,
    t.close_time::TEXT,
    t.pic_name,
    t.pic_phone,
    t.notes,
    EXISTS(SELECT 1 FROM tps_qr WHERE tps_id = t.id AND is_active = TRUE) as has_active_qr,
    t.created_at,
    t.updated_at
FROM tps t
WHERE t.id = $1
`
	var t TPSDTO
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Capacity,
		&t.IsActive,
		&t.OpenTime,
		&t.CloseTime,
		&t.PICName,
		&t.PICPhone,
		&t.Notes,
		&t.HasActiveQR,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTPSNotFound
		}
		return nil, err
	}
	return &t, nil
}

// Create creates new TPS
func (r *PgAdminRepository) Create(ctx context.Context, req TPSCreateRequest) (*TPSDTO, error) {
	openTime := "08:00"
	if req.OpenTime != nil {
		openTime = *req.OpenTime
	}
	closeTime := "17:00"
	if req.CloseTime != nil {
		closeTime = *req.CloseTime
	}

	const q = `
INSERT INTO tps (
    election_id,
    code,
    name,
    location,
    capacity_estimate,
    status,
    voting_date,
    open_time,
    close_time,
    pic_name,
    pic_phone,
    notes
) VALUES (1, $1, $2, $3, $4, 'ACTIVE', CURRENT_DATE, $5::TIME, $6::TIME, $7, $8, $9)
RETURNING 
    id,
    code,
    name,
    location,
    capacity_estimate,
    CASE WHEN status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    open_time::TEXT,
    close_time::TEXT,
    pic_name,
    pic_phone,
    notes,
    FALSE as has_active_qr,
    created_at,
    updated_at
`
	var t TPSDTO
	err := r.db.QueryRow(ctx, q,
		req.Code,
		req.Name,
		req.Location,
		req.Capacity,
		openTime,
		closeTime,
		req.PICName,
		req.PICPhone,
		req.Notes,
	).Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Capacity,
		&t.IsActive,
		&t.OpenTime,
		&t.CloseTime,
		&t.PICName,
		&t.PICPhone,
		&t.Notes,
		&t.HasActiveQR,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

// Update updates TPS
func (r *PgAdminRepository) Update(ctx context.Context, id int64, req TPSUpdateRequest) (*TPSDTO, error) {
	updates := []string{}
	args := []interface{}{}
	argPos := 1

	if req.Code != nil {
		updates = append(updates, fmt.Sprintf("code = $%d", argPos))
		args = append(args, *req.Code)
		argPos++
	}

	if req.Name != nil {
		updates = append(updates, fmt.Sprintf("name = $%d", argPos))
		args = append(args, *req.Name)
		argPos++
	}

	if req.Location != nil {
		updates = append(updates, fmt.Sprintf("location = $%d", argPos))
		args = append(args, *req.Location)
		argPos++
	}

	if req.Capacity != nil {
		updates = append(updates, fmt.Sprintf("capacity_estimate = $%d", argPos))
		args = append(args, *req.Capacity)
		argPos++
	}

	if req.IsActive != nil {
		status := "DRAFT"
		if *req.IsActive {
			status = "ACTIVE"
		}
		updates = append(updates, fmt.Sprintf("status = $%d", argPos))
		args = append(args, status)
		argPos++
	}

	if req.OpenTime != nil {
		updates = append(updates, fmt.Sprintf("open_time = $%d::TIME", argPos))
		args = append(args, *req.OpenTime)
		argPos++
	}

	if req.CloseTime != nil {
		updates = append(updates, fmt.Sprintf("close_time = $%d::TIME", argPos))
		args = append(args, *req.CloseTime)
		argPos++
	}

	if req.PICName != nil {
		updates = append(updates, fmt.Sprintf("pic_name = $%d", argPos))
		args = append(args, *req.PICName)
		argPos++
	}

	if req.PICPhone != nil {
		updates = append(updates, fmt.Sprintf("pic_phone = $%d", argPos))
		args = append(args, *req.PICPhone)
		argPos++
	}

	if req.Notes != nil {
		updates = append(updates, fmt.Sprintf("notes = $%d", argPos))
		args = append(args, *req.Notes)
		argPos++
	}

	if len(updates) == 0 {
		return r.GetByID(ctx, id)
	}

	query := fmt.Sprintf(`
UPDATE tps
SET %s
WHERE id = $%d
RETURNING 
    id,
    code,
    name,
    location,
    capacity_estimate,
    CASE WHEN status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    open_time::TEXT,
    close_time::TEXT,
    pic_name,
    pic_phone,
    notes,
    EXISTS(SELECT 1 FROM tps_qr WHERE tps_id = id AND is_active = TRUE) as has_active_qr,
    created_at,
    updated_at
`, strings.Join(updates, ", "), argPos)

	args = append(args, id)

	var t TPSDTO
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Capacity,
		&t.IsActive,
		&t.OpenTime,
		&t.CloseTime,
		&t.PICName,
		&t.PICPhone,
		&t.Notes,
		&t.HasActiveQR,
		&t.CreatedAt,
		&t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTPSNotFound
		}
		return nil, err
	}
	return &t, nil
}

// Delete deletes TPS
func (r *PgAdminRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM tps WHERE id = $1`
	result, err := r.db.Exec(ctx, q, id)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrTPSNotFound
	}

	return nil
}

// ListOperators returns operators for a TPS
func (r *PgAdminRepository) ListOperators(ctx context.Context, tpsID int64) ([]TPSOperatorDTO, error) {
	const q = `
SELECT 
    ua.id,
    ua.username,
    COALESCE(v.name, '') as name,
    COALESCE(v.email, '') as email
FROM user_accounts ua
LEFT JOIN voters v ON v.id = ua.voter_id
WHERE ua.role = 'TPS_OPERATOR'
  AND ua.tps_id = $1
ORDER BY ua.username
`
	rows, err := r.db.Query(ctx, q, tpsID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TPSOperatorDTO
	for rows.Next() {
		var o TPSOperatorDTO
		if err := rows.Scan(&o.UserID, &o.Username, &o.Name, &o.Email); err != nil {
			return nil, err
		}
		items = append(items, o)
	}
	return items, rows.Err()
}

// CreateOperator creates a new TPS operator
func (r *PgAdminRepository) CreateOperator(
	ctx context.Context,
	tpsID int64,
	username, password, name, email string,
) (*TPSOperatorDTO, error) {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	passwordHash := string(hashBytes)

	const q = `
INSERT INTO user_accounts (
    username,
    password_hash,
    role,
    tps_id,
    is_active
) VALUES ($1, $2, 'TPS_OPERATOR', $3, TRUE)
RETURNING id, username
`
	var dto TPSOperatorDTO
	err = r.db.QueryRow(ctx, q,
		username,
		passwordHash,
		tpsID,
	).Scan(&dto.UserID, &dto.Username)
	if err != nil {
		return nil, err
	}

	dto.Name = name
	dto.Email = email

	return &dto, nil
}

// RemoveOperator removes an operator from TPS
func (r *PgAdminRepository) RemoveOperator(ctx context.Context, tpsID, userID int64) error {
	const q = `
DELETE FROM user_accounts
WHERE id = $1 AND tps_id = $2 AND role = 'TPS_OPERATOR'
`
	_, err := r.db.Exec(ctx, q, userID, tpsID)
	return err
}

// ListMonitorForElection returns monitoring data for all TPS in an election
func (r *PgAdminRepository) ListMonitorForElection(
	ctx context.Context,
	electionID int64,
) ([]TPSMonitorDTO, error) {
	const q = `
SELECT
    t.id,
    t.code,
    t.name,
    t.location,
    COALESCE(stats.total_checkins, 0)    AS total_checkins,
    COALESCE(stats.approved_checkins, 0) AS approved_checkins,
    COALESCE(stats.total_votes, 0)       AS total_votes,
    stats.last_activity_at
FROM tps t
LEFT JOIN (
    SELECT
        tc.tps_id,
        COUNT(*) FILTER (WHERE tc.status IS NOT NULL)         AS total_checkins,
        COUNT(*) FILTER (WHERE tc.status = 'APPROVED')        AS approved_checkins,
        COUNT(v.id)                                           AS total_votes,
        GREATEST(
            MAX(tc.scan_at),
            MAX(tc.approved_at),
            MAX(v.created_at)
        ) AS last_activity_at
    FROM tps t2
    LEFT JOIN tps_checkins tc
        ON tc.tps_id = t2.id
        AND tc.election_id = $1
    LEFT JOIN votes v
        ON v.tps_id = t2.id
        AND v.election_id = $1
        AND v.channel = 'TPS'
    GROUP BY tc.tps_id
) stats
  ON stats.tps_id = t.id
WHERE t.election_id = $1
ORDER BY t.code
`
	rows, err := r.db.Query(ctx, q, electionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []TPSMonitorDTO
	for rows.Next() {
		var m TPSMonitorDTO
		if err := rows.Scan(
			&m.TPSID,
			&m.Code,
			&m.Name,
			&m.Location,
			&m.TotalCheckins,
			&m.ApprovedCheckins,
			&m.TotalVotes,
			&m.LastActivityAt,
		); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// GetTPSQRMetadata returns QR metadata for TPS
func (r *PgAdminRepository) GetTPSQRMetadata(ctx context.Context, tpsID int64) (*TPSQRMetadataResponse, error) {
const q = `
SELECT 
    t.id,
    t.code,
    t.name,
    qr.id,
    qr.qr_token,
    qr.created_at
FROM tps t
LEFT JOIN tps_qr qr ON qr.tps_id = t.id AND qr.is_active = TRUE
WHERE t.id = $1
`
var resp TPSQRMetadataResponse
var qrID *int64
var qrToken *string
var qrCreatedAt *interface{}

err := r.db.QueryRow(ctx, q, tpsID).Scan(
&resp.TPSID,
&resp.Code,
&resp.Name,
&qrID,
&qrToken,
&qrCreatedAt,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrTPSNotFound
}
return nil, err
}

if qrID != nil && qrToken != nil {
resp.ActiveQR = &ActiveQRDTO{
ID:      *qrID,
QRToken: *qrToken,
}
}

return &resp, nil
}

// generateQRToken generates a secure random token for QR
func generateQRToken(tpsID int64) (string, error) {
b := make([]byte, 32)
if _, err := rand.Read(b); err != nil {
return "", err
}
token := fmt.Sprintf("tps_qr_%d_%s", tpsID, base64.URLEncoding.EncodeToString(b)[:32])
return token, nil
}

// RotateTPSQR rotates (generates new) QR for TPS
func (r *PgAdminRepository) RotateTPSQR(ctx context.Context, tpsID int64) (*TPSQRRotateResponse, error) {
tx, err := r.db.Begin(ctx)
if err != nil {
return nil, err
}
defer tx.Rollback(ctx)

// Verify TPS exists and get details
const qTPS = `SELECT code, name FROM tps WHERE id = $1`
var code, name string
if err := tx.QueryRow(ctx, qTPS, tpsID).Scan(&code, &name); err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrTPSNotFound
}
return nil, err
}

// Deactivate existing QR
const qDeactivate = `
UPDATE tps_qr 
SET is_active = FALSE, rotated_at = NOW() 
WHERE tps_id = $1 AND is_active = TRUE
`
if _, err := tx.Exec(ctx, qDeactivate, tpsID); err != nil {
return nil, err
}

// Generate new token
token, err := generateQRToken(tpsID)
if err != nil {
return nil, err
}

// Insert new QR
const qInsert = `
INSERT INTO tps_qr (tps_id, qr_token, is_active)
VALUES ($1, $2, TRUE)
RETURNING id, created_at
`
var newQR ActiveQRDTO
newQR.QRToken = token
if err := tx.QueryRow(ctx, qInsert, tpsID, token).Scan(&newQR.ID, &newQR.CreatedAt); err != nil {
return nil, err
}

if err := tx.Commit(ctx); err != nil {
return nil, err
}

return &TPSQRRotateResponse{
TPSID:    tpsID,
Code:     code,
Name:     name,
ActiveQR: newQR,
}, nil
}

// GetTPSQRForPrint returns QR payload for printing
func (r *PgAdminRepository) GetTPSQRForPrint(ctx context.Context, tpsID int64) (*TPSQRPrintResponse, error) {
const q = `
SELECT 
    t.id,
    t.code,
    t.name,
    qr.qr_token
FROM tps t
LEFT JOIN tps_qr qr ON qr.tps_id = t.id AND qr.is_active = TRUE
WHERE t.id = $1
`
var resp TPSQRPrintResponse
var qrToken *string

err := r.db.QueryRow(ctx, q, tpsID).Scan(
&resp.TPSID,
&resp.Code,
&resp.Name,
&qrToken,
)
if err != nil {
if errors.Is(err, pgx.ErrNoRows) {
return nil, ErrTPSNotFound
}
return nil, err
}

if qrToken == nil {
return nil, fmt.Errorf("no active QR found for TPS")
}

resp.QRPayload = fmt.Sprintf("pemira://tps-checkin?t=%s", *qrToken)
return &resp, nil
}
