package auth

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
	r.Post("/auth/login", h.Login)
	r.Get("/auth/me", h.Me)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body", nil)
		return
	}

	if err := h.validate.Struct(req); err != nil {
		response.BadRequest(w, "Validation failed", err.Error())
		return
	}

	token, user, err := h.service.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		response.Unauthorized(w, "Invalid credentials")
		return
	}

	response.Success(w, http.StatusOK, LoginResponse{
		Token: token,
		User:  user,
	})
}

func (h *Handler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int64)
	if !ok {
		response.Unauthorized(w, "Unauthorized")
		return
	}

	user, err := h.service.repo.FindByID(r.Context(), userID)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	response.Success(w, http.StatusOK, MeResponse{User: user})
}
