package voter

import (
	"context"
	
	"github.com/jackc/pgx/v5/pgxpool"
)

// AuthRepositoryAdapter adapts the voter.AuthRepository interface to use pgx
type AuthRepositoryAdapter struct {
	db *pgxpool.Pool
}

func NewAuthRepositoryAdapter(db *pgxpool.Pool) *AuthRepositoryAdapter {
	return &AuthRepositoryAdapter{db: db}
}

func (a *AuthRepositoryAdapter) GetUserByID(ctx context.Context, userID int64) (*User, error) {
	query := `
		SELECT id, username, password_hash
		FROM user_accounts
		WHERE id = $1
	`

	var user User
	err := a.db.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (a *AuthRepositoryAdapter) UpdatePassword(ctx context.Context, userID int64, hashedPassword string) error {
	query := `
		UPDATE user_accounts
		SET password_hash = $2, updated_at = NOW()
		WHERE id = $1
	`

	_, err := a.db.Exec(ctx, query, userID, hashedPassword)
	return err
}
