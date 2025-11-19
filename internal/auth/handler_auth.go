package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

type AuthHandler struct {
	service *AuthService
}

func NewAuthHandler(service *AuthService) *AuthHandler {
	return &AuthHandler{service: service}
}

// Login handles POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body.", nil)
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		response.UnprocessableEntity(w, "Username and password are required.", nil)
		return
	}

	// Extract user agent and IP
	userAgent := r.Header.Get("User-Agent")
	ipAddress := r.Header.Get("X-Real-IP")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
		if ipAddress != "" {
			// Take first IP if multiple
			ipAddress = strings.Split(ipAddress, ",")[0]
		}
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	loginResp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, loginResp)
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body.", nil)
		return
	}

	if req.RefreshToken == "" {
		response.UnprocessableEntity(w, "Refresh token is required.", nil)
		return
	}

	refreshResp, err := h.service.RefreshToken(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, refreshResp)
}

// Logout handles POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body.", nil)
		return
	}

	if req.RefreshToken == "" {
		response.UnprocessableEntity(w, "Refresh token is required.", nil)
		return
	}

	if err := h.service.Logout(r.Context(), req); err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Logged out successfully.",
	})
}

// Me handles GET /auth/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Get user ID from context (set by auth middleware)
	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "Unauthorized.")
		return
	}

	authUser, err := h.service.GetCurrentUser(r.Context(), userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, authUser)
}

// handleError maps service errors to HTTP responses
func (h *AuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(w, "Invalid username or password.")

	case errors.Is(err, ErrInactiveUser):
		response.Forbidden(w, "Your account is inactive. Please contact administrator.")

	case errors.Is(err, ErrInvalidRefreshToken):
		response.Unauthorized(w, "Invalid or expired refresh token.")

	case errors.Is(err, ErrUserNotFound):
		response.NotFound(w, "User not found.")

	default:
		// Log internal error
		response.InternalServerError(w, "An error occurred. Please try again.")
	}
}
