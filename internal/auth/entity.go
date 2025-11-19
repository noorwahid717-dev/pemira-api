package auth

import (
	"time"
	"pemira-api/internal/shared/constants"
)

type User struct {
	ID           int64          `json:"id"`
	Username     string         `json:"username"`
	PasswordHash string         `json:"-"`
	Role         constants.Role `json:"role"`
	FullName     string         `json:"full_name"`
	Email        string         `json:"email"`
	IsActive     bool           `json:"is_active"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
}

type Session struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Token     string    `json:"token"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}
