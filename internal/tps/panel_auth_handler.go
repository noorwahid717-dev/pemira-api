package tps

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"pemira-api/internal/auth"
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/constants"
)

// PanelAuthHandler handles TPS panel authentication (login).
type PanelAuthHandler struct {
	authService *auth.AuthService
	tpsRepo     Repository
}

func NewPanelAuthHandler(authSvc *auth.AuthService, repo Repository) *PanelAuthHandler {
	return &PanelAuthHandler{
		authService: authSvc,
		tpsRepo:     repo,
	}
}

// PanelLogin handles POST /tps-panel/auth/login
func (h *PanelAuthHandler) PanelLogin(w http.ResponseWriter, r *http.Request) {
	var req auth.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || strings.TrimSpace(req.Password) == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "username dan password wajib diisi.")
		return
	}

	userAgent := r.Header.Get("User-Agent")
	ipAddress := r.Header.Get("X-Real-IP")
	if ipAddress == "" {
		ipAddress = r.Header.Get("X-Forwarded-For")
		if ipAddress != "" {
			ipAddress = strings.Split(ipAddress, ",")[0]
		}
	}
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}

	loginResp, err := h.authService.Login(r.Context(), req, userAgent, ipAddress)
	if err != nil {
		h.handleError(w, err)
		return
	}

	if loginResp.User.Role != constants.RoleTPSOperator {
		response.Error(w, http.StatusForbidden, "NOT_TPS_OPERATOR", "Akun ini bukan operator TPS.", nil)
		return
	}

	if loginResp.User.TPSID == nil {
		response.Error(w, http.StatusForbidden, "TPS_NOT_ASSIGNED", "Akun operator belum terhubung ke TPS.", nil)
		return
	}

	tpsRow, err := h.tpsRepo.GetByID(r.Context(), *loginResp.User.TPSID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, "TPS tidak ditemukan.", nil)
		return
	}
	if tpsRow.Status != StatusActive {
		response.Error(w, http.StatusBadRequest, "TPS_INACTIVE", "TPS belum aktif.", nil)
		return
	}

	resp := map[string]interface{}{
		"access_token":  loginResp.AccessToken,
		"refresh_token": loginResp.RefreshToken,
		"token_type":    loginResp.TokenType,
		"expires_in":    loginResp.ExpiresIn,
		"operator": map[string]interface{}{
			"id":       loginResp.User.ID,
			"name":     strings.TrimSpace(loginResp.User.Profile.Name),
			"username": loginResp.User.Username,
			"tps": map[string]interface{}{
				"id":       tpsRow.ID,
				"code":     tpsRow.Code,
				"name":     tpsRow.Name,
				"location": tpsRow.Location,
			},
		},
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *PanelAuthHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, auth.ErrInvalidCredentials):
		response.Error(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Username atau password salah.", nil)
	case errors.Is(err, auth.ErrInactiveUser):
		response.Error(w, http.StatusForbidden, "USER_INACTIVE", "Akun tidak aktif.", nil)
	default:
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal login.")
	}
}
