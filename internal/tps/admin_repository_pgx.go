package tps

import (
	"context"
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
    id,
    code,
    name,
    location,
    capacity_estimate,
    CASE WHEN status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    created_at,
    updated_at
FROM tps
ORDER BY code
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
    id,
    code,
    name,
    location,
    capacity_estimate,
    CASE WHEN status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    created_at,
    updated_at
FROM tps
WHERE id = $1
`
	var t TPSDTO
	err := r.db.QueryRow(ctx, q, id).Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Capacity,
		&t.IsActive,
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
	const q = `
INSERT INTO tps (
    election_id,
    code,
    name,
    location,
    capacity_estimate,
    status,
    open_time,
    close_time
) VALUES (1, $1, $2, $3, $4, 'ACTIVE', '08:00', '17:00')
RETURNING 
    id,
    code,
    name,
    location,
    capacity_estimate,
    CASE WHEN status = 'ACTIVE' THEN TRUE ELSE FALSE END as is_active,
    created_at,
    updated_at
`
	var t TPSDTO
	err := r.db.QueryRow(ctx, q,
		req.Code,
		req.Name,
		req.Location,
		req.Capacity,
	).Scan(
		&t.ID,
		&t.Code,
		&t.Name,
		&t.Location,
		&t.Capacity,
		&t.IsActive,
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
