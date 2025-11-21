package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"pemira-api/internal/shared/constants"
)

var (
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInactiveUser        = errors.New("user account is inactive")
	ErrInvalidRefreshToken = errors.New("invalid or expired refresh token")
	ErrInvalidRegisterType = errors.New("invalid registration type")
	ErrInvalidRegistration = errors.New("invalid registration data")
	ErrModeNotAvailable    = errors.New("voting mode not available")
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

// RegisterStudent registers a new student account and linked voter profile.
func (s *AuthService) RegisterStudent(ctx context.Context, req RegisterStudentRequest) (*AuthUser, error) {
	nim := strings.TrimSpace(req.NIM)
	name := strings.TrimSpace(req.Name)
	if nim == "" || name == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidRegistration
	}
	if len(req.Password) < 6 {
		return nil, ErrInvalidRegistration
	}

	email := strings.TrimSpace(req.Email)
	if email == "" {
		email = fmt.Sprintf("%s@pemira.ac.id", nim)
	}

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	voterID, err := s.repo.CreateVoter(ctx, VoterRegistration{
		NIM:              nim,
		Name:             name,
		Email:            email,
		FacultyName:      req.FacultyName,
		StudyProgramName: req.StudyProgramName,
		CohortYear:       req.CohortYear,
	})
	if err != nil {
		return nil, err
	}

	// Determine registration election & mode (default ONLINE)
	mode := normalizeVotingMode(req.VotingMode)
	regElection, err := s.repo.FindOrCreateRegistrationElection(ctx)
	if err != nil {
		return nil, err
	}
	if mode == "ONLINE" && !regElection.OnlineEnabled {
		return nil, ErrModeNotAvailable
	}
	if mode == "TPS" && !regElection.TPSEnabled {
		return nil, ErrModeNotAvailable
	}

	user, err := s.repo.CreateUserAccount(ctx, &UserAccount{
		Username:     nim,
		Email:        email,
		FullName:     name,
		PasswordHash: passwordHash,
		Role:         constants.RoleStudent,
		VoterID:      &voterID,
		IsActive:     true,
	})
	if err != nil {
		_ = s.repo.DeleteVoter(ctx, voterID)
		return nil, err
	}

	// Upsert voter_status preference/allowed flags
	onlineAllowed := mode == "ONLINE"
	tpsAllowed := mode == "TPS"
	_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

	profile := &UserProfile{
		Name:             name,
		FacultyName:      req.FacultyName,
		StudyProgramName: req.StudyProgramName,
		CohortYear:       req.CohortYear,
	}

	return &AuthUser{
		ID:       user.ID,
		Username: user.Username,
		Role:     user.Role,
		VoterID:  user.VoterID,
		Profile:  profile,
		// voting_mode echoed for client convenience
	}, nil
}

// RegisterLecturerStaff registers a lecturer or staff account.
func (s *AuthService) RegisterLecturerStaff(ctx context.Context, req RegisterLecturerStaffRequest) (*AuthUser, error) {
	roleType := strings.ToUpper(strings.TrimSpace(req.Type))
	if strings.TrimSpace(req.Name) == "" || strings.TrimSpace(req.Password) == "" {
		return nil, ErrInvalidRegistration
	}
	if len(req.Password) < 6 {
		return nil, ErrInvalidRegistration
	}

	email := strings.TrimSpace(req.Email)

	passwordHash, err := HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	mode := normalizeVotingMode(req.VotingMode)
	regElection, err := s.repo.FindOrCreateRegistrationElection(ctx)
	if err != nil {
		return nil, err
	}
	if mode == "ONLINE" && !regElection.OnlineEnabled {
		return nil, ErrModeNotAvailable
	}
	if mode == "TPS" && !regElection.TPSEnabled {
		return nil, ErrModeNotAvailable
	}

	switch roleType {
	case "LECTURER":
		nidn := strings.TrimSpace(req.NIDN)
		if nidn == "" {
			return nil, ErrInvalidRegistration
		}
		if email == "" {
			email = fmt.Sprintf("%s@pemira.ac.id", nidn)
		}

		lecturerID, err := s.repo.CreateLecturer(ctx, LecturerRegistration{
			NIDN:           nidn,
			Name:           req.Name,
			Email:          email,
			FacultyName:    req.FacultyName,
			DepartmentName: req.DepartmentName,
			Position:       req.Position,
		})
		if err != nil {
			return nil, err
		}

		voterID, err := s.repo.CreateVoter(ctx, VoterRegistration{
			NIM:              nidn,
			Name:             req.Name,
			Email:            email,
			FacultyName:      req.FacultyName,
			StudyProgramName: req.DepartmentName,
		})
		if err != nil {
			_ = s.repo.DeleteLecturer(ctx, lecturerID)
			return nil, err
		}

		user, err := s.repo.CreateUserAccount(ctx, &UserAccount{
			Username:     nidn,
			Email:        email,
			FullName:     req.Name,
			PasswordHash: passwordHash,
			Role:         constants.RoleLecturer,
			LecturerID:   &lecturerID,
			VoterID:      &voterID,
			IsActive:     true,
		})
		if err != nil {
			_ = s.repo.DeleteLecturer(ctx, lecturerID)
			_ = s.repo.DeleteVoter(ctx, voterID)
			return nil, err
		}

		onlineAllowed := mode == "ONLINE"
		tpsAllowed := mode == "TPS"
		_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

		profile := &UserProfile{
			Name:           req.Name,
			FacultyName:    req.FacultyName,
			DepartmentName: req.DepartmentName,
			Position:       req.Position,
		}

		return &AuthUser{
			ID:         user.ID,
			Username:   user.Username,
			Role:       user.Role,
			VoterID:    user.VoterID,
			LecturerID: user.LecturerID,
			Profile:    profile,
		}, nil

	case "STAFF":
		nip := strings.TrimSpace(req.NIP)
		if nip == "" {
			return nil, ErrInvalidRegistration
		}
		if email == "" {
			email = fmt.Sprintf("%s@pemira.ac.id", nip)
		}

		staffID, err := s.repo.CreateStaff(ctx, StaffRegistration{
			NIP:      nip,
			Name:     req.Name,
			Email:    email,
			UnitName: req.UnitName,
			Position: req.Position,
			Status:   "ACTIVE",
		})
		if err != nil {
			return nil, err
		}

		voterID, err := s.repo.CreateVoter(ctx, VoterRegistration{
			NIM:         nip,
			Name:        req.Name,
			Email:       email,
			FacultyName: req.UnitName,
		})
		if err != nil {
			_ = s.repo.DeleteStaff(ctx, staffID)
			return nil, err
		}

		user, err := s.repo.CreateUserAccount(ctx, &UserAccount{
			Username:     nip,
			Email:        email,
			FullName:     req.Name,
			PasswordHash: passwordHash,
			Role:         constants.RoleStaff,
			StaffID:      &staffID,
			VoterID:      &voterID,
			IsActive:     true,
		})
		if err != nil {
			_ = s.repo.DeleteStaff(ctx, staffID)
			_ = s.repo.DeleteVoter(ctx, voterID)
			return nil, err
		}

		onlineAllowed := mode == "ONLINE"
		tpsAllowed := mode == "TPS"
		_ = s.repo.EnsureVoterStatus(ctx, regElection.ID, voterID, mode, onlineAllowed, tpsAllowed)

		profile := &UserProfile{
			Name:     req.Name,
			UnitName: req.UnitName,
			Position: req.Position,
		}

		return &AuthUser{
			ID:       user.ID,
			Username: user.Username,
			Role:     user.Role,
			VoterID:  user.VoterID,
			StaffID:  user.StaffID,
			Profile:  profile,
		}, nil

	default:
		return nil, ErrInvalidRegisterType
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
		ID:         user.ID,
		Username:   user.Username,
		Role:       user.Role,
		VoterID:    user.VoterID,
		TPSID:      user.TPSID,
		LecturerID: user.LecturerID,
		StaffID:    user.StaffID,
		Profile:    profile,
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

func normalizeVotingMode(input string) string {
	mode := strings.ToUpper(strings.TrimSpace(input))
	if mode == "TPS" {
		return "TPS"
	}
	return "ONLINE"
}
