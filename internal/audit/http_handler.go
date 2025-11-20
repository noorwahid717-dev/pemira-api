package audit

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
	// Super admin only
	r.Get("/admin/audit-logs", h.List)
	r.Get("/admin/audit-logs/{id}", h.GetByID)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	
	params := shared.NewPaginationParams(page, perPage)
	
	filters := make(map[string]interface{})
	if entityType := r.URL.Query().Get("entity_type"); entityType != "" {
		filters["entity_type"] = entityType
	}
	if action := r.URL.Query().Get("action"); action != "" {
		filters["action"] = action
	}
	
	logs, total, err := h.service.repo.List(r.Context(), params, filters)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Failed to fetch audit logs")
		return
	}

	response.Success(w, http.StatusOK, shared.NewPaginatedResponse(logs, params, total))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Invalid audit log ID")
		return
	}

	log, err := h.service.repo.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "NOT_FOUND", "Audit log not found")
		return
	}

	response.Success(w, http.StatusOK, log)
}
