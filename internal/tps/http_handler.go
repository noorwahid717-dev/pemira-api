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

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/tps", h.List)
	r.Get("/tps/{id}", h.GetByID)
	r.Post("/tps/{id}/checkin", h.CreateCheckin)
	r.Post("/tps/checkin/{checkinID}/approve", h.ApproveCheckin)
	r.Get("/tps/{id}/checkins", h.ListCheckins)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	electionID, _ := strconv.ParseInt(r.URL.Query().Get("election_id"), 10, 64)
	
	tpsList, err := h.service.List(r.Context(), electionID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch TPS list")
		return
	}

	response.Success(w, http.StatusOK, tpsList)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid TPS ID", nil)
		return
	}

	tps, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.NotFound(w, "TPS not found")
		return
	}

	response.Success(w, http.StatusOK, tps)
}

func (h *Handler) CreateCheckin(w http.ResponseWriter, r *http.Request) {
	var req CheckinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(w, "Validation failed", err.Error())
		return
	}

	checkin := &TPSCheckin{
		TPSID:   req.TPSID,
		VoterID: req.VoterID,
		Status:  "PENDING",
	}

	if err := h.service.CreateCheckin(r.Context(), checkin); err != nil {
		response.InternalServerError(w, "Failed to create checkin")
		return
	}

	response.Success(w, http.StatusCreated, checkin)
}

func (h *Handler) ApproveCheckin(w http.ResponseWriter, r *http.Request) {
	checkinID, err := strconv.ParseInt(chi.URLParam(r, "checkinID"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid checkin ID", nil)
		return
	}

	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	if err := h.service.ApproveCheckin(r.Context(), checkinID, userID); err != nil {
		response.InternalServerError(w, "Failed to approve checkin")
		return
	}

	response.Success(w, http.StatusOK, map[string]bool{"success": true})
}

func (h *Handler) ListCheckins(w http.ResponseWriter, r *http.Request) {
	tpsID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		response.BadRequest(w, "Invalid TPS ID", nil)
		return
	}

	checkins, err := h.service.ListCheckins(r.Context(), tpsID)
	if err != nil {
		response.InternalServerError(w, "Failed to fetch checkins")
		return
	}

	response.Success(w, http.StatusOK, checkins)
}
