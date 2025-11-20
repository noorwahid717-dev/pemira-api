package tps

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

// List handles GET /admin/tps
func (h *AdminHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	items, err := h.svc.List(ctx)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil daftar TPS.")
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// Create handles POST /admin/tps
func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req TPSCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.Code == "" || req.Name == "" || req.Location == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "code, name, dan location wajib diisi.")
		return
	}

	dto, err := h.svc.Create(ctx, req)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal membuat TPS.")
		return
	}

	response.JSON(w, http.StatusCreated, dto)
}

// Get handles GET /admin/tps/{tpsID}
func (h *AdminHandler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "tpsID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	dto, err := h.svc.Get(ctx, id)
	if err != nil {
		if errors.Is(err, ErrTPSNotFound) {
			response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil detail TPS.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// Update handles PUT /admin/tps/{tpsID}
func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "tpsID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	var req TPSUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	dto, err := h.svc.Update(ctx, id, req)
	if err != nil {
		if errors.Is(err, ErrTPSNotFound) {
			response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengubah TPS.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// Delete handles DELETE /admin/tps/{tpsID}
func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "tpsID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	if err := h.svc.Delete(ctx, id); err != nil {
		if errors.Is(err, ErrTPSNotFound) {
			response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menghapus TPS.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListOperators handles GET /admin/tps/{tpsID}/operators
func (h *AdminHandler) ListOperators(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tpsID, err := parseIDParam(r, "tpsID")
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	items, err := h.svc.ListOperators(ctx, tpsID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil operator TPS.")
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// CreateOperator handles POST /admin/tps/{tpsID}/operators
func (h *AdminHandler) CreateOperator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tpsID, err := parseIDParam(r, "tpsID")
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	var req CreateOperatorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.Username == "" || req.Password == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "username dan password wajib diisi.")
		return
	}

	op, err := h.svc.CreateOperator(ctx, tpsID, req.Username, req.Password, req.Name, req.Email)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal membuat operator TPS.")
		return
	}

	response.JSON(w, http.StatusCreated, op)
}

// RemoveOperator handles DELETE /admin/tps/{tpsID}/operators/{userID}
func (h *AdminHandler) RemoveOperator(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	tpsID, err := parseIDParam(r, "tpsID")
	if err != nil || tpsID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "tpsID tidak valid.")
		return
	}

	userID, err := parseIDParam(r, "userID")
	if err != nil || userID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "userID tidak valid.")
		return
	}

	if err := h.svc.RemoveOperator(ctx, tpsID, userID); err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menghapus operator TPS.")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Monitor handles GET /admin/elections/{electionID}/tps/monitor
func (h *AdminHandler) Monitor(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	items, err := h.svc.Monitor(ctx, electionID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil monitoring TPS.")
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GetQRMetadata handles GET /admin/tps/{tpsID}/qr
func (h *AdminHandler) GetQRMetadata(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

tpsID, err := parseIDParam(r, "tpsID")
if err != nil {
response.BadRequest(w, "VALIDATION_ERROR", "ID TPS tidak valid.")
return
}

metadata, err := h.svc.GetQRMetadata(ctx, tpsID)
if err != nil {
if errors.Is(err, ErrTPSNotFound) {
response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
return
}
response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil metadata QR.")
return
}

response.JSON(w, http.StatusOK, metadata)
}

// RotateQR handles POST /admin/tps/{tpsID}/qr/rotate
func (h *AdminHandler) RotateQR(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

tpsID, err := parseIDParam(r, "tpsID")
if err != nil {
response.BadRequest(w, "VALIDATION_ERROR", "ID TPS tidak valid.")
return
}

result, err := h.svc.RotateQR(ctx, tpsID)
if err != nil {
if errors.Is(err, ErrTPSNotFound) {
response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
return
}
response.InternalServerError(w, "INTERNAL_ERROR", "Gagal rotate QR.")
return
}

response.JSON(w, http.StatusOK, result)
}

// GetQRForPrint handles GET /admin/tps/{tpsID}/qr/print
func (h *AdminHandler) GetQRForPrint(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

tpsID, err := parseIDParam(r, "tpsID")
if err != nil {
response.BadRequest(w, "VALIDATION_ERROR", "ID TPS tidak valid.")
return
}

printData, err := h.svc.GetQRForPrint(ctx, tpsID)
if err != nil {
if errors.Is(err, ErrTPSNotFound) {
response.NotFound(w, "TPS_NOT_FOUND", "TPS tidak ditemukan.")
return
}
response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil data cetak QR.")
return
}

response.JSON(w, http.StatusOK, printData)
}
