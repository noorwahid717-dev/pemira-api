package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"pemira-api/internal/shared/constants"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type JWTConfig struct {
	Secret           string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
}

type JWTManager struct {
	config JWTConfig
}

func NewJWTManager(config JWTConfig) *JWTManager {
	return &JWTManager{config: config}
}

// GenerateAccessToken generates a new JWT access token
func (j *JWTManager) GenerateAccessToken(user *UserAccount) (string, error) {
	now := time.Now()
	expiresAt := now.Add(j.config.AccessTokenTTL)

	claims := jwt.MapClaims{
		"sub":    user.ID,
		"role":   string(user.Role),
		"iat":    now.Unix(),
		"exp":    expiresAt.Unix(),
	}

	// Add optional claims
	if user.VoterID != nil {
		claims["voter_id"] = *user.VoterID
	}
	if user.TPSID != nil {
		claims["tps_id"] = *user.TPSID
	}
	if user.LecturerID != nil {
		claims["lecturer_id"] = *user.LecturerID
	}
	if user.StaffID != nil {
		claims["staff_id"] = *user.StaffID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.config.Secret))
}

// ValidateAccessToken validates and parses JWT access token
func (j *JWTManager) ValidateAccessToken(tokenString string) (*JWTClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return []byte(j.config.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	if !token.Valid {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, ErrInvalidToken
	}

	// Extract claims
	userID, ok := claims["sub"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	role, ok := claims["role"].(string)
	if !ok {
		return nil, ErrInvalidToken
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	iat, ok := claims["iat"].(float64)
	if !ok {
		return nil, ErrInvalidToken
	}

	jwtClaims := &JWTClaims{
		UserID: int64(userID),
		Role:   constants.Role(role),
		Exp:    int64(exp),
		Iat:    int64(iat),
	}

	// Extract optional claims
	if voterID, ok := claims["voter_id"].(float64); ok {
		vid := int64(voterID)
		jwtClaims.VoterID = &vid
	}

	if tpsID, ok := claims["tps_id"].(float64); ok {
		tid := int64(tpsID)
		jwtClaims.TPSID = &tid
	}

	if lecturerID, ok := claims["lecturer_id"].(float64); ok {
		lid := int64(lecturerID)
		jwtClaims.LecturerID = &lid
	}

	if staffID, ok := claims["staff_id"].(float64); ok {
		sid := int64(staffID)
		jwtClaims.StaffID = &sid
	}

	return jwtClaims, nil
}
