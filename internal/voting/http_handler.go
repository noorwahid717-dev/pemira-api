package voting

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

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

// NewVotingHandler creates handler with new voting service
func NewVotingHandler(svc *Service) *Handler {
	return &Handler{
		service:  svc,
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

// Mount registers routes to chi router (alternative to RegisterRoutes)
func (h *Handler) Mount(r chi.Router) {
	r.Post("/voting/online/cast", h.CastOnlineVote)
	r.Post("/voting/tps/cast", h.CastTPSVote)
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
// Handles online voting with full validation
func (h *Handler) CastOnlineVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 1. Get voter ID from context (set by auth middleware)
	voterID, ok := ctx.Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau tidak memiliki akses.", nil)
		return
	}

	// 2. Parse and validate request body
	var req CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Format body tidak valid.", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "candidate_id wajib diisi.", map[string]string{
			"field":      "candidate_id",
			"constraint": "required",
		})
		return
	}

	// 3. Call service
	receipt, err := h.service.CastOnlineVote(ctx, voterID, req.CandidateID)
	if err != nil {
		h.handleVotingError(w, err)
		return
	}

	// 4. Map to response DTO
	dto := h.mapToVoteResponse(receipt)
	
	response.Success(w, http.StatusOK, dto)
}

// POST /voting/tps/cast
// Handles TPS voting after check-in approval
func (h *Handler) CastTPSVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// 1. Get voter ID from context (set by auth middleware)
	voterID, ok := ctx.Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau tidak memiliki akses.", nil)
		return
	}

	// 2. Parse and validate request body
	var req CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Format body tidak valid.", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "candidate_id wajib diisi.", map[string]string{
			"field":      "candidate_id",
			"constraint": "required",
		})
		return
	}

	// 3. Call service
	receipt, err := h.service.CastTPSVote(ctx, voterID, req.CandidateID)
	if err != nil {
		h.handleVotingError(w, err)
		return
	}

	// 4. Map to response DTO
	dto := h.mapToVoteResponse(receipt)
	
	response.Success(w, http.StatusOK, dto)
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

// handleVotingError maps domain errors to HTTP responses
func (h *Handler) handleVotingError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrElectionNotFound):
		response.Error(w, http.StatusNotFound, "ELECTION_NOT_FOUND", "Pemilu aktif tidak ditemukan.", nil)

	case errors.Is(err, ErrElectionNotOpen):
		response.Error(w, http.StatusBadRequest, "ELECTION_NOT_OPEN", "Fase voting belum dibuka atau sudah ditutup.", nil)

	case errors.Is(err, ErrNotEligible):
		response.Error(w, http.StatusForbidden, "NOT_ELIGIBLE", "Anda tidak termasuk dalam DPT atau tidak berhak memilih.", nil)

	case errors.Is(err, ErrAlreadyVoted):
		response.Error(w, http.StatusConflict, "ALREADY_VOTED", "Anda sudah menggunakan hak suara untuk pemilu ini.", nil)

	case errors.Is(err, ErrCandidateNotFound):
		response.Error(w, http.StatusNotFound, "CANDIDATE_NOT_FOUND", "Kandidat tidak ditemukan untuk pemilu ini.", nil)

	case errors.Is(err, ErrCandidateInactive):
		response.Error(w, http.StatusBadRequest, "CANDIDATE_INACTIVE", "Kandidat tidak aktif.", nil)

	case errors.Is(err, ErrMethodNotAllowed):
		response.Error(w, http.StatusBadRequest, "METHOD_NOT_ALLOWED", "Metode voting ini tidak diizinkan untuk pemilu sekarang.", nil)

	case errors.Is(err, ErrTPSCheckinNotFound):
		response.Error(w, http.StatusBadRequest, "TPS_CHECKIN_NOT_FOUND", "Anda belum melakukan check-in TPS yang valid.", nil)

	case errors.Is(err, ErrTPSCheckinNotApproved):
		response.Error(w, http.StatusBadRequest, "TPS_CHECKIN_NOT_APPROVED", "Check-in Anda belum disetujui panitia TPS.", nil)

	case errors.Is(err, ErrCheckinExpired):
		response.Error(w, http.StatusBadRequest, "CHECKIN_EXPIRED", "Waktu validasi check-in Anda sudah habis, silakan ulangi di TPS.", nil)

	case errors.Is(err, ErrTPSNotFound):
		response.Error(w, http.StatusNotFound, "TPS_NOT_FOUND", "TPS tidak ditemukan.", nil)

	default:
		// Log internal error (production: use structured logger)
		// logger.Error("voting handler error", "err", err)
		response.InternalServerError(w, "Terjadi kesalahan pada sistem.")
	}
}

// mapToVoteResponse converts VoteReceipt to HTTP response DTO
func (h *Handler) mapToVoteResponse(receipt *VoteReceipt) map[string]interface{} {
	dto := map[string]interface{}{
		"election_id": receipt.ElectionID,
		"voter_id":    receipt.VoterID,
		"method":      receipt.Method,
		"voted_at":    receipt.VotedAt.Format(time.RFC3339),
		"receipt": map[string]interface{}{
			"token_hash": receipt.Receipt.TokenHash,
			"note":       receipt.Receipt.Note,
		},
	}
	
	if receipt.TPS != nil {
		dto["tps"] = map[string]interface{}{
			"id":   receipt.TPS.ID,
			"code": receipt.TPS.Code,
			"name": receipt.TPS.Name,
		}
	}
	
	return dto
}
