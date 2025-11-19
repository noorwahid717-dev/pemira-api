package analytics

import (
"context"
"errors"
"net/http"
"strconv"

"github.com/go-chi/chi/v5"
)

// Response helper interface (compatible with internal/http/response)
type ResponseWriter interface {
Success(w http.ResponseWriter, statusCode int, data interface{})
BadRequest(w http.ResponseWriter, message string, details interface{})
InternalServerError(w http.ResponseWriter, message string)
NotFound(w http.ResponseWriter, message string)
}

// AnalyticsService defines the interface for analytics operations
type AnalyticsService interface {
GetDashboardCharts(ctx context.Context, electionID int64) (*DashboardCharts, error)
GetHourlyVotesByChannel(ctx context.Context, electionID int64) ([]HourlyVotes, error)
GetHourlyVotesByCandidate(ctx context.Context, electionID int64) ([]HourlyCandidateVotes, error)
GetTurnoutTimeline(ctx context.Context, electionID int64) ([]TurnoutPoint, error)
GetFacultyCandidateHeatmap(ctx context.Context, electionID int64) ([]FacultyCandidateHeatmapRow, error)
GetCohortCandidateVotes(ctx context.Context, electionID int64) ([]CohortCandidateVotes, error)
GetPeakHours(ctx context.Context, electionID int64) ([]PeakHour, error)
GetVotingVelocity(ctx context.Context, electionID int64) (*VotingVelocity, error)
}

// Handler handles HTTP requests for analytics endpoints
type Handler struct {
svc AnalyticsService
res ResponseWriter
}

// NewHandler creates a new analytics handler
func NewHandler(svc AnalyticsService, res ResponseWriter) *Handler {
return &Handler{
svc: svc,
res: res,
}
}

// Mount registers analytics routes to the given chi router
// Expected to be mounted at: /admin/elections/{electionID}/analytics
func (h *Handler) Mount(r chi.Router) {
// GET /admin/elections/{electionID}/analytics/dashboard
r.Get("/dashboard", h.GetDashboardCharts)

// GET /admin/elections/{electionID}/analytics/timeline/votes
r.Get("/timeline/votes", h.GetHourlyVotesByChannel)

// GET /admin/elections/{electionID}/analytics/timeline/candidates
r.Get("/timeline/candidates", h.GetHourlyVotesByCandidate)

// GET /admin/elections/{electionID}/analytics/timeline/turnout
r.Get("/timeline/turnout", h.GetTurnoutTimeline)

// GET /admin/elections/{electionID}/analytics/heatmap/faculty-candidate
r.Get("/heatmap/faculty-candidate", h.GetFacultyCandidateHeatmap)

// GET /admin/elections/{electionID}/analytics/cohort-breakdown
r.Get("/cohort-breakdown", h.GetCohortCandidateVotes)

// GET /admin/elections/{electionID}/analytics/peak-hours
r.Get("/peak-hours", h.GetPeakHours)

// GET /admin/elections/{electionID}/analytics/voting-velocity
r.Get("/voting-velocity", h.GetVotingVelocity)
}

// parseElectionID extracts and validates electionID from URL parameters
func parseElectionID(r *http.Request) (int64, error) {
electionIDStr := chi.URLParam(r, "electionID")
return strconv.ParseInt(electionIDStr, 10, 64)
}

// GetDashboardCharts handles GET /dashboard
func (h *Handler) GetDashboardCharts(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

result, err := h.svc.GetDashboardCharts(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, result)
}

// GetHourlyVotesByChannel handles GET /timeline/votes
func (h *Handler) GetHourlyVotesByChannel(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetHourlyVotesByChannel(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetHourlyVotesByCandidate handles GET /timeline/candidates
func (h *Handler) GetHourlyVotesByCandidate(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetHourlyVotesByCandidate(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetTurnoutTimeline handles GET /timeline/turnout
func (h *Handler) GetTurnoutTimeline(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetTurnoutTimeline(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetFacultyCandidateHeatmap handles GET /heatmap/faculty-candidate
func (h *Handler) GetFacultyCandidateHeatmap(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetFacultyCandidateHeatmap(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetCohortCandidateVotes handles GET /cohort-breakdown
func (h *Handler) GetCohortCandidateVotes(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetCohortCandidateVotes(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetPeakHours handles GET /peak-hours
func (h *Handler) GetPeakHours(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetPeakHours(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// GetVotingVelocity handles GET /voting-velocity
func (h *Handler) GetVotingVelocity(w http.ResponseWriter, r *http.Request) {
ctx := r.Context()

electionID, err := parseElectionID(r)
if err != nil || electionID <= 0 {
h.res.BadRequest(w, "electionID tidak valid.", nil)
return
}

data, err := h.svc.GetVotingVelocity(ctx, electionID)
if err != nil {
h.handleError(w, err)
return
}

h.res.Success(w, http.StatusOK, data)
}

// handleError maps service errors to HTTP responses
func (h *Handler) handleError(w http.ResponseWriter, err error) {
// Common error types (adjust based on your domain errors)
var notFoundErr interface{ NotFound() bool }

switch {
case errors.As(err, &notFoundErr):
h.res.NotFound(w, "Pemilu tidak ditemukan.")

default:
// Log internal error here if needed
h.res.InternalServerError(w, "Terjadi kesalahan pada sistem.")
}
}
