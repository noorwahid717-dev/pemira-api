package tps

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

type PanelHandler struct {
	svc *PanelService
}

func NewPanelHandler(svc *PanelService) *PanelHandler {
	return &PanelHandler{svc: svc}
}

// GET /tps-panel/me
func (h *PanelHandler) Me(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}
	tpsID, ok := ctxkeys.GetTPSID(ctx)
	if !ok {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Token tidak memiliki TPS.")
		return
	}

	operator, err := h.svc.repo.GetOperatorInfo(ctx, userID)
	if err != nil {
		response.Error(w, http.StatusForbidden, "TPS_ACCESS_DENIED", "Akses ditolak.", nil)
		return
	}
	if operator.TPSID == nil || *operator.TPSID != tpsID {
		response.Error(w, http.StatusForbidden, "TPS_ACCESS_DENIED", "Akses ditolak.", nil)
		return
	}

	tpsRow, err := h.svc.repo.GetByID(ctx, tpsID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, "TPS tidak ditemukan.", nil)
		return
	}

	payload := map[string]interface{}{
		"user_id": operator.ID,
		"name":    operator.Name,
		"role":    "TPS_OPERATOR",
		"tps": map[string]interface{}{
			"id":         tpsRow.ID,
			"code":       tpsRow.Code,
			"name":       tpsRow.Name,
			"location":   tpsRow.Location,
			"open_time":  tpsRow.OpenTime,
			"close_time": tpsRow.CloseTime,
			"status":     h.svc.derivePanelStatus(tpsRow),
		},
	}

	response.JSON(w, http.StatusOK, payload)
}

// GET /tps-panel/dashboard
func (h *PanelHandler) Dashboard(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}

	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	if _, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	data, err := h.svc.Dashboard(ctx, tpsID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, data)
}

// GET /tps-panel/checkins
func (h *PanelHandler) ListCheckins(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	if _, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	q := r.URL.Query()
	status := q.Get("status")
	search := q.Get("search")
	limit := parseIntDefault(q.Get("limit"), 50)
	offset := parseIntDefault(q.Get("offset"), 0)

	switch strings.ToUpper(status) {
	case "CHECKED_IN":
		status = CheckinStatusApproved
	case "VOTED":
		status = CheckinStatusVoted
	case "ALL", "":
		status = ""
	default:
		status = CheckinStatusApproved
	}

	items, total, err := h.svc.ListCheckins(ctx, tpsID, status, search, limit, offset)
	if err != nil {
		code, statusCode := GetErrorCode(err)
		response.Error(w, statusCode, code, err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"items": items,
		"total": total,
	})
}

// GET /tps-panel/checkins/{checkinId}
func (h *PanelHandler) GetCheckin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	if _, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	checkinIDStr := chi.URLParam(r, "checkinId")
	checkinID, err := strconv.ParseInt(checkinIDStr, 10, 64)
	if err != nil || checkinID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "checkinId tidak valid.")
		return
	}

	row, err := h.svc.GetCheckin(ctx, checkinID)
	if err != nil {
		if err == ErrCheckinNotFound {
			response.NotFound(w, "CHECKIN_NOT_FOUND", "Check-in tidak ditemukan.")
			return
		}
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}
	if row.TPSID != tpsID {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Tidak bisa mengakses check-in TPS lain.")
		return
	}

	response.JSON(w, http.StatusOK, row)
}

// POST /tps-panel/checkin/scan
func (h *PanelHandler) ScanCheckin(w http.ResponseWriter, r *http.Request) {
	h.handleCheckin(w, r, true)
}

// POST /tps-panel/checkin/manual
func (h *PanelHandler) ManualCheckin(w http.ResponseWriter, r *http.Request) {
	h.handleCheckin(w, r, false)
}

func (h *PanelHandler) handleCheckin(w http.ResponseWriter, r *http.Request, useQR bool) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	if _, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	var payload struct {
		RegistrationQRPayload string `json:"registration_qr_payload"`
		RegistrationCode      string `json:"registration_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	raw := payload.RegistrationQRPayload
	if !useQR {
		raw = payload.RegistrationCode
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "Kode registrasi wajib diisi.")
		return
	}

	result, err := h.svc.repo.ParseRegistrationCode(ctx, raw)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_REGISTRATION_QR", "Kode QR pendaftaran tidak dikenali.", nil)
		return
	}

	if result.TPSID != nil && *result.TPSID != tpsID {
		response.Error(w, http.StatusBadRequest, "TPS_MISMATCH", "Kode QR tidak sesuai dengan TPS ini.", nil)
		return
	}

	if result.ElectionID != electionID {
		response.Error(w, http.StatusBadRequest, "INVALID_REGISTRATION_QR", "Kode QR tidak sesuai dengan pemilu ini.", nil)
		return
	}

	// Ensure TPS matches token
	result.TPSID = &tpsID

	checkin, err := h.svc.repo.CreatePanelCheckin(ctx, *result)
	if err != nil {
		switch err {
		case ErrNotEligible:
			response.Error(w, http.StatusBadRequest, "NOT_TPS_VOTER", "Pemilih ini terdaftar sebagai pemilih online, bukan TPS.", nil)
			return
		case ErrNotTPSVoter:
			response.Error(w, http.StatusBadRequest, "NOT_TPS_VOTER", "Pemilih ini terdaftar sebagai pemilih online, bukan TPS.", nil)
			return
		case ErrAlreadyVoted:
			response.Error(w, http.StatusConflict, "ALREADY_VOTED", "Pemilih ini sudah memberikan suara pada pemilu ini.", nil)
			return
		case ErrCheckinAlreadyExists:
			response.Error(w, http.StatusBadRequest, "CHECKIN_EXISTS", "Pemilih sudah memiliki check-in aktif.", nil)
			return
		case ErrTPSMismatch:
			response.Error(w, http.StatusBadRequest, "TPS_MISMATCH", "Kode QR tidak sesuai dengan TPS ini.", nil)
			return
		default:
			code, status := GetErrorCode(err)
			response.Error(w, status, code, err.Error(), nil)
			return
		}
	}

	resp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"checkin_id":  checkin.ID,
			"election_id": checkin.ElectionID,
			"tps_id":      checkin.TPSID,
			"voter": map[string]interface{}{
				"id":      checkin.VoterID,
				"name":    checkin.VoterName,
				"nim":     checkin.VoterNIM,
				"faculty": checkin.Faculty,
				"program": checkin.Program,
			},
			"status":       mapCheckinStatus(checkin.Status),
			"checkin_time": checkin.ScanAt,
		},
	}

	response.JSON(w, http.StatusOK, resp)
}

// GET /tps-panel/status
func (h *PanelHandler) Status(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	if _, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	resp, err := h.svc.Status(ctx, tpsID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, resp)
}

// GET /tps-panel/stats/timeline
func (h *PanelHandler) Timeline(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role == "" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}

	tpsRow, err := h.svc.EnsureAccess(ctx, electionID, tpsID, role, &tokenTPS)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	points, err := h.svc.Timeline(ctx, tpsID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"election_id": tpsRow.ElectionID,
		"tps_id":      tpsID,
		"points":      points,
	})
}

// POST /tps/{tpsID}/checkins  (for operator without election path)
func (h *PanelHandler) CreateCheckinSimple(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tpsID"), 10, 64)
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsId tidak valid.")
		return
	}
	role, _ := ctxkeys.GetUserRole(ctx)
	tokenTPS, _ := ctxkeys.GetTPSID(ctx)
	if role != "TPS_OPERATOR" {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
		return
	}
	if tokenTPS != tpsID {
		response.Forbidden(w, "TPS_ACCESS_DENIED", "Token tidak sesuai TPS.")
		return
	}

	var payload struct {
		QRPayload        string `json:"qr_payload"`
		RegistrationCode string `json:"registration_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}
	raw := strings.TrimSpace(payload.QRPayload)
	if raw == "" {
		raw = strings.TrimSpace(payload.RegistrationCode)
	}
	if raw == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "qr_payload atau registration_code wajib diisi.")
		return
	}

	checkin, err := h.svc.CreateCheckinViaQR(ctx, tpsID, raw)
	if err != nil {
		switch err {
		case ErrQRInvalid:
			response.Error(w, http.StatusBadRequest, "INVALID_REGISTRATION_QR", "Kode QR pendaftaran tidak dikenali.", nil)
			return
		case ErrNotTPSVoter:
			response.Error(w, http.StatusBadRequest, "NOT_TPS_VOTER", "Pemilih ini terdaftar sebagai pemilih online, bukan TPS.", nil)
			return
		case ErrAlreadyVoted:
			response.Error(w, http.StatusConflict, "ALREADY_VOTED", "Pemilih ini sudah memberikan suara pada pemilu ini.", nil)
			return
		case ErrCheckinAlreadyExists:
			response.Error(w, http.StatusBadRequest, "CHECKIN_EXISTS", "Pemilih sudah memiliki check-in aktif.", nil)
			return
		case ErrTPSMismatch:
			response.Error(w, http.StatusBadRequest, "TPS_MISMATCH", "Kode QR tidak sesuai dengan TPS ini.", nil)
			return
		default:
			code, status := GetErrorCode(err)
			response.Error(w, status, code, err.Error(), nil)
			return
		}
	}

	resp := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"checkin_id":  checkin.ID,
			"election_id": checkin.ElectionID,
			"tps_id":      checkin.TPSID,
			"voter": map[string]interface{}{
				"id":      checkin.VoterID,
				"name":    checkin.VoterName,
				"nim":     checkin.VoterNIM,
				"faculty": checkin.Faculty,
				"program": checkin.Program,
			},
			"status":       mapCheckinStatus(checkin.Status),
			"checkin_time": checkin.ScanAt,
		},
	}
	response.JSON(w, http.StatusOK, resp)
}

// GET /admin/elections/{electionID}/tps
func (h *PanelHandler) ListTPSByElection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionId tidak valid.")
		return
	}

	role, _ := ctxkeys.GetUserRole(ctx)
	if role != "ADMIN" && role != "SUPER_ADMIN" {
		response.Forbidden(w, "FORBIDDEN", "Hanya admin yang dapat mengakses daftar TPS.")
		return
	}

	items, err := h.svc.ListTPSByElection(ctx, electionID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"election_id": electionID,
		"items":       items,
	})
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	val, err := strconv.Atoi(s)
	if err != nil || val < 0 {
		return def
	}
	return val
}
