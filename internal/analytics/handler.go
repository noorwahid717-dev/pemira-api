package analytics

import (
"net/http"
"strconv"

"github.com/gin-gonic/gin"
)

// Handler handles HTTP requests for analytics endpoints
type Handler struct {
service *Service
}

// NewHandler creates a new analytics handler
func NewHandler(service *Service) *Handler {
return &Handler{service: service}
}

// RegisterRoutes registers analytics routes
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
analytics := r.Group("/analytics")
{
analytics.GET("/elections/:id/dashboard", h.GetDashboardCharts)
analytics.GET("/elections/:id/hourly-votes", h.GetHourlyVotesByChannel)
analytics.GET("/elections/:id/hourly-by-candidate", h.GetHourlyVotesByCandidate)
analytics.GET("/elections/:id/faculty-heatmap", h.GetFacultyCandidateHeatmap)
analytics.GET("/elections/:id/turnout-timeline", h.GetTurnoutTimeline)
analytics.GET("/elections/:id/cohort-breakdown", h.GetCohortCandidateVotes)
analytics.GET("/elections/:id/peak-hours", h.GetPeakHours)
analytics.GET("/elections/:id/voting-velocity", h.GetVotingVelocity)
}
}

// GetDashboardCharts handles GET /analytics/elections/:id/dashboard
func (h *Handler) GetDashboardCharts(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetDashboardCharts(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetHourlyVotesByChannel handles GET /analytics/elections/:id/hourly-votes
func (h *Handler) GetHourlyVotesByChannel(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetHourlyVotesByChannel(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetHourlyVotesByCandidate handles GET /analytics/elections/:id/hourly-by-candidate
func (h *Handler) GetHourlyVotesByCandidate(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetHourlyVotesByCandidate(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetFacultyCandidateHeatmap handles GET /analytics/elections/:id/faculty-heatmap
func (h *Handler) GetFacultyCandidateHeatmap(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetFacultyCandidateHeatmap(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetTurnoutTimeline handles GET /analytics/elections/:id/turnout-timeline
func (h *Handler) GetTurnoutTimeline(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetTurnoutTimeline(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetCohortCandidateVotes handles GET /analytics/elections/:id/cohort-breakdown
func (h *Handler) GetCohortCandidateVotes(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetCohortCandidateVotes(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetPeakHours handles GET /analytics/elections/:id/peak-hours
func (h *Handler) GetPeakHours(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetPeakHours(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}

// GetVotingVelocity handles GET /analytics/elections/:id/voting-velocity
func (h *Handler) GetVotingVelocity(c *gin.Context) {
electionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
if err != nil {
c.JSON(http.StatusBadRequest, gin.H{"error": "invalid election ID"})
return
}

result, err := h.service.GetVotingVelocity(c.Request.Context(), electionID)
if err != nil {
c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
return
}

c.JSON(http.StatusOK, result)
}
