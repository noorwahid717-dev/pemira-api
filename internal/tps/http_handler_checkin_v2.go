package tps

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

// ScanQR handles student scanning QR code at TPS
// POST /tps/checkin/scan
func (h *Handler) ScanQR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get voter ID from context (set by auth middleware)
	voterID, ok := ctxkeys.GetVoterID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau tidak memiliki akses.", nil)
		return
	}

	// 2. Parse and validate request body
	var req ScanQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "Format body tidak valid.", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "qr_payload wajib diisi.", map[string]string{
			"field":      "qr_payload",
			"constraint": "required",
		})
		return
	}

	// 3. Call service
	result, err := h.service.ScanQR(ctx, voterID, &req)
	if err != nil {
		h.handleTPSError(w, err)
		return
	}

	// 4. Map to response DTO
	dto := ScanQRResponse{
		CheckinID: result.CheckinID,
		TPS: TPSInfo{
			ID:   result.TPS.ID,
			Code: result.TPS.Code,
			Name: result.TPS.Name,
		},
		Status:  result.Status,
		ScanAt:  result.ScanAt,
		Message: "Check-in berhasil. Silakan menunggu verifikasi panitia TPS.",
	}

	response.Success(w, http.StatusOK, dto)
}

// ApproveCheckin handles TPS operator approving a check-in
// POST /tps/{tpsID}/checkins/{checkinID}/approve
func (h *Handler) ApproveCheckin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 1. Get operator user ID from context
	operatorUserID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Token tidak valid atau tidak memiliki akses.", nil)
		return
	}

	// 2. Parse path parameters
	tpsIDStr := chi.URLParam(r, "tpsID")
	checkinIDStr := chi.URLParam(r, "checkinID")

	tpsID, err := strconv.ParseInt(tpsIDStr, 10, 64)
	if err != nil || tpsID <= 0 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "tps_id tidak valid.", nil)
		return
	}

	checkinID, err := strconv.ParseInt(checkinIDStr, 10, 64)
	if err != nil || checkinID <= 0 {
		response.Error(w, http.StatusBadRequest, "VALIDATION_ERROR", "checkin_id tidak valid.", nil)
		return
	}

	// 3. Call service
	result, err := h.service.ApproveCheckin(ctx, tpsID, checkinID, operatorUserID)
	if err != nil {
		h.handleTPSError(w, err)
		return
	}

	// 4. Map to response DTO (use existing ApproveCheckinResponse)
	dto := ApproveCheckinResponse{
		CheckinID:  result.CheckinID,
		Status:     result.Status,
		Voter:      result.Voter,
		TPS:        result.TPS,
		ApprovedAt: result.ApprovedAt,
	}

	response.Success(w, http.StatusOK, dto)
}

// handleTPSError maps TPS domain errors to HTTP responses
func (h *Handler) handleTPSError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrQRInvalid):
		response.Error(w, http.StatusBadRequest, "QR_INVALID", "Kode QR tidak valid.", nil)

	case errors.Is(err, ErrQRRevoked):
		response.Error(w, http.StatusBadRequest, "QR_REVOKED", "Kode QR ini sudah tidak berlaku.", nil)

	case errors.Is(err, ErrTPSNotFound):
		response.Error(w, http.StatusNotFound, "TPS_NOT_FOUND", "TPS tidak ditemukan.", nil)

	case errors.Is(err, ErrTPSInactive):
		response.Error(w, http.StatusBadRequest, "TPS_INACTIVE", "TPS belum atau tidak aktif.", nil)

	case errors.Is(err, ErrTPSClosed):
		response.Error(w, http.StatusBadRequest, "TPS_CLOSED", "TPS sudah ditutup.", nil)

	case errors.Is(err, ErrElectionNotOpen):
		response.Error(w, http.StatusBadRequest, "ELECTION_NOT_OPEN", "Fase voting belum dibuka atau sudah ditutup.", nil)

	case errors.Is(err, ErrNotEligible):
		response.Error(w, http.StatusForbidden, "NOT_ELIGIBLE", "Anda tidak berhak memilih untuk pemilu ini.", nil)

	case errors.Is(err, ErrAlreadyVoted):
		response.Error(w, http.StatusConflict, "ALREADY_VOTED", "Anda sudah menggunakan hak suara.", nil)

	case errors.Is(err, ErrCheckinNotFound):
		response.Error(w, http.StatusNotFound, "CHECKIN_NOT_FOUND", "Data check-in tidak ditemukan.", nil)

	case errors.Is(err, ErrCheckinNotPending):
		response.Error(w, http.StatusBadRequest, "CHECKIN_NOT_PENDING", "Check-in bukan dalam status menunggu (PENDING).", nil)

	case errors.Is(err, ErrCheckinExpired):
		response.Error(w, http.StatusBadRequest, "CHECKIN_EXPIRED", "Waktu check-in sudah kadaluarsa, silakan scan ulang di TPS.", nil)

	case errors.Is(err, ErrTPSAccessDenied):
		response.Error(w, http.StatusForbidden, "TPS_ACCESS_DENIED", "Anda tidak memiliki akses ke TPS ini.", nil)

	case errors.Is(err, ErrTPSCodeDuplicate):
		response.Error(w, http.StatusConflict, "TPS_CODE_DUPLICATE", "Kode TPS sudah digunakan.", nil)

	case errors.Is(err, ErrInvalidTimeFormat):
		response.Error(w, http.StatusBadRequest, "INVALID_TIME_FORMAT", "Format waktu tidak valid.", nil)

	default:
		// Log internal error (production: use structured logger)
		// logger.Error("tps handler error", "err", err)
		response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.", nil)
	}
}
