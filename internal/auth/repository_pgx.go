package auth

import (
	"context"
	"errors"

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
		INSERT INTO user_accounts (username, password_hash, role, voter_id, tps_id, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, username, password_hash, role, voter_id, tps_id, is_active, created_at, updated_at
	`

	var created UserAccount
	err := r.db.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.VoterID,
		user.TPSID,
		user.IsActive,
	).Scan(
		&created.ID,
		&created.Username,
		&created.PasswordHash,
		&created.Role,
		&created.VoterID,
		&created.TPSID,
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
		SELECT id, username, password_hash, role, voter_id, tps_id, is_active, created_at, updated_at
		FROM user_accounts
		WHERE username = $1
	`

	var user UserAccount
	err := r.db.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.VoterID,
		&user.TPSID,
		&user.IsActive,
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
		SELECT id, username, password_hash, role, voter_id, tps_id, is_active, created_at, updated_at
		FROM user_accounts
		WHERE id = $1
	`

	var user UserAccount
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.VoterID,
		&user.TPSID,
		&user.IsActive,
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
		SET password_hash = $2, role = $3, voter_id = $4, tps_id = $5, 
		    is_active = $6, updated_at = NOW()
		WHERE id = $1
	`

	result, err := r.db.Exec(ctx, query,
		user.ID,
		user.PasswordHash,
		user.Role,
		user.VoterID,
		user.TPSID,
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

// CreateSession creates a new user session
func (r *PgRepository) CreateSession(ctx context.Context, session *UserSession) (*UserSession, error) {
	query := `
		INSERT INTO user_sessions (user_id, refresh_token_hash, user_agent, ip_address, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, refresh_token_hash, user_agent, ip_address, created_at, expires_at, revoked_at
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
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, created_at, expires_at, revoked_at
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
		SELECT id, user_id, refresh_token_hash, user_agent, ip_address, created_at, expires_at, revoked_at
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
			SELECT name, faculty_name, study_program_name, cohort_year
			FROM voters
			WHERE id = $1
		`

		err := r.db.QueryRow(ctx, query, *user.VoterID).Scan(
			&profile.Name,
			&profile.FacultyName,
			&profile.StudyProgramName,
			&profile.CohortYear,
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
