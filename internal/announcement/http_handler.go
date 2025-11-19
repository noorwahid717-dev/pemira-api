package announcement

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared"
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

func (h *Handler) RegisterRoutes(r chi.Router) {
	// Public routes
	r.Get("/announcements", h.ListPublished)
	r.Get("/announcements/{id}", h.GetByID)
	
	// Admin routes (should be wrapped with auth middleware)
	r.Post("/admin/announcements", h.Create)
	r.Put("/admin/announcements/{id}", h.Update)
	r.Delete("/admin/announcements/{id}", h.Delete)
	r.Post("/admin/announcements/{id}/publish", h.Publish)
}

func (h *Handler) ListPublished(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	perPage, _ := strconv.Atoi(r.URL.Query().Get("per_page"))
	electionID, _ := strconv.ParseInt(r.URL.Query().Get("election_id"), 10, 64)
	
	params := shared.NewPaginationParams(page, perPage)
	announcements, total, err := h.service.ListPublished(r.Context(), electionID, params)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch announcements")
		return
	}

	response.Success(w, http.StatusOK, shared.NewPaginatedResponse(announcements, params, total))
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid announcement ID", nil)
		return
	}

	announcement, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "Announcement not found")
		return
	}

	response.Success(w, http.StatusOK, announcement)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(w, "Validation failed", err.Error())
		return
	}

	userID, _ := r.Context().Value(ctxkeys.UserIDKey).(int64)

	announcement := &Announcement{
		ElectionID:  req.ElectionID,
		Title:       req.Title,
		Content:     req.Content,
		Type:        req.Type,
		Priority:    req.Priority,
		IsPublished: req.IsPublished,
		CreatedBy:   userID,
	}

	if err := h.service.Create(r.Context(), announcement); err != nil {
		response.InternalServerError(w, "Failed to create announcement")
		return
	}

	response.Success(w, http.StatusCreated, announcement)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid announcement ID", nil)
		return
	}

	var req UpdateAnnouncementRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(w, "Validation failed", err.Error())
		return
	}

	announcement, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "Announcement not found")
		return
	}

	announcement.Title = req.Title
	announcement.Content = req.Content
	announcement.Type = req.Type
	announcement.Priority = req.Priority
	announcement.IsPublished = req.IsPublished

	if err := h.service.Update(r.Context(), announcement); err != nil {
		response.InternalServerError(w, "Failed to update announcement")
		return
	}

	response.Success(w, http.StatusOK, announcement)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid announcement ID", nil)
		return
	}

	if err := h.service.repo.Delete(r.Context(), id); err != nil {
		response.InternalServerError(w, "Failed to delete announcement")
		return
	}

	response.Success(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) Publish(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid announcement ID", nil)
		return
	}

	if err := h.service.Publish(r.Context(), id); err != nil {
		response.InternalServerError(w, "Failed to publish announcement")
		return
	}

	response.Success(w, http.StatusOK, map[string]bool{"success": true})
}
