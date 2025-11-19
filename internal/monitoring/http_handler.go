package monitoring

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	
	"pemira-api/internal/http/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/admin/monitoring/summary", h.GetSummary)
	r.Get("/admin/monitoring/live-count/{electionID}", h.GetLiveCount)
	r.Get("/admin/monitoring/participation/{electionID}", h.GetParticipation)
}

func (h *Handler) GetSummary(w http.ResponseWriter, r *http.Request) {
	electionID, _ := strconv.ParseInt(r.URL.Query().Get("election_id"), 10, 64)
	
	summary, err := h.service.GetDashboardSummary(r.Context(), electionID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch summary")
		return
	}

	response.Success(w, http.StatusOK, summary)
}

func (h *Handler) GetLiveCount(w http.ResponseWriter, r *http.Request) {
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid election ID", nil)
		return
	}

	snapshot, err := h.service.GetLiveCountSnapshot(r.Context(), electionID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch live count")
		return
	}

	response.Success(w, http.StatusOK, snapshot)
}

func (h *Handler) GetParticipation(w http.ResponseWriter, r *http.Request) {
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid election ID", nil)
		return
	}

	participation, err := h.service.repo.GetParticipationStats(r.Context(), electionID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch participation")
		return
	}

	response.Success(w, http.StatusOK, participation)
}
