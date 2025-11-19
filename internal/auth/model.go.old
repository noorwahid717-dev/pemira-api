package auth

import (
	"time"

	"pemira-api/internal/shared/constants"
)

// UserAccount represents a user account in the system
type UserAccount struct {
	ID           int64          `json:"id"`
	Username     string         `json:"username"`
	PasswordHash string         `json:"-"`
	Role         constants.Role `json:"role"`
	VoterID      *int64         `json:"voter_id,omitempty"`
	TPSID        *int64         `json:"tps_id,omitempty"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// UserSession represents a user session with refresh token
type UserSession struct {
	ID                int64      `json:"id"`
	UserID            int64      `json:"user_id"`
	RefreshTokenHash  string     `json:"-"`
	UserAgent         *string    `json:"user_agent,omitempty"`
	IPAddress         *string    `json:"ip_address,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	ExpiresAt         time.Time  `json:"expires_at"`
	RevokedAt         *time.Time `json:"revoked_at,omitempty"`
}

// UserProfile represents user profile info (varies by role)
type UserProfile struct {
	Name              string `json:"name,omitempty"`
	FacultyName       string `json:"faculty_name,omitempty"`
	StudyProgramName  string `json:"study_program_name,omitempty"`
	CohortYear        *int   `json:"cohort_year,omitempty"`
	TPSCode           string `json:"tps_code,omitempty"`
	TPSName           string `json:"tps_name,omitempty"`
}

// AuthUser represents authenticated user with profile
type AuthUser struct {
	ID       int64          `json:"id"`
	Username string         `json:"username"`
	Role     constants.Role `json:"role"`
	VoterID  *int64         `json:"voter_id,omitempty"`
	TPSID    *int64         `json:"tps_id,omitempty"`
	Profile  *UserProfile   `json:"profile,omitempty"`
}

// JWTClaims represents JWT token claims
type JWTClaims struct {
	UserID  int64          `json:"sub"`
	Role    constants.Role `json:"role"`
	VoterID *int64         `json:"voter_id,omitempty"`
	TPSID   *int64         `json:"tps_id,omitempty"`
	Exp     int64          `json:"exp"`
	Iat     int64          `json:"iat"`
}
