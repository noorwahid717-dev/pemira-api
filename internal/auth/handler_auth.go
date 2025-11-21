package auth

import (
	"encoding/json"
	"errors"
	"log/slog"
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
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	// Validate input
	if req.Username == "" || req.Password == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "username dan password wajib diisi.")
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

	// Remove port from IP address if present
	if idx := strings.LastIndex(ipAddress, ":"); idx != -1 {
		// Check if it's IPv6 or IPv4 with port
		if strings.Count(ipAddress, ":") == 1 || strings.HasPrefix(ipAddress, "[") {
			ipAddress = strings.TrimRight(strings.Split(ipAddress, ":")[0], "]")
			ipAddress = strings.TrimLeft(ipAddress, "[")
		}
	}

	loginResp, err := h.service.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, loginResp)
}

// RegisterStudent handles POST /auth/register/student
func (h *AuthHandler) RegisterStudent(w http.ResponseWriter, r *http.Request) {
	var req RegisterStudentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	user, err := h.service.RegisterStudent(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user":        user,
		"message":     "Registrasi mahasiswa berhasil.",
		"voting_mode": normalizeVotingMode(req.VotingMode),
	})
}

// RegisterLecturerStaff handles POST /auth/register/lecturer-staff
func (h *AuthHandler) RegisterLecturerStaff(w http.ResponseWriter, r *http.Request) {
	var req RegisterLecturerStaffRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	user, err := h.service.RegisterLecturerStaff(r.Context(), req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"user":        user,
		"message":     "Registrasi berhasil.",
		"voting_mode": normalizeVotingMode(req.VotingMode),
	})
}

// RefreshToken handles POST /auth/refresh
func (h *AuthHandler) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.RefreshToken == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "Refresh token wajib diisi.")
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
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.RefreshToken == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "Refresh token wajib diisi.")
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
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid atau tidak ditemukan.")
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
		response.Unauthorized(w, "INVALID_CREDENTIALS", "Username atau password salah.")

	case errors.Is(err, ErrInactiveUser):
		response.Forbidden(w, "USER_INACTIVE", "Akun tidak aktif.")

	case errors.Is(err, ErrInvalidRefreshToken):
		response.Unauthorized(w, "INVALID_REFRESH_TOKEN", "Refresh token tidak valid atau sudah kadaluarsa.")

	case errors.Is(err, ErrUserNotFound):
		response.NotFound(w, "USER_NOT_FOUND", "Pengguna tidak ditemukan.")

	case errors.Is(err, ErrUsernameExists):
		response.Conflict(w, "USERNAME_EXISTS", "Username sudah terdaftar.")

	case errors.Is(err, ErrNIMExists):
		response.Conflict(w, "NIM_EXISTS", "NIM sudah terdaftar.")

	case errors.Is(err, ErrNIDNExists):
		response.Conflict(w, "NIDN_EXISTS", "NIDN sudah terdaftar.")

	case errors.Is(err, ErrNIPExists):
		response.Conflict(w, "NIP_EXISTS", "NIP sudah terdaftar.")

	case errors.Is(err, ErrInvalidRegisterType):
		response.BadRequest(w, "INVALID_REQUEST", "Tipe registrasi tidak valid.")

	case errors.Is(err, ErrInvalidRegistration):
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "Data registrasi tidak lengkap atau tidak valid.")

	case errors.Is(err, ErrModeNotAvailable):
		response.UnprocessableEntity(w, "MODE_NOT_AVAILABLE", "Mode tidak tersedia untuk pemilu ini.")

	default:
		// Log internal error
		slog.Error("auth handler error", "error", err)
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}
