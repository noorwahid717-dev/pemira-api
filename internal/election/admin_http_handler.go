package election

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"
	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
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

func parseBrandingSlotParam(r *http.Request) (BrandingSlot, error) {
	slotParam := chi.URLParam(r, "slot")
	return ParseBrandingSlot(slotParam)
}

func newBrandingFileID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}

	// UUID v4 layout
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80

	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

type generalInfoResponse struct {
	ID            int64              `json:"id"`
	Year          int                `json:"year"`
	Slug          string             `json:"slug"`
	Name          string             `json:"name"`
	Description   string             `json:"description"`
	AcademicYear  *string            `json:"academic_year,omitempty"`
	Status        ElectionStatus     `json:"status"`
	CurrentPhase  string             `json:"current_phase,omitempty"`
	OnlineEnabled bool               `json:"online_enabled"`
	TPSEnabled    bool               `json:"tps_enabled"`
	VotingWindow  VotingWindow       `json:"voting_window"`
	Phases        []ElectionPhaseDTO `json:"phases,omitempty"`
	CreatedAt     time.Time          `json:"created_at"`
	UpdatedAt     time.Time          `json:"updated_at"`
}

func successPayload(data interface{}) map[string]interface{} {
	return map[string]interface{}{
		"success": true,
		"data":    data,
	}
}

func buildGeneralInfoResponse(dto *AdminElectionDTO) generalInfoResponse {
	currentPhase := ""
	if dto.CurrentPhase != nil {
		currentPhase = *dto.CurrentPhase
	}

	displayStatus := dto.Status
	if currentPhase != "" {
		displayStatus = ElectionStatus(currentPhase)
	}

	return generalInfoResponse{
		ID:            dto.ID,
		Year:          dto.Year,
		Slug:          dto.Slug,
		Name:          dto.Name,
		Description:   dto.Description,
		AcademicYear:  dto.AcademicYear,
		Status:        displayStatus,
		CurrentPhase:  currentPhase,
		OnlineEnabled: dto.OnlineEnabled,
		TPSEnabled:    dto.TPSEnabled,
		VotingWindow: VotingWindow{
			StartAt: dto.VotingStartAt,
			EndAt:   dto.VotingEndAt,
		},
		Phases:    phaseDTOsFromElection(dto),
		CreatedAt: dto.CreatedAt,
		UpdatedAt: dto.UpdatedAt,
	}
}

const maxBrandingLogoSize = int64(2 * 1024 * 1024) // ~2MB

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

	response.JSON(w, http.StatusOK, buildGeneralInfoResponse(dto))
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

func (h *AdminHandler) PatchGeneralInfo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	var req AdminElectionGeneralUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	if req.Name == nil && req.Description == nil && req.AcademicYear == nil {
		response.UnprocessableEntity(w, "INVALID_INPUT", "Minimal satu field diisi.")
		return
	}

	dto, err := h.svc.PatchGeneralInfo(ctx, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrElectionArchived):
			response.BadRequest(w, "ELECTION_ARCHIVED", "Pemilu sudah diarsipkan.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengubah informasi pemilu.")
			return
		}
	}

	resp := map[string]interface{}{
		"id":            dto.ID,
		"name":          dto.Name,
		"description":   dto.Description,
		"academic_year": dto.AcademicYear,
		"status":        dto.Status,
	}

	response.JSON(w, http.StatusOK, successPayload(resp))
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
		case errors.Is(err, ErrElectionAlreadyOpened):
			response.BadRequest(w, "ELECTION_ALREADY_OPENED", "Pemilu sudah dalam status voting terbuka.")
			return
		case errors.Is(err, ErrInvalidStatusChange):
			response.BadRequest(w, "INVALID_STATUS_CHANGE", "Status pemilu tidak dapat dibuka untuk voting.")
			return
		case errors.Is(err, ErrElectionNotInVotingPhase):
			response.BadRequest(w, "ELECTION_NOT_IN_VOTING_PHASE", "Pemilu belum memasuki jadwal voting.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal membuka voting.")
			return
		}
	}

	currentPhase := ""
	if dto.CurrentPhase != nil {
		currentPhase = *dto.CurrentPhase
	}
	resp := map[string]interface{}{
		"id":            dto.ID,
		"status":        dto.Status,
		"current_phase": currentPhase,
		"voting_window": VotingWindow{
			StartAt: dto.VotingStartAt,
			EndAt:   dto.VotingEndAt,
		},
	}

	response.JSON(w, http.StatusOK, successPayload(resp))
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
		case errors.Is(err, ErrElectionAlreadyClosed):
			response.BadRequest(w, "ELECTION_ALREADY_CLOSED", "Pemilu sudah ditutup.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menutup voting.")
			return
		}
	}

	currentPhase := ""
	if dto.CurrentPhase != nil {
		currentPhase = *dto.CurrentPhase
	}
	resp := map[string]interface{}{
		"id":            dto.ID,
		"status":        dto.Status,
		"current_phase": currentPhase,
	}

	response.JSON(w, http.StatusOK, successPayload(resp))
}

func (h *AdminHandler) Archive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.Archive(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrElectionArchived):
			response.BadRequest(w, "ELECTION_ALREADY_ARCHIVED", "Pemilu sudah diarsipkan.")
			return
		case errors.Is(err, ErrElectionNotClosable):
			response.BadRequest(w, "ELECTION_NOT_CLOSABLE", "Pemilu belum bisa diarsipkan.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengarsipkan pemilu.")
			return
		}
	}

	resp := map[string]interface{}{
		"id":     dto.ID,
		"status": dto.Status,
	}

	response.JSON(w, http.StatusOK, successPayload(resp))
}

func (h *AdminHandler) GetPhases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	resp, err := h.svc.GetPhases(ctx, id)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil jadwal tahapan.")
		return
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) UpdatePhases(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	var req UpdateElectionPhasesRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	respData, err := h.svc.UpdatePhases(ctx, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrInvalidPhaseKey):
			response.UnprocessableEntity(w, "INVALID_PHASE_KEY", "Key tahapan tidak valid atau tidak lengkap.")
			return
		case errors.Is(err, ErrPhaseTimeConflict):
			response.BadRequest(w, "PHASE_TIME_CONFLICT", "Rentang waktu tahapan bertabrakan.")
			return
		case errors.Is(err, ErrVotingPhaseLocked):
			response.BadRequest(w, "VOTING_PHASE_LOCKED", "Jadwal voting tidak dapat diubah saat voting sudah dibuka.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal memperbarui jadwal tahapan.")
			return
		}
	}

	response.JSON(w, http.StatusOK, successPayload(respData))
}

func (h *AdminHandler) GetModeSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.GetModeSettings(ctx, id)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil pengaturan mode.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

func (h *AdminHandler) UpdateModeSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	var req ModeSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Body tidak valid.")
		return
	}

	dto, err := h.svc.UpdateModeSettings(ctx, id, req)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrInvalidModeCombination):
			response.BadRequest(w, "INVALID_MODE_COMBINATION", "online_enabled dan tps_enabled tidak boleh keduanya false.")
			return
		case errors.Is(err, ErrElectionAlreadyStarted):
			response.BadRequest(w, "ELECTION_ALREADY_STARTED", "Mode tidak bisa diubah karena pemilu sudah berjalan.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal memperbarui pengaturan mode.")
			return
		}
	}

	response.JSON(w, http.StatusOK, successPayload(dto))
}

func (h *AdminHandler) GetSummary(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	id, err := parseIDParam(r, "electionID")
	if err != nil || id <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	dto, err := h.svc.GetSummary(ctx, id)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil ringkasan pemilu.")
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

func (h *AdminHandler) GetBranding(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	branding, err := h.svc.GetBranding(ctx, electionID)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil branding.")
		return
	}

	response.JSON(w, http.StatusOK, branding)
}

func (h *AdminHandler) GetBrandingLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	slot, err := parseBrandingSlotParam(r)
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "slot harus primary atau secondary.")
		return
	}

	file, err := h.svc.GetBrandingLogo(ctx, electionID, slot)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrBrandingFileNotFound):
			response.NotFound(w, "BRANDING_LOGO_NOT_FOUND", "Logo belum diunggah.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil logo.")
			return
		}
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(file.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(file.Data)
}

func (h *AdminHandler) UploadBrandingLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	slot, err := parseBrandingSlotParam(r)
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "slot harus primary atau secondary.")
		return
	}

	adminID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}

	if err := r.ParseMultipartForm(maxBrandingLogoSize + (512 << 10)); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Gagal membaca form upload.")
		return
	}

	filePart, _, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Field file wajib diisi.")
		return
	}
	defer filePart.Close()

	limitedReader := io.LimitReader(filePart, maxBrandingLogoSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Gagal membaca file upload.")
		return
	}

	if int64(len(data)) > maxBrandingLogoSize {
		response.UnprocessableEntity(w, "FILE_TOO_LARGE", "Ukuran logo maksimal 2MB.")
		return
	}

	mime := mimetype.Detect(data)
	if mime == nil || !(mime.Is("image/png") || mime.Is("image/jpeg")) {
		response.UnprocessableEntity(w, "INVALID_FILE_TYPE", "Logo harus berupa PNG atau JPEG.")
		return
	}

	fileID, err := newBrandingFileID()
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menyiapkan penyimpanan logo.")
		return
	}

	file := BrandingFileCreate{
		ID:          fileID,
		ContentType: mime.String(),
		SizeBytes:   int64(len(data)),
		Data:        data,
		CreatedByID: adminID,
	}

	saved, err := h.svc.UploadBrandingLogo(ctx, electionID, slot, file)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrInvalidBrandingSlot):
			response.BadRequest(w, "VALIDATION_ERROR", "slot tidak valid.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menyimpan logo.")
			return
		}
	}

	resp := map[string]interface{}{
		"id":           saved.ID,
		"content_type": saved.ContentType,
		"size":         saved.SizeBytes,
	}

	response.JSON(w, http.StatusOK, resp)
}

func (h *AdminHandler) DeleteBrandingLogo(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	slot, err := parseBrandingSlotParam(r)
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "slot harus primary atau secondary.")
		return
	}

	adminID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}

	branding, err := h.svc.DeleteBrandingLogo(ctx, electionID, slot, adminID)
	if err != nil {
		switch {
		case errors.Is(err, ErrElectionNotFound):
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		case errors.Is(err, ErrInvalidBrandingSlot):
			response.BadRequest(w, "VALIDATION_ERROR", "slot tidak valid.")
			return
		default:
			response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menghapus logo.")
			return
		}
	}

	response.JSON(w, http.StatusOK, branding)
}

// GET /admin/elections/{electionID}/settings
func (h *AdminHandler) GetAllSettings(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseIDParam(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	// Get general info
	election, err := h.svc.Get(ctx, electionID)
	if err != nil {
		if errors.Is(err, ErrElectionNotFound) {
			response.NotFound(w, "ELECTION_NOT_FOUND", "Pemilu tidak ditemukan.")
			return
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil detail pemilu.")
		return
	}

	// Get phases
	phases, err := h.svc.GetPhases(ctx, electionID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil jadwal tahapan.")
		return
	}

	// Get mode settings
	modeSettings, err := h.svc.GetModeSettings(ctx, electionID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil pengaturan mode.")
		return
	}

	// Get branding
	branding, err := h.svc.GetBranding(ctx, electionID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil branding.")
		return
	}

	// Combine all settings
	settings := map[string]interface{}{
		"election":      buildGeneralInfoResponse(election),
		"phases":        phases,
		"mode_settings": modeSettings,
		"branding":      branding,
	}

	response.JSON(w, http.StatusOK, settings)
}
