package tps

import (
	"encoding/json"
	"net/http"
	"strconv"

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

// NewTPSHandler creates a new TPS handler
func NewTPSHandler(svc *Service) *Handler {
	return &Handler{
		service:  svc,
		validate: validator.New(),
	}
}

// MountPublic registers public TPS routes (for students)
func (h *Handler) MountPublic(r chi.Router) {
	r.Post("/tps/checkin/scan", h.ScanQR)
	r.Get("/tps/checkin/status", h.StudentCheckinStatus)
}

// MountPanel registers TPS panel routes (for TPS operators)
func (h *Handler) MountPanel(r chi.Router) {
	r.Post("/tps/{tpsID}/checkins/{checkinID}/approve", h.ApproveCheckin)
	r.Post("/tps/{tpsID}/checkins/{checkinID}/reject", h.PanelRejectCheckin)
	r.Get("/tps/{tpsID}/checkins", h.PanelListCheckins)
	r.Get("/tps/{tpsID}/summary", h.PanelGetSummary)
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Admin TPS Management
	r.Route("/admin/tps", func(r chi.Router) {
		r.Get("/", h.AdminListTPS)
		r.Post("/", h.AdminCreateTPS)
		r.Get("/{id}", h.AdminGetTPS)
		r.Put("/{id}", h.AdminUpdateTPS)
		r.Put("/{id}/panitia", h.AdminAssignPanitia)
		r.Post("/{id}/qr/regenerate", h.AdminRegenerateQR)
	})

	// Student Check-in
	r.Route("/tps", func(r chi.Router) {
		r.Post("/checkin/scan", h.ScanQR)
		r.Get("/checkin/status", h.StudentCheckinStatus)
	})

	// TPS Panel (Panitia)
	r.Route("/tps/{tps_id}", func(r chi.Router) {
		r.Get("/summary", h.PanelGetSummary)
		r.Get("/checkins", h.PanelListCheckins)
		r.Post("/checkins/{checkin_id}/approve", h.ApproveCheckin)
		r.Post("/checkins/{checkin_id}/reject", h.PanelRejectCheckin)
	})
}

// ===== ADMIN ENDPOINTS =====

func (h *Handler) AdminListTPS(w http.ResponseWriter, r *http.Request) {
	filter := ListFilter{
		Status:     r.URL.Query().Get("status"),
		ElectionID: parseInt64(r.URL.Query().Get("election_id")),
		Page:       parseInt(r.URL.Query().Get("page"), 1),
		Limit:      parseInt(r.URL.Query().Get("limit"), 20),
	}

	result, err := h.service.List(r.Context(), filter)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Failed to fetch TPS list")
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) AdminGetTPS(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	result, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) AdminCreateTPS(w http.ResponseWriter, r *http.Request) {
	var req CreateTPSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", err.Error())
		return
	}

	id, err := h.service.Create(r.Context(), &req)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusCreated, map[string]interface{}{
		"id":     id,
		"code":   req.Code,
		"status": req.Status,
	})
}

func (h *Handler) AdminUpdateTPS(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	var req UpdateTPSRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", err.Error())
		return
	}

	if err := h.service.Update(r.Context(), id, &req); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"id":     id,
		"status": req.Status,
	})
}

func (h *Handler) AdminAssignPanitia(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	var req AssignPanitiaRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", err.Error())
		return
	}

	if err := h.service.AssignPanitia(r.Context(), id, &req); err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, map[string]interface{}{
		"tps_id":        id,
		"total_members": len(req.Members),
	})
}

func (h *Handler) AdminRegenerateQR(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	result, err := h.service.RegenerateQR(r.Context(), id)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

// ===== STUDENT ENDPOINTS =====

func (h *Handler) StudentScanQR(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}

	var req ScanQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", err.Error())
		return
	}

	result, err := h.service.ScanQR(r.Context(), userID, &req)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) StudentCheckinStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}

	electionID := parseInt64(r.URL.Query().Get("election_id"))
	if electionID == 0 {
		electionID = 1 // Default to active election
	}

	result, err := h.service.GetCheckinStatus(r.Context(), userID, electionID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Failed to get check-in status")
		return
	}

	response.Success(w, http.StatusOK, result)
}

// ===== TPS PANEL ENDPOINTS =====

func (h *Handler) PanelGetSummary(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tps_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}
	role, _ := ctxkeys.GetUserRole(r.Context())

	result, err := h.service.GetTPSSummary(r.Context(), tpsID, userID, role)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) PanelListCheckins(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tps_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	status := r.URL.Query().Get("status")
	if status == "" {
		status = CheckinStatusPending
	}

	page := parseInt(r.URL.Query().Get("page"), 1)
	limit := parseInt(r.URL.Query().Get("limit"), 50)

	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}
	role, _ := ctxkeys.GetUserRole(r.Context())

	result, err := h.service.ListCheckinQueue(r.Context(), tpsID, status, page, limit, role, userID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Failed to fetch check-in queue")
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) PanelApproveCheckin(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tps_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	checkinID, err := strconv.ParseInt(chi.URLParam(r, "checkin_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid check-in ID")
		return
	}

	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}
	role, _ := ctxkeys.GetUserRole(r.Context())
	if !hasPanelAccess(role) {
		hasAccess, _ := h.service.repo.IsPanitiaAssigned(r.Context(), tpsID, userID)
		if !hasAccess {
			response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
			return
		}
	}

	result, err := h.service.ApproveCheckin(r.Context(), tpsID, checkinID, userID)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

func (h *Handler) PanelRejectCheckin(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "tps_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid TPS ID")
		return
	}

	checkinID, err := strconv.ParseInt(chi.URLParam(r, "checkin_id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid check-in ID")
		return
	}

	userID, ok := ctxkeys.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized")
		return
	}
	role, _ := ctxkeys.GetUserRole(r.Context())
	if !hasPanelAccess(role) {
		hasAccess, _ := h.service.repo.IsPanitiaAssigned(r.Context(), tpsID, userID)
		if !hasAccess {
			response.Forbidden(w, "TPS_ACCESS_DENIED", "Akses ditolak.")
			return
		}
	}

	var req RejectCheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "VALIDATION_ERROR", "Validation failed", err.Error())
		return
	}

	result, err := h.service.RejectCheckin(r.Context(), tpsID, checkinID, userID, req.Reason)
	if err != nil {
		code, status := GetErrorCode(err)
		response.Error(w, status, code, err.Error(), nil)
		return
	}

	response.Success(w, http.StatusOK, result)
}

// Helper functions
func parseInt64(s string) int64 {
	i, _ := strconv.ParseInt(s, 10, 64)
	return i
}

func parseInt(s string, defaultVal int) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	return i
}
