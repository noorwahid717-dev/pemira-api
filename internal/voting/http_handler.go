package voting

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	
	"pemira-api/internal/auth"
	"pemira-api/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// NewVotingHandler creates handler with new voting service
func NewVotingHandler(svc *Service) *Handler {
	return &Handler{service: svc}
}

// Request DTOs
type onlineVoteRequest struct {
	ElectionID  int64 `json:"election_id"`
	CandidateID int64 `json:"candidate_id"`
}

type tpsVoteRequest struct {
	ElectionID  int64 `json:"election_id"`
	CandidateID int64 `json:"candidate_id"`
	TPSID       int64 `json:"tps_id"`
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/voting/online/cast", h.CastOnlineVote)
	r.Post("/voting/tps/cast", h.CastTPSVote)
	r.Get("/voting/tps/status", h.GetTPSVotingStatus)
	r.Get("/voting/receipt", h.GetVotingReceipt)
}

// POST /voting/online/cast
func (h *Handler) CastOnlineVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get auth user from context
	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	// Parse request body
	var reqBody onlineVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	// Validate required fields
	if reqBody.ElectionID <= 0 || reqBody.CandidateID <= 0 {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "election_id dan candidate_id wajib diisi.")
		return
	}

	// Build service request
	req := CastOnlineVoteRequest{
		ElectionID:  reqBody.ElectionID,
		CandidateID: reqBody.CandidateID,
	}

	// Call service
	if err := h.service.CastOnlineVote(ctx, authUser, req); err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Suara online berhasil direkam.",
	})
}

// POST /voting/tps/cast
func (h *Handler) CastTPSVote(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get auth user from context
	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	// Parse request body
	var reqBody tpsVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	// Validate required fields
	if reqBody.ElectionID <= 0 || reqBody.CandidateID <= 0 || reqBody.TPSID <= 0 {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "election_id, candidate_id, dan tps_id wajib diisi.")
		return
	}

	// Build service request
	req := CastTPSVoteRequest{
		ElectionID:  reqBody.ElectionID,
		CandidateID: reqBody.CandidateID,
		TPSID:       reqBody.TPSID,
	}

	// Call service
	if err := h.service.CastTPSVote(ctx, authUser, req); err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Suara TPS berhasil direkam.",
	})
}

// GET /voting/tps/status
func (h *Handler) GetTPSVotingStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	if authUser.VoterID == nil {
		response.Forbidden(w, "VOTER_MAPPING_MISSING", "Akun ini belum terhubung dengan data pemilih.")
		return
	}

	status, err := h.service.GetTPSVotingStatus(ctx, *authUser.VoterID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, status)
}

// GET /voting/receipt
func (h *Handler) GetVotingReceipt(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	if authUser.VoterID == nil {
		response.Forbidden(w, "VOTER_MAPPING_MISSING", "Akun ini belum terhubung dengan data pemilih.")
		return
	}

	receipt, err := h.service.GetVotingReceipt(ctx, *authUser.VoterID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, receipt)
}

// handleError maps domain errors to HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrElectionNotFound):
		response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu aktif tidak ditemukan.")

	case errors.Is(err, ErrElectionNotOpen):
		response.BadRequest(w, "ELECTION_NOT_OPEN", "Pemilu belum dibuka atau sudah ditutup untuk voting.")

	case errors.Is(err, ErrMethodNotAllowed):
		response.BadRequest(w, "CHANNEL_NOT_ALLOWED", "Metode voting ini tidak diizinkan untuk pemilu ini.")

	case errors.Is(err, ErrNotEligible):
		response.Forbidden(w, "VOTER_NOT_ELIGIBLE", "Anda tidak terdaftar sebagai pemilih pada pemilu ini.")

	case errors.Is(err, ErrAlreadyVoted):
		response.Conflict(w, "ALREADY_VOTED", "Anda sudah memberikan suara pada pemilu ini.")

	case errors.Is(err, ErrCandidateNotFound):
		response.NotFound(w, "CANDIDATE_NOT_FOUND", "Kandidat tidak ditemukan untuk pemilu ini.")

	case errors.Is(err, ErrCandidateInactive):
		response.BadRequest(w, "CANDIDATE_INACTIVE", "Kandidat tidak aktif.")

	case errors.Is(err, ErrTPSCheckinNotFound):
		response.BadRequest(w, "TPS_CHECKIN_NOT_FOUND", "Anda belum melakukan check-in TPS yang valid.")

	case errors.Is(err, ErrTPSCheckinNotApproved):
		response.BadRequest(w, "TPS_CHECKIN_NOT_APPROVED", "Check-in Anda belum disetujui panitia TPS.")

	case errors.Is(err, ErrCheckinExpired):
		response.BadRequest(w, "CHECKIN_EXPIRED", "Waktu validasi check-in Anda sudah habis, silakan ulangi di TPS.")

	case errors.Is(err, ErrTPSNotFound):
		response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")

	case errors.Is(err, ErrVoterMappingMissing):
		response.Forbidden(w, "VOTER_MAPPING_MISSING", "Akun ini belum terhubung dengan data pemilih.")

	default:
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}


