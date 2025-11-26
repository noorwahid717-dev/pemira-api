package voter

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

type ProfileHandler struct {
	service *Service
}

func NewProfileHandler(service *Service) *ProfileHandler {
	return &ProfileHandler{service: service}
}

func (h *ProfileHandler) RegisterRoutes(r chi.Router) {
	r.Get("/voters/me/complete-profile", h.GetCompleteProfile)
	r.Put("/voters/me/profile", h.UpdateProfile)
	r.Put("/voters/me/voting-method", h.UpdateVotingMethod)
	r.Post("/voters/me/change-password", h.ChangePassword)
	r.Get("/voters/me/participation-stats", h.GetParticipationStats)
	r.Delete("/voters/me/photo", h.DeletePhoto)
}

// GetCompleteProfile handles GET /voters/me/complete-profile
func (h *ProfileHandler) GetCompleteProfile(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid atau tidak ditemukan.")
		return
	}

	voterID, ok := ctxkeys.GetVoterID(r.Context())
	if !ok {
		response.Forbidden(w, "FORBIDDEN", "Hanya voter yang dapat mengakses profil.")
		return
	}

	profile, err := h.service.GetCompleteProfile(r.Context(), voterID, userID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, profile)
}

// UpdateProfile handles PUT /voters/me/profile
func (h *ProfileHandler) UpdateProfile(w http.ResponseWriter, r *http.Request) {
	voterID, ok := ctxkeys.GetVoterID(r.Context())
	if !ok {
		response.Forbidden(w, "FORBIDDEN", "Hanya voter yang dapat mengakses profil.")
		return
	}

	var req UpdateProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Format JSON tidak valid.")
		return
	}

	err := h.service.UpdateProfile(r.Context(), voterID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Build updated fields list
	updatedFields := []string{}
	if req.Email != nil {
		updatedFields = append(updatedFields, "email")
	}
	if req.Phone != nil {
		updatedFields = append(updatedFields, "phone")
	}
	if req.PhotoURL != nil {
		updatedFields = append(updatedFields, "photo")
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"success":        true,
		"message":        "Profil berhasil diperbarui",
		"updated_fields": updatedFields,
	})
}

// UpdateVotingMethod handles PUT /voters/me/voting-method
func (h *ProfileHandler) UpdateVotingMethod(w http.ResponseWriter, r *http.Request) {
	voterID, ok := ctxkeys.GetVoterID(r.Context())
	if !ok {
		response.Forbidden(w, "FORBIDDEN", "Hanya voter yang dapat mengakses profil.")
		return
	}

	var req UpdateVotingMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Format JSON tidak valid.")
		return
	}

	// Validation
	if req.ElectionID == 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "election_id wajib diisi")
		return
	}

	if req.PreferredMethod != "ONLINE" && req.PreferredMethod != "TPS" {
		response.BadRequest(w, "VALIDATION_ERROR", "preferred_method harus ONLINE atau TPS")
		return
	}

	err := h.service.UpdateVotingMethod(r.Context(), voterID, req.ElectionID, req.PreferredMethod)
	if err != nil {
		h.handleError(w, err)
		return
	}

	warning := ""
	if req.PreferredMethod == "ONLINE" {
		warning = "Jika sudah check-in TPS, perubahan tidak berlaku untuk election ini"
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"message":    "Metode voting berhasil diubah ke " + req.PreferredMethod,
		"new_method": req.PreferredMethod,
		"warning":    warning,
	})
}

// ChangePassword handles POST /voters/me/change-password
func (h *ProfileHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid atau tidak ditemukan.")
		return
	}

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_JSON", "Format JSON tidak valid.")
		return
	}

	// Validation
	if req.CurrentPassword == "" {
		response.BadRequest(w, "VALIDATION_ERROR", "current_password wajib diisi")
		return
	}

	if req.NewPassword == "" {
		response.BadRequest(w, "VALIDATION_ERROR", "new_password wajib diisi")
		return
	}

	if req.ConfirmPassword == "" {
		response.BadRequest(w, "VALIDATION_ERROR", "confirm_password wajib diisi")
		return
	}

	err := h.service.ChangePassword(r.Context(), userID, &req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Password berhasil diubah",
	})
}

// GetParticipationStats handles GET /voters/me/participation-stats
func (h *ProfileHandler) GetParticipationStats(w http.ResponseWriter, r *http.Request) {
	voterID, ok := ctxkeys.GetVoterID(r.Context())
	if !ok {
		response.Forbidden(w, "FORBIDDEN", "Hanya voter yang dapat mengakses profil.")
		return
	}

	stats, err := h.service.GetParticipationStats(r.Context(), voterID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, stats)
}

// DeletePhoto handles DELETE /voters/me/photo
func (h *ProfileHandler) DeletePhoto(w http.ResponseWriter, r *http.Request) {
	voterID, ok := ctxkeys.GetVoterID(r.Context())
	if !ok {
		response.Forbidden(w, "FORBIDDEN", "Hanya voter yang dapat mengakses profil.")
		return
	}

	err := h.service.DeletePhoto(r.Context(), voterID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Foto profil berhasil dihapus",
	})
}

func (h *ProfileHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrVoterNotFound):
		response.NotFound(w, "VOTER_NOT_FOUND", "Voter tidak ditemukan.")

	case errors.Is(err, ErrInvalidEmail):
		response.BadRequest(w, "INVALID_EMAIL", "Format email tidak valid.")

	case errors.Is(err, ErrInvalidPhone):
		response.BadRequest(w, "INVALID_PHONE", "Format nomor telepon tidak valid. Gunakan format 08xxx atau +62xxx.")

	case errors.Is(err, ErrInvalidVotingMethod):
		response.BadRequest(w, "INVALID_METHOD", "Metode voting tidak valid. Gunakan ONLINE atau TPS.")

	case errors.Is(err, ErrPasswordMismatch):
		response.BadRequest(w, "PASSWORD_MISMATCH", "Konfirmasi password tidak cocok.")

	case errors.Is(err, ErrPasswordTooShort):
		response.BadRequest(w, "PASSWORD_TOO_SHORT", "Password minimal 8 karakter.")

	case errors.Is(err, ErrPasswordSameAsCurrent):
		response.BadRequest(w, "PASSWORD_SAME", "Password baru tidak boleh sama dengan password lama.")

	case errors.Is(err, ErrInvalidCurrentPassword):
		response.Unauthorized(w, "INVALID_PASSWORD", "Password saat ini salah.")

	case errors.Is(err, ErrAlreadyVoted):
		response.BadRequest(w, "ALREADY_VOTED", "Tidak dapat mengubah metode voting karena sudah voting.")

	case errors.Is(err, ErrAlreadyCheckedIn):
		response.BadRequest(w, "ALREADY_CHECKED_IN", "Tidak dapat mengubah ke ONLINE karena sudah check-in di TPS.")

	default:
		slog.Error("profile handler error", "error", err)
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}
