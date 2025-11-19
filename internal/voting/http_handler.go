package voting

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

type Handler struct {
	service  *Service
	validate *validator.Validate
}

func NewHandler(service *Service) *Handler {
	return &Handler{
		service:  service,
		validate: validator.New(),
	}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public endpoints (with auth middleware applied at router level)
	r.Get("/voting/config", h.GetVotingConfig)
	r.Post("/voting/online/cast", h.CastOnlineVote)
	r.Post("/voting/tps/cast", h.CastTPSVote)
	r.Get("/voting/tps/status", h.GetTPSVotingStatus)
	r.Get("/voting/receipt", h.GetVotingReceipt)
}

// GET /voting/config
func (h *Handler) GetVotingConfig(w http.ResponseWriter, r *http.Request) {
	voterID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	config, err := h.service.GetVotingConfig(r.Context(), voterID)
	if err != nil {
		response.InternalServerError(w, "Failed to get voting config")
		return
	}

	response.Success(w, http.StatusOK, config)
}

// POST /voting/online/cast
func (h *Handler) CastOnlineVote(w http.ResponseWriter, r *http.Request) {
	var req CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "candidate_id tidak boleh kosong", map[string]string{
			"field": "candidate_id",
			"constraint": "required",
		})
		return
	}

	voterID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau sudah expire", nil)
		return
	}

	receipt, err := h.service.CastOnlineVote(r.Context(), voterID, req.CandidateID)
	if err != nil {
		h.handleVotingError(w, err)
		return
	}

	response.Success(w, http.StatusOK, receipt)
}

// POST /voting/tps/cast
func (h *Handler) CastTPSVote(w http.ResponseWriter, r *http.Request) {
	var req CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "candidate_id tidak boleh kosong", nil)
		return
	}

	voterID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau sudah expire", nil)
		return
	}

	receipt, err := h.service.CastTPSVote(r.Context(), voterID, req.CandidateID)
	if err != nil {
		h.handleVotingError(w, err)
		return
	}

	response.Success(w, http.StatusOK, receipt)
}

// GET /voting/tps/status
func (h *Handler) GetTPSVotingStatus(w http.ResponseWriter, r *http.Request) {
	voterID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau sudah expire", nil)
		return
	}

	status, err := h.service.GetTPSVotingStatus(r.Context(), voterID)
	if err != nil {
		response.InternalServerError(w, "Failed to get TPS voting status")
		return
	}

	response.Success(w, http.StatusOK, status)
}

// GET /voting/receipt
func (h *Handler) GetVotingReceipt(w http.ResponseWriter, r *http.Request) {
	voterID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau sudah expire", nil)
		return
	}

	receipt, err := h.service.GetVotingReceipt(r.Context(), voterID)
	if err != nil {
		response.InternalServerError(w, "Failed to get voting receipt")
		return
	}

	response.Success(w, http.StatusOK, receipt)
}

// Helper to handle voting-specific errors
func (h *Handler) handleVotingError(w http.ResponseWriter, err error) {
	switch err.Error() {
	case "ALREADY_VOTED":
		response.Error(w, http.StatusConflict, "ALREADY_VOTED", "Anda sudah menggunakan hak suara untuk pemilu ini.", nil)
	case "NOT_ELIGIBLE":
		response.Error(w, http.StatusBadRequest, "NOT_ELIGIBLE", "Anda tidak eligible untuk voting.", nil)
	case "ELECTION_NOT_OPEN":
		response.Error(w, http.StatusBadRequest, "ELECTION_NOT_OPEN", "Fase voting belum dibuka atau sudah ditutup.", nil)
	case "ELECTION_NOT_FOUND":
		response.Error(w, http.StatusNotFound, "ELECTION_NOT_FOUND", "Tidak ada pemilu aktif saat ini.", nil)
	case "CANDIDATE_NOT_FOUND":
		response.Error(w, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "Kandidat dengan ID tersebut tidak ditemukan.", nil)
	case "CANDIDATE_INACTIVE":
		response.Error(w, http.StatusBadRequest, "CANDIDATE_INACTIVE", "Kandidat tidak aktif.", nil)
	case "METHOD_NOT_ALLOWED":
		response.Error(w, http.StatusBadRequest, "METHOD_NOT_ALLOWED", "Mode voting tidak diizinkan.", nil)
	case "TPS_REQUIRED":
		response.Error(w, http.StatusBadRequest, "TPS_CHECKIN_NOT_FOUND", "Anda belum melakukan check-in di TPS.", nil)
	case "TPS_CHECKIN_NOT_APPROVED":
		response.Error(w, http.StatusBadRequest, "TPS_CHECKIN_NOT_APPROVED", "Check-in TPS Anda belum disetujui panitia.", nil)
	case "TPS_CHECKIN_EXPIRED":
		response.Error(w, http.StatusBadRequest, "TPS_CHECKIN_EXPIRED", "Check-in TPS Anda sudah kadaluarsa.", nil)
	default:
		response.InternalServerError(w, "Failed to cast vote")
	}
}
