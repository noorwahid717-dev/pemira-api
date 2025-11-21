package voting

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

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

type setMethodRequest struct {
	ElectionID int64  `json:"election_id"`
	Method     string `json:"method"`           // ONLINE or TPS
	TPSID      *int64 `json:"tps_id,omitempty"` // required if method=TPS
}

type scanCandidateRequest struct {
	BallotQRPayload string `json:"ballot_qr_payload"`
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Post("/voting/online/cast", h.CastOnlineVote)
	r.Post("/voting/tps/cast", h.CastTPSVote)
	r.Get("/voting/tps/status", h.GetTPSVotingStatus)
	r.Get("/voting/receipt", h.GetVotingReceipt)
	r.Post("/voting/method", h.SetVoterMethod)
	r.Post("/tps/{tpsID}/checkins/{checkinID}/scan-candidate", h.ScanTPSCandidate)
	r.Post("/tps/ballots/parse-qr", h.ParseBallotQR)
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

// POST /voting/method
func (h *Handler) SetVoterMethod(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	var reqBody setMethodRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if reqBody.ElectionID <= 0 || strings.TrimSpace(reqBody.Method) == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "election_id dan method wajib diisi.")
		return
	}

	err := h.service.SetVoterMethod(ctx, authUser, SetMethodRequest{
		ElectionID: reqBody.ElectionID,
		Method:     reqBody.Method,
		TPSID:      reqBody.TPSID,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Preferensi metode voting disimpan.",
	})
}

// POST /tps/{tpsID}/checkins/{checkinID}/scan-candidate
func (h *Handler) ScanTPSCandidate(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}
	checkinID, err := strconv.ParseInt(chi.URLParam(r, "checkinID"), 10, 64)
	if err != nil || checkinID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "checkinID tidak valid.")
		return
	}

	var reqBody scanCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}
	if strings.TrimSpace(reqBody.BallotQRPayload) == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "ballot_qr_payload wajib diisi.")
		return
	}

	result, err := h.service.ScanCandidateAtTPS(ctx, authUser, ScanCandidateRequest{
		TPSID:     tpsID,
		CheckinID: checkinID,
		Payload:   reqBody.BallotQRPayload,
	})
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    result,
	})
}

// POST /tps/ballots/parse-qr (helper)
func (h *Handler) ParseBallotQR(w http.ResponseWriter, r *http.Request) {
	var reqBody scanCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}
	qr, err := parseBallotQR(reqBody.BallotQRPayload)
	if err != nil {
		response.BadRequest(w, "INVALID_BALLOT_QR", "Kode QR surat suara tidak dikenali.")
		return
	}
	response.JSON(w, http.StatusOK, qr)
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

	case errors.Is(err, ErrInvalidBallotQR):
		response.BadRequest(w, "INVALID_BALLOT_QR", "Kode QR surat suara tidak dikenali.")

	case errors.Is(err, ErrElectionMismatch):
		response.BadRequest(w, "ELECTION_MISMATCH", "Kode QR tidak sesuai dengan pemilu di TPS ini.")

	default:
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}
