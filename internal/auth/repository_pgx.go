package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"pemira-api/internal/shared/constants"
)

type PgRepository struct {
	db *pgxpool.Pool
}

func NewPgRepository(db *pgxpool.Pool) *PgRepository {
	return &PgRepository{db: db}
}

// CreateUserAccount creates a new user account
func (r *PgRepository) CreateUserAccount(ctx context.Context, user *UserAccount) (*UserAccount, error) {
	query := `
		INSERT INTO user_accounts (username, email, password_hash, full_name, role, voter_id, tps_id, lecturer_id, staff_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, username, email, password_hash, full_name, role, voter_id, tps_id, lecturer_id, staff_id, is_active, created_at, updated_at
	`

	var created UserAccount
	email := user.Email
	fullName := user.FullName
	if email == "" {
		email = user.Username + "@pemira.ac.id"
	}
	if fullName == "" {
		fullName = user.Username
	}
	err := r.db.QueryRow(ctx, query,
		user.Username,
		email,
		user.PasswordHash,
		fullName,
		user.Role,
		user.VoterID,
		user.TPSID,
		user.LecturerID,
		user.StaffID,
		user.IsActive,
	).Scan(
		&created.ID,
		&created.Username,
		&created.Email,
		&created.PasswordHash,
		&created.FullName,
		&created.Role,
		&created.VoterID,
		&created.TPSID,
		&created.LecturerID,
		&created.StaffID,
		&created.IsActive,
		&created.CreatedAt,
		&created.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return nil, ErrUsernameExists
		}
		return nil, err
	}

	return &created, nil
}

// GetUserByUsername retrieves a user by username
func (r *PgRepository) GetUserByUsername(ctx context.Context, username string) (*UserAccount, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, voter_id, tps_id, lecturer_id, staff_id, is_active, last_login_at, login_count, created_at, updated_at
		FROM user_accounts
		WHERE username = $1
	`

	var user UserAccount
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.VoterID,
		&user.TPSID,
		&user.LecturerID,
		&user.StaffID,
		&user.IsActive,
		&user.LastLoginAt,
		&user.LoginCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// GetUserByID retrieves a user by ID
func (r *PgRepository) GetUserByID(ctx context.Context, userID int64) (*UserAccount, error) {
	query := `
		SELECT id, username, email, password_hash, full_name, role, voter_id, tps_id, lecturer_id, staff_id, is_active, last_login_at, login_count, created_at, updated_at
		FROM user_accounts
		WHERE id = $1
	`

	var user UserAccount
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.FullName,
		&user.Role,
		&user.VoterID,
		&user.TPSID,
		&user.LecturerID,
		&user.StaffID,
		&user.IsActive,
		&user.LastLoginAt,
		&user.LoginCount,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}

	return &user, nil
}

// UpdateUserAccount updates a user account
func (r *PgRepository) UpdateUserAccount(ctx context.Context, user *UserAccount) error {
	query := `
		UPDATE user_accounts
		SET password_hash = $2, role = $3, voter_id = $4, tps_id = $5, lecturer_id = $6, staff_id = $7, is_active = $8, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.PasswordHash,
		user.Role,
		user.VoterID,
		user.TPSID,
		user.LecturerID,
		user.StaffID,
		user.IsActive,
	)

	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// DeactivateUser deactivates a user account
func (r *PgRepository) DeactivateUser(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_accounts
		SET is_active = false, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrUserNotFound
	}

	return nil
}

// UpdateLoginTracking updates last_login_at and increments login_count
func (r *PgRepository) UpdateLoginTracking(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_accounts
		SET last_login_at = NOW(), 
		    login_count = COALESCE(login_count, 0) + 1,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// CreateSession creates a new user session
func (r *PgRepository) CreateSession(ctx context.Context, session *UserSession) (*UserSession, error) {
	query := `
		INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, refresh_token_hash, user_agent, ip_address::text, created_at, expires_at, revoked_at
	`

	var created UserSession
	err := r.db.QueryRow(ctx, query,
		session.UserID,
		session.RefreshTokenHash,
		session.UserAgent,
		session.IPAddress,
		session.ExpiresAt,
	).Scan(
		&created.ID,
		&created.UserID,
		&created.RefreshTokenHash,
		&created.UserAgent,
		&created.IPAddress,
		&created.CreatedAt,
		&created.ExpiresAt,
		&created.RevokedAt,
	)

	if err != nil {
		return nil, err
	}

	return &created, nil
}

// GetSessionByTokenHash retrieves a session by refresh token hash
func (r *PgRepository) GetSessionByTokenHash(ctx context.Context, tokenHash string) (*UserSession, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address::text, created_at, expires_at, revoked_at
		FROM user_sessions
		WHERE refresh_token_hash = $1
	`

	var session UserSession
	err := r.db.QueryRow(ctx, query, tokenHash).Scan(
		&session.ID,
		&session.UserID,
		&session.RefreshTokenHash,
		&session.UserAgent,
		&session.IPAddress,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.RevokedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, err
	}

	return &session, nil
}

// GetUserSessions retrieves all sessions for a user
func (r *PgRepository) GetUserSessions(ctx context.Context, userID int64) ([]UserSession, error) {
	query := `
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address::text, created_at, expires_at, revoked_at
		FROM user_sessions
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []UserSession
	for rows.Next() {
		var session UserSession
		if err := rows.Scan(
			&session.ID,
			&session.UserID,
			&session.RefreshTokenHash,
			&session.UserAgent,
			&session.IPAddress,
			&session.CreatedAt,
			&session.ExpiresAt,
			&session.RevokedAt,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

// RevokeSession revokes a session
func (r *PgRepository) RevokeSession(ctx context.Context, sessionID int64) error {
	query := `
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE id = $1 AND revoked_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return ErrSessionNotFound
	}

	return nil
}

// RevokeAllUserSessions revokes all sessions for a user
func (r *PgRepository) RevokeAllUserSessions(ctx context.Context, userID int64) error {
	query := `
		UPDATE user_sessions
		SET revoked_at = NOW()
		WHERE user_id = $1 AND revoked_at IS NULL
	`

	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// CleanupExpiredSessions removes expired sessions
func (r *PgRepository) CleanupExpiredSessions(ctx context.Context) error {
	query := `
		DELETE FROM user_sessions
		WHERE expires_at < NOW() OR revoked_at IS NOT NULL AND revoked_at < NOW() - INTERVAL '30 days'
	`

	_, err := r.db.Exec(ctx, query)
	return err
}

// GetUserProfile retrieves user profile based on role
func (r *PgRepository) GetUserProfile(ctx context.Context, user *UserAccount) (*UserProfile, error) {
	profile := &UserProfile{}

	switch user.Role {
	case constants.RoleStudent:
		if user.VoterID == nil {
			return profile, nil
		}

		query := `
			SELECT name, faculty_name, study_program_name, cohort_year, class_label
			FROM voters
			WHERE id = $1
		`

		err := r.db.QueryRow(ctx, query, *user.VoterID).Scan(
			&profile.Name,
			&profile.FacultyName,
			&profile.StudyProgramName,
			&profile.CohortYear,
			&profile.Semester,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

	case constants.RoleLecturer:
		if user.LecturerID == nil {
			return profile, nil
		}

		query := `
			SELECT name, faculty_name, department_name, position
			FROM lecturers
			WHERE id = $1
		`

		err := r.db.QueryRow(ctx, query, *user.LecturerID).Scan(
			&profile.Name,
			&profile.FacultyName,
			&profile.DepartmentName,
			&profile.Position,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

	case constants.RoleStaff:
		if user.StaffID == nil {
			return profile, nil
		}

		query := `
			SELECT name, unit_name, position
			FROM staff_members
			WHERE id = $1
		`

		err := r.db.QueryRow(ctx, query, *user.StaffID).Scan(
			&profile.Name,
			&profile.UnitName,
			&profile.Position,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

	case constants.RoleTPSOperator:
		if user.TPSID == nil {
			return profile, nil
		}

		query := `
			SELECT code, name
			FROM tps
			WHERE id = $1
		`

		err := r.db.QueryRow(ctx, query, *user.TPSID).Scan(
			&profile.TPSCode,
			&profile.TPSName,
		)

		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, err
		}

	case constants.RoleAdmin, constants.RoleSuperAdmin:
		// Admin has no additional profile
	}

	return profile, nil
}

// CreateVoter inserts a new voter record and returns its ID.
func (r *PgRepository) CreateVoter(ctx context.Context, voter VoterRegistration) (int64, error) {
	query := `
		INSERT INTO voters (nim, name, email, faculty_name, study_program_name, cohort_year, class_label, academic_status)
		VALUES ($1, $2, $3, $4, $5, NULL, $6, 'ACTIVE')
		RETURNING id
	`

	var id int64
	classLabel := strings.TrimSpace(voter.Semester)
	if classLabel == "" {
		classLabel = "Semester tidak diisi"
	}
	if err := r.db.QueryRow(ctx, query,
		voter.NIM,
		voter.Name,
		voter.Email,
		voter.FacultyName,
		voter.StudyProgramName,
		classLabel,
	).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_voters_nim" {
			return 0, ErrNIMExists
		}
		return 0, err
	}

	return id, nil
}

// DeleteVoter removes a voter row by ID (used for cleanup on failure).
func (r *PgRepository) DeleteVoter(ctx context.Context, voterID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM voters WHERE id = $1`, voterID)
	return err
}

// CreateLecturer inserts a new lecturer record and returns its ID.
func (r *PgRepository) CreateLecturer(ctx context.Context, lecturer LecturerRegistration) (int64, error) {
	query := `
		INSERT INTO lecturers (nidn, name, email, faculty_name, department_name, position, unit_id, position_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRow(ctx, query,
		lecturer.NIDN,
		lecturer.Name,
		lecturer.Email,
		lecturer.FacultyName,
		lecturer.DepartmentName,
		lecturer.Position,
		lecturer.UnitID,
		lecturer.PositionID,
	).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_lecturers_nidn" {
			return 0, ErrNIDNExists
		}
		return 0, err
	}

	return id, nil
}

// DeleteLecturer removes a lecturer row by ID (used for cleanup on failure).
func (r *PgRepository) DeleteLecturer(ctx context.Context, lecturerID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM lecturers WHERE id = $1`, lecturerID)
	return err
}

// CreateStaff inserts a new staff record and returns its ID.
func (r *PgRepository) CreateStaff(ctx context.Context, staff StaffRegistration) (int64, error) {
	query := `
		INSERT INTO staff_members (nip, name, email, unit_name, position, employment_status, unit_id, position_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`

	var id int64
	if err := r.db.QueryRow(ctx, query,
		staff.NIP,
		staff.Name,
		staff.Email,
		staff.UnitName,
		staff.Position,
		staff.Status,
		staff.UnitID,
		staff.PositionID,
	).Scan(&id); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "ux_staff_members_nip" {
			return 0, ErrNIPExists
		}
		return 0, err
	}

	return id, nil
}

// DeleteStaff removes a staff row by ID (used for cleanup on failure).
func (r *PgRepository) DeleteStaff(ctx context.Context, staffID int64) error {
	_, err := r.db.Exec(ctx, `DELETE FROM staff_members WHERE id = $1`, staffID)
	return err
}

// FindOrCreateRegistrationElection finds the latest election with status not ARCHIVED (REGISTRATION/CAMPAIGN/VOTING_OPEN/CLOSED).
// If none exists, it will create a placeholder in REGISTRATION status with both channels enabled.
func (r *PgRepository) FindOrCreateRegistrationElection(ctx context.Context) (*RegistrationElection, error) {
	const findQuery = `
		SELECT id, status, online_enabled, tps_enabled
		FROM elections
		WHERE status IN ('REGISTRATION','CAMPAIGN','VOTING_OPEN','CLOSED')
		ORDER BY created_at DESC
		LIMIT 1
	`
	var e RegistrationElection
	err := r.db.QueryRow(ctx, findQuery).Scan(&e.ID, &e.Status, &e.OnlineEnabled, &e.TPSEnabled)
	if err == nil {
		return &e, nil
	}

	// Create placeholder election if not found
	code := fmt.Sprintf("AUTO-%d", time.Now().Unix())
	year := time.Now().Year()
	createQuery := `
		INSERT INTO elections (code, name, year, status, online_enabled, tps_enabled, created_at, updated_at)
		VALUES ($1, $2, $3, 'REGISTRATION', TRUE, TRUE, NOW(), NOW())
		RETURNING id, status, online_enabled, tps_enabled
	`
	err = r.db.QueryRow(ctx, createQuery, code, "Pemira Auto", year).Scan(&e.ID, &e.Status, &e.OnlineEnabled, &e.TPSEnabled)
	if err != nil {
		return nil, fmt.Errorf("create placeholder election: %w", err)
	}
	return &e, nil
}

// EnsureVoterStatus upserts preferred method and allowed flags for a voter in an election.
func (r *PgRepository) EnsureVoterStatus(ctx context.Context, electionID, voterID int64, preferredMethod string, onlineAllowed, tpsAllowed bool) error {
	query := `
		INSERT INTO voter_status (election_id, voter_id, is_eligible, has_voted, preferred_method, online_allowed, tps_allowed)
		VALUES ($1,$2,TRUE,FALSE,$3,$4,$5)
		ON CONFLICT (election_id, voter_id)
		DO UPDATE SET preferred_method = EXCLUDED.preferred_method,
		              online_allowed = EXCLUDED.online_allowed,
		              tps_allowed = EXCLUDED.tps_allowed,
		              updated_at = NOW()
	`
	_, err := r.db.Exec(ctx, query, electionID, voterID, preferredMethod, onlineAllowed, tpsAllowed)
	return err
}
