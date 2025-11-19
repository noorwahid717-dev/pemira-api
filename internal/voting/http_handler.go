package voting

import (
	"encoding/json"
	"net/http"

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
	r.Post("/voting/cast", h.CastVote)
	r.Get("/voting/live-count/{electionID}", h.GetLiveCount)
}

func (h *Handler) CastVote(w http.ResponseWriter, r *http.Request) {
	var req CastVoteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(w, "Validation failed", err.Error())
		return
	}

	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	if err := h.service.CastVote(r.Context(), req.ElectionID, userID, req.CandidateID, "ONLINE"); err != nil {
		response.InternalServerError(w, "Failed to cast vote")
		return
	}

	response.Success(w, http.StatusOK, CastVoteResponse{
		Success: true,
	})
}

func (h *Handler) GetLiveCount(w http.ResponseWriter, r *http.Request) {
	response.Success(w, http.StatusOK, map[string]string{"status": "not implemented"})
}
