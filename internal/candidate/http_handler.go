package candidate

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
	r.Get("/elections/{electionID}/candidates", h.ListByElection)
	r.Get("/candidates/{id}", h.GetByID)
}

func (h *Handler) ListByElection(w http.ResponseWriter, r *http.Request) {
	electionID, err := strconv.ParseInt(chi.URLParam(r, "electionID"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid election ID", nil)
		return
	}

	candidates, err := h.service.ListByElection(r.Context(), electionID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch candidates")
		return
	}

	response.Success(w, http.StatusOK, candidates)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid candidate ID", nil)
		return
	}

	candidate, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "Candidate not found")
		return
	}

	response.Success(w, http.StatusOK, candidate)
}
