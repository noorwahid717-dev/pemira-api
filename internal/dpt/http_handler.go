package dpt

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"pemira-api/internal/http/response"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

func parseElectionID(r *http.Request) (int64, error) {
	s := chi.URLParam(r, "electionID")
	return strconv.ParseInt(s, 10, 64)
}

// POST /admin/elections/{electionID}/voters/import
func (h *Handler) Import(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	if err := r.ParseMultipartForm(10 << 20); err != nil { // 10MB
		response.BadRequest(w, "VALIDATION_ERROR", "Gagal membaca form upload.")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Field file wajib diisi.")
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.TrimLeadingSpace = true

	// Read header
	header, err := reader.Read()
	if err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Gagal membaca header CSV.")
		return
	}

	// Validate header
	headerMap := make(map[string]int)
	for i, col := range header {
		headerMap[col] = i
	}
	requiredCols := []string{"nim", "name", "faculty", "study_program", "cohort_year"}
	for _, col := range requiredCols {
		if _, ok := headerMap[col]; !ok {
			response.UnprocessableEntity(w, "VALIDATION_ERROR", fmt.Sprintf("Kolom '%s' wajib ada di CSV.", col))
			return
		}
	}

	// Read rows
	var rows []ImportRow
	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			response.BadRequest(w, "VALIDATION_ERROR", "CSV tidak valid.")
			return
		}

		cohortStr := record[headerMap["cohort_year"]]
		cohortYear, err := strconv.Atoi(cohortStr)
		if err != nil {
			response.UnprocessableEntity(w, "VALIDATION_ERROR", "cohort_year harus angka.")
			return
		}

		row := ImportRow{
			NIM:          record[headerMap["nim"]],
			Name:         record[headerMap["name"]],
			FacultyName:  record[headerMap["faculty"]],
			StudyProgram: record[headerMap["study_program"]],
			CohortYear:   cohortYear,
		}

		if col, ok := headerMap["email"]; ok && col < len(record) {
			row.Email = record[col]
		}
		if col, ok := headerMap["phone"]; ok && col < len(record) {
			row.Phone = record[col]
		}

		rows = append(rows, row)
	}

	if len(rows) == 0 {
		response.UnprocessableEntity(w, "VALIDATION_ERROR", "CSV tidak berisi data.")
		return
	}

	result, err := h.svc.Import(ctx, electionID, rows)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengimpor DPT.")
		return
	}

	response.JSON(w, http.StatusOK, result)
}

// GET /admin/elections/{electionID}/voters
func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	q := r.URL.Query()
	filter := ListFilter{
		Faculty:      q.Get("faculty"),
		StudyProgram: q.Get("study_program"),
		Search:       q.Get("search"),
	}

	if cyStr := q.Get("cohort_year"); cyStr != "" {
		if cy, err := strconv.Atoi(cyStr); err == nil {
			filter.CohortYear = &cy
		}
	}

	if hvStr := q.Get("has_voted"); hvStr != "" {
		b, err := strconv.ParseBool(hvStr)
		if err == nil {
			filter.HasVoted = &b
		}
	}

	if elStr := q.Get("eligible"); elStr != "" {
		b, err := strconv.ParseBool(elStr)
		if err == nil {
			filter.Eligible = &b
		}
	}

	page := parseIntDefault(q.Get("page"), 1)
	limit := parseIntDefault(q.Get("limit"), 50)

	items, pag, err := h.svc.List(ctx, electionID, filter, page, limit)
	if err != nil {
		slog.Error("failed to list voters", "error", err, "election_id", electionID)
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil daftar pemilih.")
		return
	}

	resp := struct {
		Items      []VoterWithStatusDTO `json:"items"`
		Pagination Pagination           `json:"pagination"`
	}{
		Items:      items,
		Pagination: pag,
	}

	response.JSON(w, http.StatusOK, resp)
}

// GET /admin/elections/{electionID}/voters/export
func (h *Handler) Export(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	q := r.URL.Query()
	filter := ListFilter{
		Faculty:      q.Get("faculty"),
		StudyProgram: q.Get("study_program"),
		Search:       q.Get("search"),
	}
	if cyStr := q.Get("cohort_year"); cyStr != "" {
		if cy, err := strconv.Atoi(cyStr); err == nil {
			filter.CohortYear = &cy
		}
	}
	if hvStr := q.Get("has_voted"); hvStr != "" {
		if b, err := strconv.ParseBool(hvStr); err == nil {
			filter.HasVoted = &b
		}
	}
	if elStr := q.Get("eligible"); elStr != "" {
		if b, err := strconv.ParseBool(elStr); err == nil {
			filter.Eligible = &b
		}
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="dpt_election_%d.csv"`, electionID))

	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"nim",
		"name",
		"faculty",
		"study_program",
		"cohort_year",
		"email",
		"has_voted",
		"last_vote_channel",
		"last_vote_at",
		"last_tps_id",
		"is_eligible",
	}
	if err := writer.Write(header); err != nil {
		return
	}

	err = h.svc.ExportStream(ctx, electionID, filter, func(v VoterWithStatusDTO) error {
		hasVoted := strconv.FormatBool(v.Status.HasVoted)
		isEligible := strconv.FormatBool(v.Status.IsEligible)

		lastChannel := ""
		if v.Status.LastVoteChannel != nil {
			lastChannel = *v.Status.LastVoteChannel
		}
		lastVoteAt := ""
		if v.Status.LastVoteAt != nil {
			lastVoteAt = v.Status.LastVoteAt.UTC().Format(time.RFC3339)
		}
		lastTPSID := ""
		if v.Status.LastTPSID != nil {
			lastTPSID = strconv.FormatInt(*v.Status.LastTPSID, 10)
		}

		cohortYear := ""
		if v.CohortYear != nil {
			cohortYear = strconv.Itoa(*v.CohortYear)
		}

		record := []string{
			v.NIM,
			v.Name,
			v.FacultyName,
			v.StudyProgramName,
			cohortYear,
			v.Email,
			hasVoted,
			lastChannel,
			lastVoteAt,
			lastTPSID,
			isEligible,
		}
		return writer.Write(record)
	})

	if err != nil {
		return
	}
}

// GET /admin/elections/{electionID}/voters/{voterID}
func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	voterIDStr := chi.URLParam(r, "voterID")
	voterID, err := strconv.ParseInt(voterIDStr, 10, 64)
	if err != nil || voterID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "voterID tidak valid.")
		return
	}

	voter, err := h.svc.GetVoterByID(ctx, electionID, voterID)
	if err != nil {
		slog.Error("failed to get voter", "error", err, "voter_id", voterID)
		response.NotFound(w, "VOTER_NOT_FOUND", "Pemilih tidak ditemukan.")
		return
	}

	response.JSON(w, http.StatusOK, voter)
}

// PUT /admin/elections/{electionID}/voters/{voterID}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	voterIDStr := chi.URLParam(r, "voterID")
	voterID, err := strconv.ParseInt(voterIDStr, 10, 64)
	if err != nil || voterID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "voterID tidak valid.")
		return
	}

	var updates VoterUpdateDTO
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		response.BadRequest(w, "VALIDATION_ERROR", "Request body tidak valid.")
		return
	}

	if updates.VotingMethod != nil {
		method := strings.ToUpper(strings.TrimSpace(*updates.VotingMethod))
		if method != "ONLINE" && method != "TPS" {
			response.UnprocessableEntity(w, "INVALID_VOTING_METHOD", "voting_method harus ONLINE atau TPS.")
			return
		}
		updates.VotingMethod = &method
	}

	if err := h.svc.UpdateVoter(ctx, electionID, voterID, updates); err != nil {
		slog.Error("failed to update voter", "error", err, "voter_id", voterID)

		errMsg := err.Error()
		if errMsg == "voter not found in this election" {
			response.NotFound(w, "VOTER_NOT_FOUND", "Pemilih tidak ditemukan.")
			return
		}
		if errMsg == "cannot update voter who has already voted" {
			response.Forbidden(w, "VOTER_HAS_VOTED", "Tidak dapat mengubah data pemilih yang sudah memilih.")
			return
		}

		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengupdate pemilih.")
		return
	}

	// Get updated voter
	voter, err := h.svc.GetVoterByID(ctx, electionID, voterID)
	if err != nil {
		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal mengambil data pemilih.")
		return
	}

	response.JSON(w, http.StatusOK, voter)
}

// DELETE /admin/elections/{electionID}/voters/{voterID}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	electionID, err := parseElectionID(r)
	if err != nil || electionID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "electionID tidak valid.")
		return
	}

	voterIDStr := chi.URLParam(r, "voterID")
	voterID, err := strconv.ParseInt(voterIDStr, 10, 64)
	if err != nil || voterID <= 0 {
		response.BadRequest(w, "VALIDATION_ERROR", "voterID tidak valid.")
		return
	}

	if err := h.svc.DeleteVoter(ctx, electionID, voterID); err != nil {
		slog.Error("failed to delete voter", "error", err, "voter_id", voterID)

		errMsg := err.Error()
		if errMsg == "voter not found in this election" {
			response.NotFound(w, "VOTER_NOT_FOUND", "Pemilih tidak ditemukan.")
			return
		}
		if errMsg == "cannot delete voter who has already voted" {
			response.Forbidden(w, "VOTER_HAS_VOTED", "Tidak dapat menghapus pemilih yang sudah memilih.")
			return
		}

		response.InternalServerError(w, "INTERNAL_ERROR", "Gagal menghapus pemilih.")
		return
	}

	response.JSON(w, http.StatusOK, map[string]string{
		"message": "Pemilih berhasil dihapus",
	})
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
