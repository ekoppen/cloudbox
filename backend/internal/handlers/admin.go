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

// AdminHandler handles admin-specific operations
type AdminHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(db *gorm.DB, cfg *config.Config) *AdminHandler {
	return &AdminHandler{db: db, cfg: cfg}
}

// SystemStatsResponse represents system statistics
type SystemStatsResponse struct {
	TotalUsers       int64 `json:"total_users"`
	ActiveUsers      int64 `json:"active_users"`
	InactiveUsers    int64 `json:"inactive_users"`
	TotalProjects    int64 `json:"total_projects"`
	TotalDeployments int64 `json:"total_deployments"`
	TotalFunctions   int64 `json:"total_functions"`
	TotalDocuments   int64 `json:"total_documents"`
	TotalFiles       int64 `json:"total_files"`
	TotalStorageSize int64 `json:"total_storage_size"`
}

// UserGrowthResponse represents user growth data
type UserGrowthResponse struct {
	Date      string `json:"date"`
	NewUsers  int64  `json:"new_users"`
	TotalUsers int64 `json:"total_users"`
}

// ProjectActivityResponse represents project activity data
type ProjectActivityResponse struct {
	Date     string `json:"date"`
	Created  int64  `json:"created"`
	Deployed int64  `json:"deployed"`
	Updated  int64  `json:"updated"`
	Accessed int64  `json:"accessed"`
}

// FunctionExecutionResponse represents function execution data
type FunctionExecutionResponse struct {
	Hour            string  `json:"hour"`
	Executions      int64   `json:"executions"`
	Errors          int64   `json:"errors"`
	AvgResponseTime float64 `json:"avg_response_time"`
}

// DeploymentStatusResponse represents deployment status data
type DeploymentStatusResponse struct {
	Status string `json:"status"`
	Count  int64  `json:"count"`
	Color  string `json:"color"`
}

// StorageStatsResponse represents storage statistics
type StorageStatsResponse struct {
	Type  string `json:"type"`
	Count int64  `json:"count"`
	Size  int64  `json:"size"`
	Color string `json:"color"`
}

// SystemHealthResponse represents system health metrics
type SystemHealthResponse struct {
	CPUUsage          float64 `json:"cpu_usage"`
	MemoryUsage       float64 `json:"memory_usage"`
	DiskUsage         float64 `json:"disk_usage"`
	APIResponseTime   float64 `json:"api_response_time"`
	ActiveConnections int64   `json:"active_connections"`
}

// GetSystemStats returns comprehensive system statistics
func (h *AdminHandler) GetSystemStats(c *gin.Context) {
	var stats SystemStatsResponse

	// Get user statistics
	h.db.Model(&models.User{}).Count(&stats.TotalUsers)
	h.db.Model(&models.User{}).Where("is_active = ?", true).Count(&stats.ActiveUsers)
	stats.InactiveUsers = stats.TotalUsers - stats.ActiveUsers

	// Get project statistics
	h.db.Model(&models.Project{}).Count(&stats.TotalProjects)

	// Get deployment statistics
	h.db.Model(&models.Deployment{}).Count(&stats.TotalDeployments)

	// Get function statistics
	h.db.Model(&models.Function{}).Count(&stats.TotalFunctions)

	// Get document statistics (simulate for now)
	var collectionCount int64
	h.db.Model(&models.Collection{}).Count(&collectionCount)
	stats.TotalDocuments = collectionCount * 25 // Simulate 25 docs per collection

	// Get storage statistics
	h.db.Model(&models.File{}).Count(&stats.TotalFiles)
	
	// Calculate total storage size
	var totalSize int64
	h.db.Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	stats.TotalStorageSize = totalSize

	c.JSON(http.StatusOK, stats)
}

// GetUserGrowth returns user growth statistics
func (h *AdminHandler) GetUserGrowth(c *gin.Context) {
	days := c.DefaultQuery("days", "30")
	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt <= 0 {
		daysInt = 30
	}

	var growth []UserGrowthResponse

	// Get growth data for the specified number of days
	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		var newUsers int64
		h.db.Model(&models.User{}).
			Where("DATE(created_at) = ?", dateStr).
			Count(&newUsers)
		
		var totalUsers int64
		h.db.Model(&models.User{}).
			Where("created_at <= ?", date.Add(24*time.Hour)).
			Count(&totalUsers)
		
		growth = append(growth, UserGrowthResponse{
			Date:       dateStr,
			NewUsers:   newUsers,
			TotalUsers: totalUsers,
		})
	}

	c.JSON(http.StatusOK, growth)
}

// GetProjectActivity returns project activity statistics
func (h *AdminHandler) GetProjectActivity(c *gin.Context) {
	days := c.DefaultQuery("days", "7")
	daysInt, err := strconv.Atoi(days)
	if err != nil || daysInt <= 0 {
		daysInt = 7
	}

	var activity []ProjectActivityResponse

	for i := daysInt - 1; i >= 0; i-- {
		date := time.Now().AddDate(0, 0, -i)
		dateStr := date.Format("2006-01-02")
		
		var created, deployed, updated int64
		
		// Count projects created on this date
		h.db.Model(&models.Project{}).
			Where("DATE(created_at) = ?", dateStr).
			Count(&created)
		
		// Count deployments on this date
		h.db.Model(&models.Deployment{}).
			Where("DATE(deployed_at) = ?", dateStr).
			Count(&deployed)
		
		// Count projects updated on this date
		h.db.Model(&models.Project{}).
			Where("DATE(updated_at) = ?", dateStr).
			Count(&updated)
		
		// Simulate access count
		accessed := created*10 + deployed*5 + updated*2
		
		activity = append(activity, ProjectActivityResponse{
			Date:     dateStr,
			Created:  created,
			Deployed: deployed,
			Updated:  updated,
			Accessed: accessed,
		})
	}

	c.JSON(http.StatusOK, activity)
}

// GetFunctionExecutions returns function execution statistics
func (h *AdminHandler) GetFunctionExecutions(c *gin.Context) {
	hours := c.DefaultQuery("hours", "24")
	hoursInt, err := strconv.Atoi(hours)
	if err != nil || hoursInt <= 0 {
		hoursInt = 24
	}

	var executions []FunctionExecutionResponse

	for i := hoursInt - 1; i >= 0; i-- {
		hour := time.Now().Add(time.Duration(-i) * time.Hour)
		hourStr := hour.Format("15:04")
		
		var executionCount, errorCount int64
		var avgResponseTime float64
		
		// Count function executions in this hour
		h.db.Model(&models.FunctionExecution{}).
			Where("created_at >= ? AND created_at < ?", 
				hour, hour.Add(time.Hour)).
			Count(&executionCount)
		
		// Count errors in this hour (status != 'success')
		h.db.Model(&models.FunctionExecution{}).
			Where("created_at >= ? AND created_at < ? AND status != ?", 
				hour, hour.Add(time.Hour), "success").
			Count(&errorCount)
		
		// Calculate average response time
		h.db.Model(&models.FunctionExecution{}).
			Where("created_at >= ? AND created_at < ?", 
				hour, hour.Add(time.Hour)).
			Select("COALESCE(AVG(execution_time), 0)").
			Scan(&avgResponseTime)
		
		executions = append(executions, FunctionExecutionResponse{
			Hour:            hourStr,
			Executions:      executionCount,
			Errors:          errorCount,
			AvgResponseTime: avgResponseTime,
		})
	}

	c.JSON(http.StatusOK, executions)
}

// GetDeploymentStats returns deployment status distribution
func (h *AdminHandler) GetDeploymentStats(c *gin.Context) {
	type StatusCount struct {
		Status string
		Count  int64
	}
	
	var statusCounts []StatusCount
	h.db.Model(&models.Deployment{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&statusCounts)

	// Map status to colors
	statusColors := map[string]string{
		"deployed": "#10B981", // green
		"building": "#F59E0B", // yellow
		"failed":   "#EF4444", // red
		"pending":  "#6B7280", // gray
	}

	var stats []DeploymentStatusResponse
	for _, sc := range statusCounts {
		color, exists := statusColors[sc.Status]
		if !exists {
			color = "#6B7280" // default gray
		}
		
		stats = append(stats, DeploymentStatusResponse{
			Status: sc.Status,
			Count:  sc.Count,
			Color:  color,
		})
	}

	c.JSON(http.StatusOK, stats)
}

// GetStorageStats returns storage statistics by file type
func (h *AdminHandler) GetStorageStats(c *gin.Context) {
	type FileTypeStats struct {
		MimeType string
		Count    int64
		Size     int64
	}
	
	var typeStats []FileTypeStats
	h.db.Model(&models.File{}).
		Select("mime_type, COUNT(*) as count, COALESCE(SUM(size), 0) as size").
		Group("mime_type").
		Scan(&typeStats)

	// Group similar types and assign colors
	statsMap := map[string]*StorageStatsResponse{
		"Images":    {Type: "Images", Color: "#3B82F6"},
		"Documents": {Type: "Documents", Color: "#10B981"},
		"Videos":    {Type: "Videos", Color: "#F59E0B"},
		"Other":     {Type: "Other", Color: "#6B7280"},
	}

	for _, ts := range typeStats {
		category := "Other"
		if ts.MimeType != "" {
			switch {
			case ts.MimeType[:5] == "image":
				category = "Images"
			case ts.MimeType[:5] == "video":
				category = "Videos"
			case ts.MimeType == "application/pdf" || 
				 ts.MimeType == "application/msword" ||
				 ts.MimeType[:4] == "text":
				category = "Documents"
			}
		}
		
		stat := statsMap[category]
		stat.Count += ts.Count
		stat.Size += ts.Size
	}

	// Convert map to slice
	var stats []StorageStatsResponse
	for _, stat := range statsMap {
		if stat.Count > 0 {
			stats = append(stats, *stat)
		}
	}

	c.JSON(http.StatusOK, stats)
}

// GetSystemHealth returns system health metrics
func (h *AdminHandler) GetSystemHealth(c *gin.Context) {
	// For demo purposes, return simulated health metrics
	// In production, you'd integrate with actual system monitoring
	health := SystemHealthResponse{
		CPUUsage:          45.2,
		MemoryUsage:       67.8,
		DiskUsage:         34.1,
		APIResponseTime:   125.5,
		ActiveConnections: 142,
	}

	c.JSON(http.StatusOK, health)
}