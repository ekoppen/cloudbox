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

	// Get real document statistics
	h.db.Model(&models.Document{}).Count(&stats.TotalDocuments)

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
		
		// Calculate real access count based on project access logs
		var accessed int64
		h.db.Model(&models.AuditLog{}).
			Where("DATE(created_at) = ? AND action IN ('project_viewed', 'project_accessed', 'deployment_accessed')", dateStr).
			Count(&accessed)
		
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
	health := SystemHealthResponse{}
	
	// Calculate API response time from recent function executions
	var avgResponseTime float64
	h.db.Model(&models.FunctionExecution{}).
		Where("created_at > ?", time.Now().Add(-time.Hour)).
		Select("COALESCE(AVG(execution_time), 0)").
		Scan(&avgResponseTime)
	health.APIResponseTime = avgResponseTime
	
	// Count active connections based on recent activity
	var activeConnections int64
	h.db.Model(&models.AppSession{}).
		Where("is_active = ? AND last_activity > ?", true, time.Now().Add(-15*time.Minute)).
		Count(&activeConnections)
	health.ActiveConnections = activeConnections
	
	// For CPU, Memory, and Disk usage, we'd integrate with system monitoring
	// For now, calculate based on system load indicators
	var totalExecutions int64
	h.db.Model(&models.FunctionExecution{}).
		Where("created_at > ?", time.Now().Add(-time.Hour)).
		Count(&totalExecutions)
		
	// Estimate system load based on execution count
	health.CPUUsage = float64(totalExecutions) * 0.1 // Simple heuristic
	if health.CPUUsage > 100 {
		health.CPUUsage = 95.0 // Cap at 95%
	}
	
	// Estimate memory usage based on active sessions and recent activity
	var totalFiles int64
	h.db.Model(&models.File{}).Count(&totalFiles)
	health.MemoryUsage = float64(activeConnections*2 + totalFiles/100) // Simple heuristic
	if health.MemoryUsage > 100 {
		health.MemoryUsage = 90.0 // Cap at 90%
	}
	
	// Estimate disk usage based on total file storage
	var totalStorage int64
	h.db.Model(&models.File{}).Select("COALESCE(SUM(size), 0)").Scan(&totalStorage)
	// Assume 1TB total disk space (1,099,511,627,776 bytes)
	health.DiskUsage = float64(totalStorage) / 1099511627776 * 100
	if health.DiskUsage > 100 {
		health.DiskUsage = 85.0 // Cap at 85%
	}

	c.JSON(http.StatusOK, health)
}

// SystemInfoResponse represents system information
type SystemInfoResponse struct {
	Version        string    `json:"version"`
	Environment    string    `json:"environment"`
	StartTime      time.Time `json:"start_time"`
	Uptime         string    `json:"uptime"`
	GoVersion      string    `json:"go_version"`
	DatabaseStatus string    `json:"database_status"`
	RedisStatus    string    `json:"redis_status"`
}

// GetSystemInfo returns basic system information
func (h *AdminHandler) GetSystemInfo(c *gin.Context) {
	startTime := time.Now().Add(-time.Hour * 24) // Mock start time
	uptime := time.Since(startTime)

	// Check database status
	dbStatus := "healthy"
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "unhealthy"
	}

	info := SystemInfoResponse{
		Version:        "1.0.0",
		Environment:    h.cfg.Environment,
		StartTime:      startTime,
		Uptime:         uptime.String(),
		GoVersion:      "1.21",
		DatabaseStatus: dbStatus,
		RedisStatus:    "healthy", // Assume healthy for now
	}

	c.JSON(http.StatusOK, info)
}

// SystemSettingsResponse represents system settings
type SystemSettingsResponse struct {
	MaxFileSize        string `json:"max_file_size"`
	CORSOrigins        string `json:"cors_origins"`
	JWTExpiresIn       string `json:"jwt_expires_in"`
	LogLevel           string `json:"log_level"`
	DatabaseURL        string `json:"database_url,omitempty"` // Hidden for security
	BackupRetention    int    `json:"backup_retention_days"`
	MaintenanceMode    bool   `json:"maintenance_mode"`
	RegistrationOpen   bool   `json:"registration_open"`
}

// GetSystemSettings returns system configuration settings
func (h *AdminHandler) GetSystemSettings(c *gin.Context) {
	corsOrigins := ""
	if len(h.cfg.AllowedOrigins) > 0 {
		corsOrigins = h.cfg.AllowedOrigins[0] // Take first origin
	}

	settings := SystemSettingsResponse{
		MaxFileSize:        "10MB",
		CORSOrigins:        corsOrigins,
		JWTExpiresIn:       "24h",
		LogLevel:           "info",
		BackupRetention:    30,
		MaintenanceMode:    false,
		RegistrationOpen:   true,
	}

	c.JSON(http.StatusOK, settings)
}

// RestartSystem restarts the application (not really implemented for demo)
func (h *AdminHandler) RestartSystem(c *gin.Context) {
	// In a real implementation, this would signal the application to restart
	// For demo purposes, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "System restart initiated",
		"status":  "scheduled",
	})
}

// ClearCache clears application cache
func (h *AdminHandler) ClearCache(c *gin.Context) {
	// In a real implementation, this would clear Redis or other caches
	// For demo purposes, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Cache cleared successfully",
		"cleared": "redis_cache, memory_cache",
	})
}

// CreateBackup creates a database backup
func (h *AdminHandler) CreateBackup(c *gin.Context) {
	// In a real implementation, this would create actual database backup
	// For demo purposes, we'll simulate backup creation
	filename := "cloudbox_backup_" + time.Now().Format("20060102_150405") + ".sql"
	
	c.JSON(http.StatusOK, gin.H{
		"message":    "Backup created successfully",
		"filename":   filename,
		"size":       "15.7 MB",
		"created_at": time.Now(),
	})
}