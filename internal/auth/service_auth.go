package auth

import (
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInactiveUser        = errors.New("user account is inactive")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
)

type AuthService struct {
	repo       Repository
	jwtManager *JWTManager
	config     JWTConfig
}

func NewAuthService(repo Repository, jwtManager *JWTManager, config JWTConfig) *AuthService {
	return &AuthService{
		repo:       repo,
		jwtManager: jwtManager,
		config:     config,
	}
}

// Login authenticates a user and returns tokens
func (s *AuthService) Login(ctx context.Context, req LoginRequest, userAgent, ipAddress string) (*LoginResponse, error) {
	// Get user by username
	user, err := s.repo.GetUserByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	// Verify password
	if err := VerifyPassword(user.PasswordHash, req.Password); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generate access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate refresh token
	refreshToken, err := GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}

	// Hash refresh token
	refreshTokenHash, err := HashRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Create session
	ua := &userAgent
	ip := &ipAddress
	if userAgent == "" {
		ua = nil
	}
	if ipAddress == "" {
		ip = nil
	}

	session := &UserSession{
		UserID:           user.ID,
		RefreshTokenHash: refreshTokenHash,
		UserAgent:        ua,
		IPAddress:        ip,
		ExpiresAt:        time.Now().Add(s.config.RefreshTokenTTL),
	}

	_, err = s.repo.CreateSession(ctx, session)
	if err != nil {
		return nil, err
	}

	// Get user profile
	var profile *UserProfile
	if pgRepo, ok := s.repo.(*PgRepository); ok {
		profile, _ = pgRepo.GetUserProfile(ctx, user)
	}
	if profile == nil {
		profile = &UserProfile{}
	}

	authUser := &AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		VoterID:  user.VoterID,
		TPSID:    user.TPSID,
		Profile:  profile,
	}

	return &LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
		User:         authUser,
	}, nil
}

// RefreshToken generates new access and refresh tokens
func (s *AuthService) RefreshToken(ctx context.Context, req RefreshRequest) (*RefreshResponse, error) {
	// Hash the provided refresh token
	tokenHash, err := HashRefreshToken(req.RefreshToken)
	if err != nil {
		return nil, err
	}

	// Find session
	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil, ErrInvalidRefreshToken
		}
		return nil, err
	}

	// Check if session is revoked
	if session.RevokedAt != nil {
		return nil, ErrInvalidRefreshToken
	}

	// Check if session is expired
	if time.Now().After(session.ExpiresAt) {
		return nil, ErrInvalidRefreshToken
	}

	// Get user
	user, err := s.repo.GetUserByID(ctx, session.UserID)
	if err != nil {
		return nil, err
	}

	// Check if user is active
	if !user.IsActive {
		return nil, ErrInactiveUser
	}

	// Generate new access token
	accessToken, err := s.jwtManager.GenerateAccessToken(user)
	if err != nil {
		return nil, err
	}

	// Generate new refresh token
	newRefreshToken, err := GenerateRandomToken(32)
	if err != nil {
		return nil, err
	}

	// Hash new refresh token
	newRefreshTokenHash, err := HashRefreshToken(newRefreshToken)
	if err != nil {
		return nil, err
	}

	// Revoke old session
	if err := s.repo.RevokeSession(ctx, session.ID); err != nil {
		return nil, err
	}

	// Create new session
	newSession := &UserSession{
		UserID:           user.ID,
		RefreshTokenHash: newRefreshTokenHash,
		UserAgent:        session.UserAgent,
		IPAddress:        session.IPAddress,
		ExpiresAt:        time.Now().Add(s.config.RefreshTokenTTL),
	}

	_, err = s.repo.CreateSession(ctx, newSession)
	if err != nil {
		return nil, err
	}

	return &RefreshResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.config.AccessTokenTTL.Seconds()),
	}, nil
}

// Logout revokes a refresh token
func (s *AuthService) Logout(ctx context.Context, req LogoutRequest) error {
	// Hash the refresh token
	tokenHash, err := HashRefreshToken(req.RefreshToken)
	if err != nil {
		return err
	}

	// Find session
	session, err := s.repo.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, ErrSessionNotFound) {
			return nil // Already logged out
		}
		return err
	}

	// Revoke session
	return s.repo.RevokeSession(ctx, session.ID)
}

// GetCurrentUser returns user info from user ID
func (s *AuthService) GetCurrentUser(ctx context.Context, userID int64) (*AuthUser, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	var profile *UserProfile
	if pgRepo, ok := s.repo.(*PgRepository); ok {
		profile, _ = pgRepo.GetUserProfile(ctx, user)
	}
	if profile == nil {
		profile = &UserProfile{}
	}

	return &AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		VoterID:  user.VoterID,
		TPSID:    user.TPSID,
		Profile:  profile,
	}, nil
}
