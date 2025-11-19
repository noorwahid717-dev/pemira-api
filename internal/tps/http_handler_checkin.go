package tps

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CheckinHandler struct {
	service *CheckinService
}

func NewCheckinHandler(service *CheckinService) *CheckinHandler {
	return &CheckinHandler{
		service: service,
	}
}

// RegisterRoutes registers all check-in related routes
func (h *CheckinHandler) RegisterRoutes(r *mux.Router) {
	// Student side
	student := r.PathPrefix("/api/v1/voter/tps").Subrouter()
	student.HandleFunc("/scan", h.ScanQR).Methods("POST")
	student.HandleFunc("/status", h.GetCheckinStatus).Methods("GET")

	// TPS Panel side
	panel := r.PathPrefix("/api/v1/tps/{tps_id}").Subrouter()
	panel.HandleFunc("/checkins", h.ListCheckinQueue).Methods("GET")
	panel.HandleFunc("/checkins/{checkin_id}/approve", h.ApproveCheckin).Methods("POST")
	panel.HandleFunc("/checkins/{checkin_id}/reject", h.RejectCheckin).Methods("POST")
}

// ScanQR handles QR code scanning by student
// POST /api/v1/voter/tps/scan
func (h *CheckinHandler) ScanQR(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get voter ID from context (set by auth middleware)
	voterID, ok := ctx.Value("voter_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Voter tidak terautentikasi")
		return
	}

	// Parse request
	var req ScanQRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request tidak valid")
		return
	}

	// Call service
	result, err := h.service.CheckinScan(ctx, voterID, req.QRPayload)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// GetCheckinStatus returns current check-in status
// GET /api/v1/voter/tps/status?election_id=1
func (h *CheckinHandler) GetCheckinStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	_, ok := ctx.Value("voter_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Voter tidak terautentikasi")
		return
	}

	electionIDStr := r.URL.Query().Get("election_id")
	_, err := strconv.ParseInt(electionIDStr, 10, 64)
	if err != nil || electionIDStr == "" {
		respondError(w, http.StatusBadRequest, "INVALID_ELECTION_ID", "Election ID tidak valid")
		return
	}

	// TODO: Implement GetCheckinStatus in service
	// status, err := h.service.GetCheckinStatus(ctx, voterID, electionID)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"has_active_checkin": false,
	})
}

// ListCheckinQueue lists pending/approved check-ins for TPS panel
// GET /api/v1/tps/{tps_id}/checkins?status=PENDING
func (h *CheckinHandler) ListCheckinQueue(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get operator ID from context (set by auth middleware)
	_, ok := ctx.Value("user_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Operator tidak terautentikasi")
		return
	}

	// Parse TPS ID
	vars := mux.Vars(r)
	_, err := strconv.ParseInt(vars["tps_id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_TPS_ID", "TPS ID tidak valid")
		return
	}

	// Parse query params
	status := r.URL.Query().Get("status")
	if status == "" {
		status = CheckinStatusPending
	}

	// TODO: Implement ListCheckinQueue in service
	// queue, err := h.service.ListCheckinQueue(ctx, operatorID, tpsID, status)

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"items": []interface{}{},
	})
}

// ApproveCheckin approves a pending check-in
// POST /api/v1/tps/{tps_id}/checkins/{checkin_id}/approve
func (h *CheckinHandler) ApproveCheckin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get operator ID from context
	operatorID, ok := ctx.Value("user_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Operator tidak terautentikasi")
		return
	}

	// Parse params
	vars := mux.Vars(r)
	tpsID, err := strconv.ParseInt(vars["tps_id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_TPS_ID", "TPS ID tidak valid")
		return
	}

	checkinID, err := strconv.ParseInt(vars["checkin_id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_CHECKIN_ID", "Check-in ID tidak valid")
		return
	}

	// Call service
	result, err := h.service.ApproveCheckin(ctx, operatorID, tpsID, checkinID)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// RejectCheckin rejects a pending check-in
// POST /api/v1/tps/{tps_id}/checkins/{checkin_id}/reject
func (h *CheckinHandler) RejectCheckin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get operator ID from context
	operatorID, ok := ctx.Value("user_id").(int64)
	if !ok {
		respondError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Operator tidak terautentikasi")
		return
	}

	// Parse params
	vars := mux.Vars(r)
	tpsID, err := strconv.ParseInt(vars["tps_id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_TPS_ID", "TPS ID tidak valid")
		return
	}

	checkinID, err := strconv.ParseInt(vars["checkin_id"], 10, 64)
	if err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_CHECKIN_ID", "Check-in ID tidak valid")
		return
	}

	// Parse request body
	var req RejectCheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "INVALID_REQUEST", "Request tidak valid")
		return
	}

	// Validate reason
	if req.Reason == "" {
		respondError(w, http.StatusBadRequest, "REASON_REQUIRED", "Alasan penolakan harus diisi")
		return
	}

	// Call service
	result, err := h.service.RejectCheckin(ctx, operatorID, tpsID, checkinID, req.Reason)
	if err != nil {
		handleServiceError(w, err)
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// Helper functions

func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"data":    data,
	})
}

func respondError(w http.ResponseWriter, status int, code string, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
}

func handleServiceError(w http.ResponseWriter, err error) {
	code, status := GetErrorCode(err)
	respondError(w, status, code, err.Error())
}
