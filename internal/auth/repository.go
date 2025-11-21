package auth

import (
	"context"
	"errors"
)

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrSessionNotFound     = errors.New("session not found")
	ErrUsernameExists      = errors.New("username already exists")
	ErrNIMExists           = errors.New("nim already exists")
	ErrNIDNExists          = errors.New("nidn already exists")
	ErrNIPExists           = errors.New("nip already exists")
	ErrElectionUnavailable = errors.New("no active election for registration")
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

	// Registration helpers
	CreateVoter(ctx context.Context, voter VoterRegistration) (int64, error)
	DeleteVoter(ctx context.Context, voterID int64) error
	CreateLecturer(ctx context.Context, lecturer LecturerRegistration) (int64, error)
	DeleteLecturer(ctx context.Context, lecturerID int64) error
	CreateStaff(ctx context.Context, staff StaffRegistration) (int64, error)
	DeleteStaff(ctx context.Context, staffID int64) error

	// Election helpers for registration
	FindOrCreateRegistrationElection(ctx context.Context) (*RegistrationElection, error)

	// Voter status helpers
	EnsureVoterStatus(ctx context.Context, electionID, voterID int64, preferredMethod string, onlineAllowed, tpsAllowed bool) error
}
