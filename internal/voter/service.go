package voter

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	
	"pemira-api/internal/shared"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	repo Repository
	authRepo AuthRepository
}

func NewService(repo Repository, authRepo AuthRepository) *Service {
	return &Service{
		repo: repo,
		authRepo: authRepo,
	}
}

func (s *Service) GetByNIM(ctx context.Context, nim string) (*Voter, error) {
	return s.repo.GetByNIM(ctx, nim)
}

func (s *Service) List(ctx context.Context, params shared.PaginationParams) ([]*Voter, int64, error) {
	return s.repo.List(ctx, params)
}

func (s *Service) GetVoterStatus(ctx context.Context, voterID, electionID int64) (*VoterElectionStatus, error) {
	return s.repo.GetElectionStatus(ctx, voterID, electionID)
}

// Profile methods

func (s *Service) GetCompleteProfile(ctx context.Context, voterID int64, userID int64) (*CompleteProfileResponse, error) {
	return s.repo.GetCompleteProfile(ctx, voterID, userID)
}

func (s *Service) UpdateProfile(ctx context.Context, voterID int64, req *UpdateProfileRequest) error {
	// Validate email format
	if req.Email != nil && *req.Email != "" {
		emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
		if !emailRegex.MatchString(*req.Email) {
			return ErrInvalidEmail
		}
	}

	// Validate phone format (Indonesian format)
	if req.Phone != nil && *req.Phone != "" {
		phoneRegex := regexp.MustCompile(`^(08|\+62)[0-9]{8,13}$`)
		if !phoneRegex.MatchString(*req.Phone) {
			return ErrInvalidPhone
		}
	}

	return s.repo.UpdateProfile(ctx, voterID, req)
}

func (s *Service) UpdateVotingMethod(ctx context.Context, voterID, electionID int64, method string) error {
	if method != "ONLINE" && method != "TPS" {
		return ErrInvalidVotingMethod
	}

	return s.repo.UpdateVotingMethod(ctx, voterID, electionID, method)
}

func (s *Service) ChangePassword(ctx context.Context, userID int64, req *ChangePasswordRequest) error {
	// Validate password match
	if req.NewPassword != req.ConfirmPassword {
		return ErrPasswordMismatch
	}

	// Validate password length
	if len(req.NewPassword) < 8 {
		return ErrPasswordTooShort
	}

	// Check if new password is same as current
	if req.CurrentPassword == req.NewPassword {
		return ErrPasswordSameAsCurrent
	}

	// Verify current password
	user, err := s.authRepo.GetUserByID(ctx, userID)
	if err != nil {
		return err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword))
	if err != nil {
		return ErrInvalidCurrentPassword
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Update password
	return s.authRepo.UpdatePassword(ctx, userID, string(hashedPassword))
}

func (s *Service) GetParticipationStats(ctx context.Context, voterID int64) (*ParticipationStatsResponse, error) {
	return s.repo.GetParticipationStats(ctx, voterID)
}

func (s *Service) DeletePhoto(ctx context.Context, voterID int64) error {
	return s.repo.DeletePhoto(ctx, voterID)
}

var (
	ErrInvalidEmail           = errors.New("invalid email format")
	ErrInvalidPhone           = errors.New("invalid phone format (use 08xxx or +62xxx)")
	ErrInvalidVotingMethod    = errors.New("invalid voting method (must be ONLINE or TPS)")
	ErrPasswordMismatch       = errors.New("password confirmation does not match")
	ErrPasswordTooShort       = errors.New("password must be at least 8 characters")
	ErrPasswordSameAsCurrent  = errors.New("new password cannot be same as current password")
	ErrInvalidCurrentPassword = errors.New("current password is incorrect")
)
