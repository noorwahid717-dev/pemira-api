package candidate

import (
	"errors"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"pemira-api/internal/http/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// ListPublic menangani GET /elections/{electionID}/candidates
func (h *Handler) ListPublic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "electionID tidak valid.")
		return
	}

	q := r.URL.Query()
	search := q.Get("search")
	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 10)

	items, pag, err := h.svc.ListPublicCandidates(ctx, electionID, search, page, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := struct {
		Items      []CandidateListItemDTO `json:"items"`
		Pagination Pagination             `json:"pagination"`
	}{
		Items:      items,
		Pagination: pag,
	}

	response.Success(w, http.StatusOK, resp)
}

// DetailPublic menangani GET /elections/{electionID}/candidates/{candidateID}
func (h *Handler) DetailPublic(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "electionID tidak valid.")
		return
	}

	candidateID, err := parseIDParam(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	dto, err := h.svc.GetPublicCandidateDetail(ctx, electionID, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.Success(w, http.StatusOK, dto)
}

// GetPublicProfileMedia handles GET /elections/{electionID}/candidates/{candidateID}/media/profile
func (h *Handler) GetPublicProfileMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseIDParam(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	media, err := h.svc.GetProfileMedia(ctx, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Download from Supabase and stream to client
	if media.URL != "" {
		// Fetch blob from Supabase public URL
		resp, err := http.Get(media.URL)
		if err != nil {
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil foto profil.")
			return
		}
		defer resp.Body.Close()

		// Copy headers
		w.Header().Set("Content-Type", media.ContentType)
		if resp.ContentLength > 0 {
			w.Header().Set("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
		}
		
		// Stream blob to client
		w.WriteHeader(http.StatusOK)
		io.Copy(w, resp.Body)
		return
	}

	// Fallback: return 404 if no URL
	response.NotFound(w, "MEDIA_NOT_FOUND", "Foto profil tidak ditemukan.")
}

// parseIDParam parses URL parameter as int64
func parseIDParam(r *http.Request, name string) (int64, error) {
	s := chi.URLParam(r, name)
	return strconv.ParseInt(s, 10, 64)
}

// parseIntQuery parses query parameter as int with default value
func parseIntQuery(r *http.Request, key string, def int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

// handleError maps service errors to HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrCandidateNotFound):
		response.NotFound(w, "NOT_FOUND", "Kandidat tidak ditemukan untuk pemilu ini.")

	case errors.Is(err, ErrCandidateNotPublished):
		// Dari sisi mahasiswa, diperlakukan sama seperti not found
		response.NotFound(w, "NOT_FOUND", "Kandidat tidak ditemukan untuk pemilu ini.")

	default:
		// TODO: log internal error dengan logger
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}
