package settings

import (
	"encoding/json"
	"net/http"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// GET /api/v1/admin/settings
func (h *Handler) GetSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	settings, err := h.svc.GetAll(ctx)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil settings.")
		return
	}
	
	response.JSON(w, http.StatusOK, settings)
}

// GET /api/v1/admin/settings/active-election
func (h *Handler) GetActiveElection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	electionID, err := h.svc.GetActiveElectionID(ctx)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil active election.")
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"active_election_id": electionID,
	})
}

// PUT /api/v1/admin/settings/active-election
func (h *Handler) UpdateActiveElection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	// Get user ID from context
	userID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Unauthorized.")
		return
	}
	
	var req struct {
		ElectionID int `json:"election_id"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}
	
	if req.ElectionID <= 0 {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "election_id harus lebih dari 0.")
		return
	}
	
	err := h.svc.UpdateActiveElectionID(ctx, req.ElectionID, userID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengupdate active election.")
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Active election berhasil diupdate",
		"active_election_id": req.ElectionID,
	})
}

// GET /api/v1/settings/default-election (Public)
func (h *Handler) GetDefaultElection(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	
	electionID, err := h.svc.repo.GetDefaultElectionID(ctx)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil default election.")
		return
	}
	
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"default_election_id": electionID,
	})
}
