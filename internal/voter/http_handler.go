package voter

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/voters", h.List)
	r.Get("/voters/{id}", h.GetByID)
	r.Get("/voters/nim/{nim}", h.GetByNIM)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	
	params := shared.NewPaginationParams(page, perPage)
	voters, total, err := h.service.List(r.Context(), params)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch voters")
		return
	}

	response.Success(w, http.StatusOK, shared.NewPaginatedResponse(voters, params, total))
}

func (h *Handler) GetByNIM(w http.ResponseWriter, r *http.Request) {
	nim := chi.URLParam(r, "nim")
	
	voter, err := h.service.GetByNIM(r.Context(), nim)
	if err != nil {
		response.NotFound(w, "Voter not found")
		return
	}

	response.Success(w, http.StatusOK, voter)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{"status": "not implemented"})
}
