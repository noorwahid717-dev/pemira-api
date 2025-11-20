package election

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"pemira-api/internal/auth"
	"pemira-api/internal/http/response"
)

type ElectionService interface {
	GetCurrentElection(ctx context.Context) (*CurrentElectionDTO, error)
	GetMeStatus(ctx context.Context, authUser auth.AuthUser, electionID int64) (*MeStatusDTO, error)
}

type Handler struct {
	svc ElectionService
}

func NewHandler(svc ElectionService) *Handler {
	return &Handler{svc: svc}
}

func (h *Handler) RegisterRoutes(r chi.Router) {
	r.Get("/elections/current", h.GetCurrent)
	r.Get("/elections/{electionID}/me/status", h.GetMeStatus)
}

func parseInt64Param(r *http.Request, name string) (int64, error) {
	s := chi.URLParam(r, name)
	return strconv.ParseInt(s, 10, 64)
}

// GetCurrent handles GET /elections/current
func (h *Handler) GetCurrent(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	dto, err := h.svc.GetCurrentElection(ctx)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Tidak ada pemilu yang sedang berlangsung.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// GetMeStatus handles GET /elections/{id}/me/status
func (h *Handler) GetMeStatus(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseInt64Param(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	authUser, ok := auth.FromContext(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "Token tidak valid.")
		return
	}

	dto, err := h.svc.GetMeStatus(ctx, authUser, electionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrUnauthorizedRole):
			response.Forbidden(w, "FORBIDDEN", "Hanya mahasiswa yang dapat mengakses status pemilih.")
			return
		case errors.Is(err, ErrVoterMappingMissing):
			response.Forbidden(w, "VOTER_MAPPING_MISSING", "Akun ini belum terhubung dengan data pemilih.")
			return
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
			return
		}
	}

	response.JSON(w, http.StatusOK, dto)
}
