package election

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"pemira-api/internal/http/response"
)

type AdminHandler struct {
	svc *AdminService
}

func NewAdminHandler(svc *AdminService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

func parseIDParam(r *http.Request, name string) (int64, error) {
	s := chi.URLParam(r, name)
	return strconv.ParseInt(s, 10, 64)
}

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func (h *AdminHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	q := r.URL.Query()

	filter := AdminElectionListFilter{
		Search: q.Get("search"),
	}

	if ys := q.Get("year"); ys != "" {
		if y, err := strconv.Atoi(ys); err == nil {
			filter.Year = &y
		}
	}

	if ss := q.Get("status"); ss != "" {
		status := ElectionStatus(ss)
		filter.Status = &status
	}

	page := parseIntDefault(q.Get("page"), 1)
	limit := parseIntDefault(q.Get("limit"), 20)

	items, pag, err := h.svc.List(ctx, filter, page, limit)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil daftar pemilu.")
		return
	}

	resp := struct {
		Items      []AdminElectionDTO `json:"items"`
		Pagination Pagination         `json:"pagination"`
	}{
		Items:      items,
		Pagination: pag,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req AdminElectionCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.Year <= 0 || req.Name == "" || req.Slug == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "year, name, dan slug wajib diisi.")
		return
	}

	dto, err := h.svc.Create(ctx, req)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal membuat pemilu.")
		return
	}

	response.JSON(w, http.StatusCreated, dto)
}

func (h *AdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil detail pemilu.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	var req AdminElectionUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	dto, err := h.svc.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengubah pemilu.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

func (h *AdminHandler) OpenVoting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.OpenVoting(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrElectionAlreadyOpen):
			response.BadRequest(w, "ELECTION_ALREADY_OPEN", "Pemilu sudah dalam status voting terbuka.")
			return
		case errors.Is(err, ErrInvalidStatusChange):
			response.BadRequest(w, "INVALID_STATUS_CHANGE", "Status pemilu tidak dapat dibuka untuk voting.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal membuka voting.")
			return
		}
	}

	response.JSON(w, http.StatusOK, dto)
}

func (h *AdminHandler) CloseVoting(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.CloseVoting(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrElectionNotInOpenState):
			response.BadRequest(w, "ELECTION_NOT_OPEN", "Pemilu tidak dalam status voting terbuka.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menutup voting.")
			return
		}
	}

	response.JSON(w, http.StatusOK, dto)
}
