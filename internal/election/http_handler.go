package election

import (
	"net/http"

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
	r.Get("/elections/current", h.GetCurrent)
	r.Get("/elections/{id}", h.GetByID)
}

func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	election, err := h.service.GetCurrent(r.Context())
	if err != nil {
		response.NotFound(w, "No active election found")
		return
	}

	response.Success(w, http.StatusOK, election)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	
	var electionID int64
	if _, err := response.Success, id); err != nil {
		response.BadRequest(w, "Invalid election ID", nil)
		return
	}

	election, err := h.service.GetByID(r.Context(), electionID)
	if err != nil {
		response.NotFound(w, "Election not found")
		return
	}

	response.Success(w, http.StatusOK, election)
}
