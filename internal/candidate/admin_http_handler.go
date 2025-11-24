package candidate

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gabriel-vasile/mimetype"
	"github.com/go-chi/chi/v5"

	"pemira-api/internal/http/response"
	"pemira-api/internal/shared/ctxkeys"
)

// AdminCreateCandidateRequest represents request to create a new candidate
type AdminCreateCandidateRequest struct {
	Number           int             `json:"number"`
	Name             string          `json:"name"`
	PhotoURL         string          `json:"photo_url"`
	ShortBio         string          `json:"short_bio"`
	LongBio          string          `json:"long_bio"`
	Tagline          string          `json:"tagline"`
	FacultyName      string          `json:"faculty_name"`
	StudyProgramName string          `json:"study_program_name"`
	CohortYear       *int            `json:"cohort_year"`
	Vision           string          `json:"vision"`
	Missions         []string        `json:"missions"`
	MainPrograms     []MainProgram   `json:"main_programs"`
	Media            Media           `json:"media"`
	SocialLinks      []SocialLink    `json:"social_links"`
	Status           CandidateStatus `json:"status"`
}

// AdminUpdateCandidateRequest represents request to update a candidate
type AdminUpdateCandidateRequest struct {
	Number           *int             `json:"number,omitempty"`
	Name             *string          `json:"name,omitempty"`
	PhotoURL         *string          `json:"photo_url,omitempty"`
	ShortBio         *string          `json:"short_bio,omitempty"`
	LongBio          *string          `json:"long_bio,omitempty"`
	Tagline          *string          `json:"tagline,omitempty"`
	FacultyName      *string          `json:"faculty_name,omitempty"`
	StudyProgramName *string          `json:"study_program_name,omitempty"`
	CohortYear       *int             `json:"cohort_year,omitempty"`
	Vision           *string          `json:"vision,omitempty"`
	Missions         *[]string        `json:"missions,omitempty"`
	MainPrograms     *[]MainProgram   `json:"main_programs,omitempty"`
	Media            *Media           `json:"media,omitempty"`
	SocialLinks      *[]SocialLink    `json:"social_links,omitempty"`
	Status           *CandidateStatus `json:"status,omitempty"`
}

// AdminCandidateService defines the interface for admin candidate operations
type AdminCandidateService interface {
	AdminListCandidates(ctx context.Context, electionID int64, search string, status *CandidateStatus, page, limit int) ([]CandidateDetailDTO, Pagination, error)
	AdminCreateCandidate(ctx context.Context, electionID int64, req AdminCreateCandidateRequest) (*CandidateDetailDTO, error)
	AdminGetCandidate(ctx context.Context, electionID, candidateID int64) (*CandidateDetailDTO, error)
	AdminUpdateCandidate(ctx context.Context, electionID, candidateID int64, req AdminUpdateCandidateRequest) (*CandidateDetailDTO, error)
	AdminDeleteCandidate(ctx context.Context, electionID, candidateID int64) error
	AdminPublishCandidate(ctx context.Context, electionID, candidateID int64) (*CandidateDetailDTO, error)
	AdminUnpublishCandidate(ctx context.Context, electionID, candidateID int64) (*CandidateDetailDTO, error)
	UploadProfileMedia(ctx context.Context, candidateID int64, media CandidateMediaCreate) (*CandidateMedia, error)
	GetProfileMedia(ctx context.Context, candidateID int64) (*CandidateMedia, error)
	DeleteProfileMedia(ctx context.Context, candidateID, adminID int64) error
	UploadMedia(ctx context.Context, candidateID int64, media CandidateMediaCreate) (*CandidateMedia, error)
	GetMedia(ctx context.Context, candidateID int64, mediaID string) (*CandidateMedia, error)
	DeleteMedia(ctx context.Context, candidateID int64, mediaID string) error
}

// Common admin errors
var (
	ErrElectionNotFound       = errors.New("election not found")
	ErrCandidateNumberTaken   = errors.New("candidate number already used")
	ErrCandidateStatusInvalid = errors.New("candidate status invalid for this action")
)

type AdminHandler struct {
	svc AdminCandidateService
}

func NewAdminHandler(svc AdminCandidateService) *AdminHandler {
	return &AdminHandler{svc: svc}
}

// List menangani GET /admin/elections/{electionID}/candidates
func (h *AdminHandler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseInt64Param(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "electionID tidak valid.")
		return
	}

	q := r.URL.Query()
	search := q.Get("search")
	statusStr := q.Get("status")
	var status *CandidateStatus
	if statusStr != "" {
		cs := CandidateStatus(statusStr)
		status = &cs
	}

	page := parseIntQuery(r, "page", 1)
	limit := parseIntQuery(r, "limit", 20)

	items, pag, err := h.svc.AdminListCandidates(ctx, electionID, search, status, page, limit)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := struct {
		Items      []CandidateDetailDTO `json:"items"`
		Pagination Pagination           `json:"pagination"`
	}{
		Items:      items,
		Pagination: pag,
	}

	response.JSON(w, http.StatusOK, resp)
}

// Create menangani POST /admin/elections/{electionID}/candidates
func (h *AdminHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseInt64Param(r, "electionID")
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "electionID tidak valid.")
		return
	}

	var req AdminCreateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Body tidak valid.")
		return
	}

	// Validasi minimal
	if req.Number <= 0 || req.Name == "" {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "number dan name wajib diisi.")
		return
	}

	dto, err := h.svc.AdminCreateCandidate(ctx, electionID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, dto)
}

// Detail menangani GET /admin/candidates/{candidateID}?election_id=...
func (h *AdminHandler) Detail(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	electionIDParam := r.URL.Query().Get("election_id")
	if electionIDParam == "" {
		response.BadRequest(w, "INVALID_REQUEST", "election_id wajib diisi.")
		return
	}
	electionID, err := strconv.ParseInt(electionIDParam, 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "election_id tidak valid.")
		return
	}

	dto, err := h.svc.AdminGetCandidate(ctx, electionID, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// Update menangani PUT /admin/candidates/{candidateID}?election_id=...
func (h *AdminHandler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	electionIDParam := r.URL.Query().Get("election_id")
	if electionIDParam == "" {
		response.BadRequest(w, "INVALID_REQUEST", "election_id wajib diisi.")
		return
	}
	electionID, err := strconv.ParseInt(electionIDParam, 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "election_id tidak valid.")
		return
	}

	var req AdminUpdateCandidateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Body tidak valid.")
		return
	}

	dto, err := h.svc.AdminUpdateCandidate(ctx, electionID, candidateID, req)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// Delete menangani DELETE /admin/candidates/{candidateID}?election_id=...
func (h *AdminHandler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	electionIDParam := r.URL.Query().Get("election_id")
	if electionIDParam == "" {
		response.BadRequest(w, "INVALID_REQUEST", "election_id wajib diisi.")
		return
	}
	electionID, err := strconv.ParseInt(electionIDParam, 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "election_id tidak valid.")
		return
	}

	if err := h.svc.AdminDeleteCandidate(ctx, electionID, candidateID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Publish menangani POST /admin/candidates/{candidateID}/publish?election_id=...
func (h *AdminHandler) Publish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	electionIDParam := r.URL.Query().Get("election_id")
	if electionIDParam == "" {
		response.BadRequest(w, "INVALID_REQUEST", "election_id wajib diisi.")
		return
	}
	electionID, err := strconv.ParseInt(electionIDParam, 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "election_id tidak valid.")
		return
	}

	dto, err := h.svc.AdminPublishCandidate(ctx, electionID, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// Unpublish menangani POST /admin/candidates/{candidateID}/unpublish?election_id=...
func (h *AdminHandler) Unpublish(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	electionIDParam := r.URL.Query().Get("election_id")
	if electionIDParam == "" {
		response.BadRequest(w, "INVALID_REQUEST", "election_id wajib diisi.")
		return
	}
	electionID, err := strconv.ParseInt(electionIDParam, 10, 64)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "election_id tidak valid.")
		return
	}

	dto, err := h.svc.AdminUnpublishCandidate(ctx, electionID, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	response.JSON(w, http.StatusOK, dto)
}

// UploadProfileMedia handles POST /admin/candidates/{candidateID}/media/profile
func (h *AdminHandler) UploadProfileMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	adminID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}

	if err := r.ParseMultipartForm(maxCandidateMediaSize + (512 << 10)); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Gagal membaca form upload.")
		return
	}

	filePart, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Field file wajib diisi.")
		return
	}
	defer filePart.Close()

	data, tooLarge, err := readCandidateMedia(filePart, maxCandidateMediaSize)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Gagal membaca file upload.")
		return
	}
	if tooLarge {
		response.UnprocessableEntity(w, "FILE_TOO_LARGE", "Ukuran file maksimal 3MB.")
		return
	}

	mime := mimetype.Detect(data)
	if !isAllowedCandidateMedia(CandidateMediaSlotProfile, mime) {
		response.UnprocessableEntity(w, "INVALID_FILE_TYPE", "Foto profil harus berupa PNG atau JPEG.")
		return
	}

	mediaID, err := newCandidateMediaID()
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menyiapkan penyimpanan media.")
		return
	}

	media := CandidateMediaCreate{
		ID:          mediaID,
		Slot:        CandidateMediaSlotProfile,
		FileName:    header.Filename,
		ContentType: mime.String(),
		SizeBytes:   int64(len(data)),
		Data:        data,
		CreatedByID: adminID,
	}

	saved, err := h.svc.UploadProfileMedia(ctx, candidateID, media)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := map[string]interface{}{
		"id":           saved.ID,
		"slot":         saved.Slot,
		"content_type": saved.ContentType,
		"size":         saved.SizeBytes,
	}
	response.JSON(w, http.StatusOK, resp)
}

// GetProfileMedia handles GET /admin/candidates/{candidateID}/media/profile
func (h *AdminHandler) GetProfileMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	media, err := h.svc.GetProfileMedia(ctx, candidateID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	// Return JSON with URL instead of serving BLOB
	response.JSON(w, http.StatusOK, map[string]interface{}{
		"url":          media.URL,
		"content_type": media.ContentType,
		"candidate_id": media.CandidateID,
	})
}

// DeleteProfileMedia handles DELETE /admin/candidates/{candidateID}/media/profile
func (h *AdminHandler) DeleteProfileMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	adminID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}

	if err := h.svc.DeleteProfileMedia(ctx, candidateID, adminID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// UploadMedia handles POST /admin/candidates/{candidateID}/media
func (h *AdminHandler) UploadMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	adminID, ok := ctxkeys.GetUserID(ctx)
	if !ok {
		response.Unauthorized(w, "UNAUTHORIZED", "User tidak valid.")
		return
	}

	if err := r.ParseMultipartForm(maxCandidateMediaSize + (512 << 10)); err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Gagal membaca form upload.")
		return
	}

	slotRaw := r.FormValue("slot")
	slot := CandidateMediaSlotPoster
	if slotRaw != "" {
		slotParsed, err := ParseCandidateMediaSlot(slotRaw)
		if err != nil {
			response.BadRequest(w, "INVALID_REQUEST", "slot tidak valid.")
			return
		}
		if slotParsed == CandidateMediaSlotProfile {
			response.BadRequest(w, "INVALID_REQUEST", "Gunakan endpoint /media/profile untuk foto profil.")
			return
		}
		slot = slotParsed
	}

	filePart, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Field file wajib diisi.")
		return
	}
	defer filePart.Close()

	data, tooLarge, err := readCandidateMedia(filePart, maxCandidateMediaSize)
	if err != nil {
		response.BadRequest(w, "INVALID_REQUEST", "Gagal membaca file upload.")
		return
	}
	if tooLarge {
		response.UnprocessableEntity(w, "FILE_TOO_LARGE", "Ukuran file maksimal 3MB.")
		return
	}

	mime := mimetype.Detect(data)
	if !isAllowedCandidateMedia(slot, mime) {
		response.UnprocessableEntity(w, "INVALID_FILE_TYPE", invalidMediaMessage(slot))
		return
	}

	mediaID, err := newCandidateMediaID()
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menyiapkan penyimpanan media.")
		return
	}

	media := CandidateMediaCreate{
		ID:          mediaID,
		Slot:        slot,
		FileName:    header.Filename,
		ContentType: mime.String(),
		SizeBytes:   int64(len(data)),
		Data:        data,
		CreatedByID: adminID,
	}

	saved, err := h.svc.UploadMedia(ctx, candidateID, media)
	if err != nil {
		h.handleError(w, err)
		return
	}

	resp := map[string]interface{}{
		"id":           saved.ID,
		"slot":         saved.Slot,
		"content_type": saved.ContentType,
		"size":         saved.SizeBytes,
	}
	response.JSON(w, http.StatusOK, resp)
}

// GetMedia handles GET /admin/candidates/{candidateID}/media/{mediaID}
func (h *AdminHandler) GetMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	mediaID := chi.URLParam(r, "mediaID")
	if mediaID == "" {
		response.BadRequest(w, "INVALID_REQUEST", "mediaID tidak valid.")
		return
	}

	media, err := h.svc.GetMedia(ctx, candidateID, mediaID)
	if err != nil {
		h.handleError(w, err)
		return
	}

	w.Header().Set("Content-Type", media.ContentType)
	w.Header().Set("Content-Length", strconv.FormatInt(media.SizeBytes, 10))
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(media.Data)
}

// DeleteMedia handles DELETE /admin/candidates/{candidateID}/media/{mediaID}
func (h *AdminHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	candidateID, err := parseInt64Param(r, "candidateID")
	if err != nil || candidateID <= 0 {
		response.BadRequest(w, "INVALID_REQUEST", "candidateID tidak valid.")
		return
	}

	mediaID := chi.URLParam(r, "mediaID")
	if mediaID == "" {
		response.BadRequest(w, "INVALID_REQUEST", "mediaID tidak valid.")
		return
	}

	if err := h.svc.DeleteMedia(ctx, candidateID, mediaID); err != nil {
		h.handleError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func readCandidateMedia(file io.Reader, maxSize int64) ([]byte, bool, error) {
	limited := io.LimitReader(file, maxSize+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, false, err
	}
	if int64(len(data)) > maxSize {
		return nil, true, nil
	}
	return data, false, nil
}

func isAllowedCandidateMedia(slot CandidateMediaSlot, mime *mimetype.MIME) bool {
	if mime == nil {
		return false
	}
	switch slot {
	case CandidateMediaSlotProfile, CandidateMediaSlotPoster, CandidateMediaSlotPhotoExtra:
		return mime.Is("image/png") || mime.Is("image/jpeg")
	case CandidateMediaSlotPDFProgram, CandidateMediaSlotPDFVisimisi:
		return mime.Is("application/pdf")
	default:
		return false
	}
}

func invalidMediaMessage(slot CandidateMediaSlot) string {
	switch slot {
	case CandidateMediaSlotPDFProgram, CandidateMediaSlotPDFVisimisi:
		return "File harus berupa PDF."
	default:
		return "File harus berupa PNG atau JPEG."
	}
}

// parseInt64Param parses URL parameter as int64
func parseInt64Param(r *http.Request, name string) (int64, error) {
	s := chi.URLParam(r, name)
	return strconv.ParseInt(s, 10, 64)
}

func newCandidateMediaID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16]), nil
}

const maxCandidateMediaSize = int64(3 * 1024 * 1024) // ~3MB

// handleError maps service errors to HTTP responses
func (h *AdminHandler) handleError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, ErrElectionNotFound):
		response.NotFound(w, "NOT_FOUND", "Pemilu tidak ditemukan.")

	case errors.Is(err, ErrCandidateNotFound):
		response.NotFound(w, "NOT_FOUND", "Kandidat tidak ditemukan.")

	case errors.Is(err, ErrCandidateNumberTaken):
		response.Conflict(w, "CANDIDATE_NUMBER_TAKEN", "Nomor kandidat sudah digunakan di pemilu ini.")

	case errors.Is(err, ErrCandidateStatusInvalid):
		response.BadRequest(w, "INVALID_REQUEST", "Perubahan status kandidat tidak diizinkan.")

	case errors.Is(err, ErrCandidateMediaNotFound):
		response.NotFound(w, "MEDIA_NOT_FOUND", "Media kandidat tidak ditemukan.")

	case errors.Is(err, ErrInvalidCandidateMediaSlot):
		response.BadRequest(w, "INVALID_REQUEST", "Slot media tidak valid.")

	default:
		// TODO: log internal error dengan logger
		log.Printf("INTERNAL_ERROR in candidate handler: %v", err)
		// Also log to file for debugging
		if f, ferr := os.OpenFile("/tmp/handler_error.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); ferr == nil {
			defer f.Close()
			f.WriteString(fmt.Sprintf("INTERNAL_ERROR in candidate handler: %v\n", err))
		}
		response.InternalServerError(w, "INTERNAL_ERROR", "Terjadi kesalahan pada sistem.")
	}
}
