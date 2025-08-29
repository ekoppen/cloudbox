package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// APIStatsHandler handles API usage statistics
type APIStatsHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAPIStatsHandler creates a new API stats handler
func NewAPIStatsHandler(db *gorm.DB, cfg *config.Config) *APIStatsHandler {
	return &APIStatsHandler{
		db:  db,
		cfg: cfg,
	}
}

// APIRouteStatsResponse represents the response for API route statistics
type APIRouteStatsResponse struct {
	ProjectID uint                    `json:"project_id"`
	Routes    []APIRouteStatsItem     `json:"routes"`
	Summary   APIStatsSummary         `json:"summary"`
	Timeline  []APIStatsTimelineItem  `json:"timeline"`
}

// APIRouteStatsItem represents statistics for a single API route
type APIRouteStatsItem struct {
	Method            string  `json:"method"`
	Endpoint          string  `json:"endpoint"`
	TotalRequests     int     `json:"total_requests"`
	SuccessRequests   int     `json:"success_requests"`
	ErrorRequests     int     `json:"error_requests"`
	SuccessRate       float64 `json:"success_rate"`
	AvgResponseTime   float64 `json:"avg_response_time_ms"`
	TotalDataTransfer int64   `json:"total_data_transfer_bytes"`
	LastUsed          *time.Time `json:"last_used"`
}

// APIStatsSummary represents overall API usage summary
type APIStatsSummary struct {
	TotalRequests     int     `json:"total_requests"`
	TotalEndpoints    int     `json:"total_endpoints"`
	OverallSuccessRate float64 `json:"overall_success_rate"`
	AvgResponseTime   float64 `json:"avg_response_time_ms"`
	TotalDataTransfer int64   `json:"total_data_transfer_bytes"`
	TopEndpoint       string  `json:"top_endpoint"`
	TopEndpointCount  int     `json:"top_endpoint_count"`
}

// APIStatsTimelineItem represents API usage over time
type APIStatsTimelineItem struct {
	Date            string `json:"date"`
	TotalRequests   int    `json:"total_requests"`
	SuccessRequests int    `json:"success_requests"`
	ErrorRequests   int    `json:"error_requests"`
	AvgResponseTime float64 `json:"avg_response_time_ms"`
}

// GetProjectAPIStats returns detailed API usage statistics for a project
func (h *APIStatsHandler) GetProjectAPIStats(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get date range from query params (default to last 30 days)
	daysStr := c.DefaultQuery("days", "30")
	days, err := strconv.Atoi(daysStr)
	if err != nil || days < 1 || days > 365 {
		days = 30
	}

	startDate := time.Now().AddDate(0, 0, -days).Format("2006-01-02")

	var response APIRouteStatsResponse
	response.ProjectID = uint(projectID)

	// Get route statistics grouped by method and endpoint
	type RouteStatsQuery struct {
		Method            string  `json:"method"`
		Endpoint          string  `json:"endpoint"`
		TotalRequests     int     `json:"total_requests"`
		SuccessRequests   int     `json:"success_requests"`
		ErrorRequests     int     `json:"error_requests"`
		AvgResponseTime   float64 `json:"avg_response_time_ms"`
		TotalDataTransfer int64   `json:"total_data_transfer_bytes"`
	}

	var routeStats []RouteStatsQuery
	h.db.Model(&models.APIRouteStats{}).
		Select(`method, endpoint, 
			    SUM(total_requests) as total_requests,
			    SUM(success_requests) as success_requests, 
			    SUM(error_requests) as error_requests,
			    AVG(avg_response_time_ms) as avg_response_time_ms,
			    SUM(total_response_size_bytes) as total_data_transfer_bytes`).
		Where("project_id = ? AND date >= ?", projectID, startDate).
		Group("method, endpoint").
		Order("total_requests DESC").
		Scan(&routeStats)

	// Convert to response format and calculate success rates
	response.Routes = make([]APIRouteStatsItem, len(routeStats))
	for i, stat := range routeStats {
		var successRate float64
		if stat.TotalRequests > 0 {
			successRate = float64(stat.SuccessRequests) / float64(stat.TotalRequests) * 100
		}

		// Get last used timestamp
		var lastUsed time.Time
		h.db.Model(&models.APIRequestLog{}).
			Select("created_at").
			Where("project_id = ? AND method = ? AND endpoint = ?", projectID, stat.Method, stat.Endpoint).
			Order("created_at DESC").
			Limit(1).
			Scan(&lastUsed)

		response.Routes[i] = APIRouteStatsItem{
			Method:            stat.Method,
			Endpoint:          stat.Endpoint,
			TotalRequests:     stat.TotalRequests,
			SuccessRequests:   stat.SuccessRequests,
			ErrorRequests:     stat.ErrorRequests,
			SuccessRate:       successRate,
			AvgResponseTime:   stat.AvgResponseTime,
			TotalDataTransfer: stat.TotalDataTransfer,
			LastUsed:          &lastUsed,
		}
	}

	// Calculate summary statistics
	var totalRequests, totalSuccess, totalErrors int
	var totalDataTransfer int64
	var avgResponseTime float64
	var topEndpoint string
	var topEndpointCount int

	for _, route := range response.Routes {
		totalRequests += route.TotalRequests
		totalSuccess += route.SuccessRequests
		totalErrors += route.ErrorRequests
		totalDataTransfer += route.TotalDataTransfer
		avgResponseTime += route.AvgResponseTime * float64(route.TotalRequests)
		
		if route.TotalRequests > topEndpointCount {
			topEndpointCount = route.TotalRequests
			topEndpoint = route.Method + " " + route.Endpoint
		}
	}

	var overallSuccessRate float64
	if totalRequests > 0 {
		overallSuccessRate = float64(totalSuccess) / float64(totalRequests) * 100
		avgResponseTime = avgResponseTime / float64(totalRequests)
	}

	response.Summary = APIStatsSummary{
		TotalRequests:     totalRequests,
		TotalEndpoints:    len(response.Routes),
		OverallSuccessRate: overallSuccessRate,
		AvgResponseTime:   avgResponseTime,
		TotalDataTransfer: totalDataTransfer,
		TopEndpoint:       topEndpoint,
		TopEndpointCount:  topEndpointCount,
	}

	// Get timeline data (daily stats for the period)
	var timelineStats []APIStatsTimelineItem
	h.db.Model(&models.APIRouteStats{}).
		Select(`date, 
				SUM(total_requests) as total_requests,
				SUM(success_requests) as success_requests,
				SUM(error_requests) as error_requests,
				AVG(avg_response_time_ms) as avg_response_time_ms`).
		Where("project_id = ? AND date >= ?", projectID, startDate).
		Group("date").
		Order("date ASC").
		Scan(&timelineStats)

	// Convert timeline to proper format
	response.Timeline = make([]APIStatsTimelineItem, len(timelineStats))
	for i, stat := range timelineStats {
		response.Timeline[i] = APIStatsTimelineItem{
			Date:            stat.Date,
			TotalRequests:   stat.TotalRequests,
			SuccessRequests: stat.SuccessRequests,
			ErrorRequests:   stat.ErrorRequests,
			AvgResponseTime: stat.AvgResponseTime,
		}
	}

	c.JSON(http.StatusOK, response)
}

// GetRecentAPILogs returns recent API request logs for a project
func (h *APIStatsHandler) GetRecentAPILogs(c *gin.Context) {
	projectIDStr := c.Param("id")
	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get limit from query params (default to 50, max 1000)
	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 1000 {
		limit = 50
	}

	var logs []models.APIRequestLog
	h.db.Where("project_id = ?", projectID).
		Order("created_at DESC").
		Limit(limit).
		Find(&logs)

	c.JSON(http.StatusOK, gin.H{
		"logs":    logs,
		"count":   len(logs),
		"project_id": projectID,
	})
}