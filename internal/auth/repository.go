package auth

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound     = errors.New("user not found")
	ErrSessionNotFound  = errors.New("session not found")
	ErrUsernameExists   = errors.New("username already exists")
)

type Repository interface {
	// User account operations
	CreateUserAccount(ctx context.Context, user *UserAccount) (*UserAccount, error)
	GetUserByUsername(ctx context.Context, username string) (*UserAccount, error)
	GetUserByID(ctx context.Context, userID int64) (*UserAccount, error)
	UpdateUserAccount(ctx context.Context, user *UserAccount) error
	DeactivateUser(ctx context.Context, userID int64) error

	// Session operations
	CreateSession(ctx context.Context, session *UserSession) (*UserSession, error)
	GetSessionByTokenHash(ctx context.Context, tokenHash string) (*UserSession, error)
	GetUserSessions(ctx context.Context, userID int64) ([]UserSession, error)
	RevokeSession(ctx context.Context, sessionID int64) error
	RevokeAllUserSessions(ctx context.Context, userID int64) error
	CleanupExpiredSessions(ctx context.Context) error
}
