package auth

import (
	"time"

	"pemira-api/internal/shared/constants"
)

// UserAccount represents an account that can access the system.
type UserAccount struct {
	ID           int64          `json:"id"`
	Username     string         `json:"username"`
	Email        string         `json:"email"`
	FullName     string         `json:"full_name"`
	PasswordHash string         `json:"-"`
	Role         constants.Role `json:"role"`
	VoterID      *int64         `json:"voter_id,omitempty"`
	TPSID        *int64         `json:"tps_id,omitempty"`
	LecturerID   *int64         `json:"lecturer_id,omitempty"`
	StaffID      *int64         `json:"staff_id,omitempty"`
	IsActive     bool           `json:"is_active"`
	LastLoginAt  *time.Time     `json:"last_login_at,omitempty"`
	LoginCount   int            `json:"login_count"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

// UserSession stores refresh token metadata.
type UserSession struct {
	ID               int64      `json:"id"`
	UserID           int64      `json:"user_id"`
	RefreshTokenHash string     `json:"-"`
	UserAgent        *string    `json:"user_agent,omitempty"`
	IPAddress        *string    `json:"ip_address,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
	ExpiresAt        time.Time  `json:"expires_at"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty"`
}

// RegistrationElection is a lightweight projection used during registration.
type RegistrationElection struct {
	ID            int64  `json:"id"`
	Status        string `json:"status"`
	OnlineEnabled bool   `json:"online_enabled"`
	TPSEnabled    bool   `json:"tps_enabled"`
}

// UserProfile captures optional profile data shown in auth responses.
type UserProfile struct {
	Name             string `json:"name,omitempty"`
	FacultyName      string `json:"faculty_name,omitempty"`
	StudyProgramName string `json:"study_program_name,omitempty"`
	CohortYear       *int   `json:"cohort_year,omitempty"`
	Semester         string `json:"semester,omitempty"`
	TPSCode          string `json:"tps_code,omitempty"`
	TPSName          string `json:"tps_name,omitempty"`
	// For lecturers
	DepartmentName string `json:"department_name,omitempty"`
	Position       string `json:"position,omitempty"`
	// For staff
	UnitName string `json:"unit_name,omitempty"`
}

// AuthUser represents an authenticated user returned to clients.
type AuthUser struct {
	ID         int64          `json:"id"`
	Username   string         `json:"username"`
	Role       constants.Role `json:"role"`
	VoterID    *int64         `json:"voter_id,omitempty"`
	TPSID      *int64         `json:"tps_id,omitempty"`
	LecturerID *int64         `json:"lecturer_id,omitempty"`
	StaffID    *int64         `json:"staff_id,omitempty"`
	Profile    *UserProfile   `json:"profile,omitempty"`
}

// JWTClaims is used by JWT middleware and tokens.
type JWTClaims struct {
	UserID     int64          `json:"sub"`
	Role       constants.Role `json:"role"`
	VoterID    *int64         `json:"voter_id,omitempty"`
	TPSID      *int64         `json:"tps_id,omitempty"`
	LecturerID *int64         `json:"lecturer_id,omitempty"`
	StaffID    *int64         `json:"staff_id,omitempty"`
	Exp        int64          `json:"exp"`
	Iat        int64          `json:"iat"`
}
