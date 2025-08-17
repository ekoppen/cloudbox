package handlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/security"
	"gorm.io/gorm"
)

type PluginHandler struct {
	db        *gorm.DB
	cfg       *config.Config
	validator *security.PluginValidator
}

func NewPluginHandler(db *gorm.DB, cfg *config.Config) *PluginHandler {
	return &PluginHandler{
		db:        db,
		cfg:       cfg,
		validator: security.NewPluginValidator(cfg),
	}
}

type PluginConfig struct {
	Name         string                 `json:"name"`
	Version      string                 `json:"version"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author"`
	Type         string                 `json:"type"`
	Main         string                 `json:"main"`
	Dependencies map[string]string      `json:"dependencies"`
	Permissions  []string               `json:"permissions"`
	UI           map[string]interface{} `json:"ui"`
	// Security fields
	Repository   string                 `json:"repository,omitempty"`
	Signature    string                 `json:"signature,omitempty"`
	Checksum     string                 `json:"checksum,omitempty"`
}

// Valid plugin name pattern (alphanumeric, dash, underscore only)
var validPluginNamePattern = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)

type Plugin struct {
	PluginConfig
	Status      string `json:"status"`
	InstalledAt string `json:"installed_at"`
	Path        string `json:"path"`
}

// GetActivePlugins returns all enabled plugins for the current project
func (h *PluginHandler) GetActivePlugins(c *gin.Context) {
	// Get project ID from context or request
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	var installations []models.PluginInstallation
	err = h.db.Where("project_id = ? AND status = 'enabled'", projectID).Find(&installations).Error
	if err != nil {
		log.Printf("Error fetching active plugins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch active plugins",
		})
		return
	}

	// Convert to response format
	plugins := make([]Plugin, 0, len(installations))
	for _, installation := range installations {
		plugin := h.convertInstallationToPlugin(installation)
		plugins = append(plugins, plugin)
	}

	// Add fallback plugin if no plugins installed
	if len(plugins) == 0 {
		plugins = h.getMockPlugins()
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": plugins,
	})
}

// GetAllPlugins returns all plugins (enabled and disabled) for admin users
func (h *PluginHandler) GetAllPlugins(c *gin.Context) {
	// Enhanced security: strict admin permission check
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Security audit logging
	h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, false, "")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := fmt.Sprintf("Insufficient privileges. Required: admin/superadmin, current: %s", userRole)
		h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Get project ID
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	var installations []models.PluginInstallation
	err = h.db.Where("project_id = ?", projectID).Find(&installations).Error
	if err != nil {
		log.Printf("Error fetching all plugins: %v", err)
		h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch plugins",
		})
		return
	}

	// Convert to response format
	plugins := make([]Plugin, 0, len(installations))
	for _, installation := range installations {
		plugin := h.convertInstallationToPlugin(installation)
		plugins = append(plugins, plugin)
	}

	// Add fallback plugin if no plugins installed
	if len(plugins) == 0 {
		plugins = h.getMockPlugins()
	}

	// Success audit log
	h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, true, "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": plugins,
	})
}

// EnablePlugin enables a plugin for a specific project
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
	// Enhanced security: strict admin permission check
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Security audit logging
	h.logPluginAction(c, "enable", "", "", "", userID, userEmail, false, "")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "enable", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	pluginName := c.Param("pluginName")
	
	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get project ID
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		errMsg := "Project ID is required"
		h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		errMsg := "Invalid project ID"
		h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "enable", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	currentStatus := installation.Status
	
	// Update installation status
	now := time.Now()
	installation.Status = "enabled"
	installation.LastEnabledAt = &now
	installation.UpdatedAt = now

	err = h.db.Save(&installation).Error
	if err != nil {
		h.logPluginAction(c, "enable", pluginName, currentStatus, "disabled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to enable plugin",
		})
		return
	}

	// Update plugin state
	h.updatePluginState(uint(projectID), pluginName, "enabled", userID)

	// Success audit log
	h.logPluginAction(c, "enable", pluginName, currentStatus, "enabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin enabled successfully",
	})
}

// DisablePlugin disables a plugin for a specific project
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
	// Enhanced security: strict admin permission check
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Security audit logging
	h.logPluginAction(c, "disable", "", "", "", userID, userEmail, false, "")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "disable", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	pluginName := c.Param("pluginName")
	
	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get project ID
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		errMsg := "Project ID is required"
		h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		errMsg := "Invalid project ID"
		h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "disable", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	currentStatus := installation.Status
	
	// Update installation status
	now := time.Now()
	installation.Status = "disabled"
	installation.LastDisabledAt = &now
	installation.UpdatedAt = now

	err = h.db.Save(&installation).Error
	if err != nil {
		h.logPluginAction(c, "disable", pluginName, currentStatus, "enabled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to disable plugin",
		})
		return
	}

	// Update plugin state
	h.updatePluginState(uint(projectID), pluginName, "disabled", userID)

	// Success audit log
	h.logPluginAction(c, "disable", pluginName, currentStatus, "disabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin disabled successfully",
	})
}

// InstallPlugin securely installs a plugin from an approved GitHub repository
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin installation"
		h.logPluginAction(c, "install", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	var req struct {
		Repository string `json:"repository" binding:"required"`
		Version    string `json:"version,omitempty"`
		ProjectID  uint   `json:"project_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logPluginAction(c, "install", "", "", "", userID, userEmail, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}
	
	// Validate GitHub repository
	repo, err := h.validator.ValidateGitHubRepository(req.Repository)
	if err != nil {
		h.logPluginAction(c, "install", "", "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Repository validation failed: " + err.Error(),
		})
		return
	}
	
	pluginName := repo.Name
	
	// Check if already installed
	var existingInstallation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, req.ProjectID).First(&existingInstallation).Error
	if err == nil {
		h.logPluginAction(c, "install", pluginName, existingInstallation.Status, existingInstallation.Status, userID, userEmail, false, "Plugin already installed")
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Plugin already installed",
		})
		return
	}

	// Parse user ID
	userIDInt, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		userIDInt = 0
	}

	// Create plugin installation record
	installation := models.PluginInstallation{
		PluginName:       pluginName,
		PluginVersion:    req.Version,
		ProjectID:        req.ProjectID,
		Status:           "disabled", // Disabled by default for security
		InstallationPath: fmt.Sprintf("./plugins/%s", pluginName),
		InstalledBy:      uint(userIDInt),
		InstalledAt:      time.Now(),
		Config:          make(map[string]interface{}),
		Environment:     make(map[string]interface{}),
	}

	err = h.db.Create(&installation).Error
	if err != nil {
		h.logPluginAction(c, "install", pluginName, "uninstalled", "uninstalled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create installation record",
		})
		return
	}

	// Create plugin state record
	state := models.PluginState{
		PluginName:     pluginName,
		ProjectID:      req.ProjectID,
		CurrentStatus:  "disabled",
		StateChangedAt: time.Now(),
		StateChangedBy: &installation.InstalledBy,
		HealthStatus:   "unknown",
		HealthDetails:  make(map[string]interface{}),
	}

	err = h.db.Create(&state).Error
	if err != nil {
		log.Printf("Warning: Failed to create plugin state record: %v", err)
	}

	// Record download attempt
	h.recordPluginDownload(pluginName, req.Version, req.ProjectID, uint(userIDInt), req.Repository, "completed", c)

	// Success audit log
	h.logPluginAction(c, "install", pluginName, "uninstalled", "disabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin installed successfully (disabled by default)",
		"plugin":  pluginName,
		"status":  "disabled",
	})
}

// UninstallPlugin removes a plugin installation
func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
	// Enhanced security: strict admin permission check
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "uninstall", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	pluginName := c.Param("pluginName")
	
	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get project ID
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		errMsg := "Project ID is required"
		h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		errMsg := "Invalid project ID"
		h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "uninstall", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	currentStatus := installation.Status

	// Delete installation record
	err = h.db.Delete(&installation).Error
	if err != nil {
		h.logPluginAction(c, "uninstall", pluginName, currentStatus, currentStatus, userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to uninstall plugin",
		})
		return
	}

	// Delete plugin state record
	h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).Delete(&models.PluginState{})

	// Clean up plugin files (if they exist)
	pluginPath := filepath.Join("./plugins", pluginName)
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		os.RemoveAll(pluginPath)
	}

	// Success audit log
	h.logPluginAction(c, "uninstall", pluginName, currentStatus, "uninstalled", userID, userEmail, true, "")
	
	log.Printf("Plugin %s uninstalled by %s", pluginName, userEmail)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin uninstalled successfully",
	})
}

// ReloadPlugins reloads all plugins (for development)
func (h *PluginHandler) ReloadPlugins(c *gin.Context) {
	// Check admin permissions - allow both admin and superadmin
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	// In a real implementation, this would reload plugin modules
	// For now, we just return success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugins reloaded successfully",
	})
}

// GetApprovedRepositories returns the list of approved plugin repositories
func (h *PluginHandler) GetApprovedRepositories(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}
	
	var repositories []models.ApprovedRepository
	err := h.db.Where("is_active = ?", true).Find(&repositories).Error
	if err != nil {
		log.Printf("Error fetching approved repositories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch approved repositories",
		})
		return
	}

	// Convert to simple list of URLs
	repos := make([]string, 0, len(repositories))
	for _, repo := range repositories {
		repos = append(repos, repo.RepositoryURL)
	}
	
	// Add static repositories from security package
	for repo := range security.ApprovedRepositories {
		fullURL := "https://" + repo
		found := false
		for _, existing := range repos {
			if existing == fullURL {
				found = true
				break
			}
		}
		if !found {
			repos = append(repos, fullURL)
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"repositories": repos,
	})
}

// GetAuditLogs retrieves plugin audit logs for security monitoring
func (h *PluginHandler) GetAuditLogs(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}
	
	// Query parameters
	limit := c.DefaultQuery("limit", "100")
	offset := c.DefaultQuery("offset", "0")
	pluginName := c.Query("plugin")
	action := c.Query("action")
	userEmail := c.Query("user")
	
	var logs []models.PluginAuditLog
	query := h.db.Model(&models.PluginAuditLog{})
	
	// Apply filters
	if pluginName != "" {
		query = query.Where("plugin_name = ?", pluginName)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if userEmail != "" {
		query = query.Where("user_email ILIKE ?", "%"+userEmail+"%")
	}
	
	// Order by most recent first
	query = query.Order("created_at DESC")
	
	// Apply pagination
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)
	if limitInt > 100 {
		limitInt = 100
	}
	query = query.Limit(limitInt).Offset(offsetInt)
	
	if err := query.Find(&logs).Error; err != nil {
		log.Printf("Failed to fetch audit logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch audit logs",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"logs":    logs,
		"count":   len(logs),
	})
}

// HotReloadPlugin hot reloads a specific plugin (for development)
func (h *PluginHandler) HotReloadPlugin(c *gin.Context) {
	pluginName := c.Param("pluginName")
	if pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	// In a real implementation, this would:
	// 1. Reload plugin configuration
	// 2. Re-register plugin routes
	// 3. Update UI components
	// 4. Notify frontend of changes
	
	fmt.Printf("Plugin %s hot reloaded\n", pluginName)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin hot reloaded successfully",
	})
}

// DebugAuth provides debug information about authentication state
func (h *PluginHandler) DebugAuth(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"debug": gin.H{
			"user_role":  userRole,
			"user_id":    userID,
			"user_email": userEmail,
			"has_auth":   userRole != "",
		},
	})
}

// Helper methods

// convertInstallationToPlugin converts a database installation record to API response format
func (h *PluginHandler) convertInstallationToPlugin(installation models.PluginInstallation) Plugin {
	return Plugin{
		PluginConfig: PluginConfig{
			Name:        installation.PluginName,
			Version:     installation.PluginVersion,
			Description: "Installed plugin", // TODO: Get from registry
			Author:      "Unknown",          // TODO: Get from registry
			Type:        "dashboard-plugin", // TODO: Get from registry
			Main:        "index.js",         // TODO: Get from registry
			UI:          installation.Config,
		},
		Status:      installation.Status,
		InstalledAt: installation.InstalledAt.Format("2006-01-02T15:04:05Z"),
		Path:        installation.InstallationPath,
	}
}

// getMockPlugins returns mock plugin data for fallback
func (h *PluginHandler) getMockPlugins() []Plugin {
	return []Plugin{
		{
			PluginConfig: PluginConfig{
				Name:        "cloudbox-script-runner",
				Version:     "1.0.0",
				Description: "Universal Script Runner for CloudBox - Database scripts en project setup",
				Author:      "CloudBox Development Team",
				Type:        "dashboard-plugin",
				Main:        "index.js",
				Dependencies: map[string]string{
					"cloudbox-sdk": "^1.0.0",
				},
				Permissions: []string{
					"database:read",
					"database:write",
					"functions:deploy",
					"webhooks:create",
					"projects:manage",
				},
				UI: map[string]interface{}{
					"project_menu": map[string]interface{}{
						"title": "Scripts",
						"icon":  "terminal",
						"path":  "/dashboard/projects/{projectId}/scripts",
					},
				},
			},
			Status:      "enabled",
			InstalledAt: "2024-08-16T12:00:00Z",
			Path:        "./plugins/script-runner",
		},
	}
}

// updatePluginState updates or creates plugin state record
func (h *PluginHandler) updatePluginState(projectID uint, pluginName, status, userID string) {
	userIDInt, _ := strconv.ParseUint(userID, 10, 32)
	userIDPtr := uint(userIDInt)

	var state models.PluginState
	err := h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&state).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new state
		state = models.PluginState{
			PluginName:     pluginName,
			ProjectID:      projectID,
			CurrentStatus:  status,
			StateChangedAt: time.Now(),
			StateChangedBy: &userIDPtr,
			HealthStatus:   "unknown",
			HealthDetails:  make(map[string]interface{}),
		}
		h.db.Create(&state)
	} else if err == nil {
		// Update existing state
		state.CurrentStatus = status
		state.StateChangedAt = time.Now()
		state.StateChangedBy = &userIDPtr
		h.db.Save(&state)
	}
}

// recordPluginDownload records a plugin download attempt
func (h *PluginHandler) recordPluginDownload(pluginName, version string, projectID, userID uint, source, status string, c *gin.Context) {
	download := models.PluginDownload{
		PluginName:     pluginName,
		PluginVersion:  version,
		ProjectID:      projectID,
		UserID:         userID,
		DownloadSource: source,
		DownloadStatus: status,
		ClientIP:       c.ClientIP(),
		UserAgent:      c.GetHeader("User-Agent"),
		StartedAt:      time.Now(),
	}

	if status == "completed" {
		now := time.Now()
		download.CompletedAt = &now
	} else if status == "failed" {
		now := time.Now()
		download.FailedAt = &now
	}

	h.db.Create(&download)
}

// validatePluginName ensures plugin names are safe and follow expected patterns
func (h *PluginHandler) validatePluginName(pluginName string) error {
	if pluginName == "" {
		return fmt.Errorf("plugin name is required")
	}
	
	if len(pluginName) > 100 {
		return fmt.Errorf("plugin name too long (max 100 characters)")
	}
	
	if !validPluginNamePattern.MatchString(pluginName) {
		return fmt.Errorf("invalid plugin name: only alphanumeric characters, dashes, and underscores allowed")
	}
	
	// Prevent common dangerous patterns
	dangerousPatterns := []string{"../", "./", "/", "\\", ":", "|", "<", ">", "?", "*"}
	for _, pattern := range dangerousPatterns {
		if strings.Contains(pluginName, pattern) {
			return fmt.Errorf("invalid plugin name: contains prohibited characters")
		}
	}
	
	return nil
}

// GetMarketplace returns available plugins from approved repositories
func (h *PluginHandler) GetMarketplace(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	// Get approved repositories
	var repositories []models.ApprovedRepository
	err := h.db.Where("is_active = ?", true).Find(&repositories).Error
	if err != nil {
		log.Printf("Error fetching approved repositories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch marketplace data",
		})
		return
	}

	// Mock marketplace data - in a real implementation, this would:
	// 1. Fetch plugin manifests from each repository
	// 2. Parse and validate plugin metadata
	// 3. Check for updates and compatibility
	// 4. Provide security status and ratings
	
	marketplace := []map[string]interface{}{
		{
			"name":        "cloudbox-script-runner",
			"version":     "1.0.0",
			"description": "Universal Script Runner for CloudBox - Database scripts en project setup",
			"author":      "CloudBox Development Team",
			"repository":  "https://github.com/cloudbox/plugins",
			"category":    "development",
			"tags":        []string{"database", "scripts", "automation"},
			"rating":      4.8,
			"downloads":   1250,
			"verified":    true,
			"official":    true,
			"last_updated": "2024-08-16T12:00:00Z",
			"permissions": []string{
				"database:read",
				"database:write",
				"functions:deploy",
				"webhooks:create",
				"projects:manage",
			},
		},
		{
			"name":        "cloudbox-backup-manager",
			"version":     "2.1.0",
			"description": "Automated backup management and scheduling for CloudBox projects",
			"author":      "CloudBox Development Team",
			"repository":  "https://github.com/cloudbox/official-plugins",
			"category":    "backup",
			"tags":        []string{"backup", "automation", "scheduling"},
			"rating":      4.6,
			"downloads":   890,
			"verified":    true,
			"official":    true,
			"last_updated": "2024-08-10T09:30:00Z",
			"permissions": []string{
				"storage:read",
				"storage:write",
				"projects:read",
			},
		},
		{
			"name":        "cloudbox-analytics",
			"version":     "1.3.2",
			"description": "Advanced analytics and reporting for CloudBox applications",
			"author":      "CloudBox Community",
			"repository":  "https://github.com/cloudbox/community-plugins",
			"category":    "analytics",
			"tags":        []string{"analytics", "reporting", "dashboard"},
			"rating":      4.2,
			"downloads":   654,
			"verified":    true,
			"official":    false,
			"last_updated": "2024-08-05T14:20:00Z",
			"permissions": []string{
				"database:read",
				"functions:execute",
				"projects:read",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": marketplace,
		"repositories": len(repositories),
	})
}

// SearchMarketplace searches for plugins in the marketplace
func (h *PluginHandler) SearchMarketplace(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	query := c.Query("q")
	category := c.Query("category")
	official := c.Query("official")
	
	// Mock search implementation - in reality this would search through
	// indexed plugin metadata from all approved repositories
	
	var results []map[string]interface{}
	
	// Simple mock filtering based on query parameters
	allPlugins := []map[string]interface{}{
		{
			"name":        "cloudbox-script-runner",
			"version":     "1.0.0",
			"description": "Universal Script Runner for CloudBox - Database scripts en project setup",
			"author":      "CloudBox Development Team",
			"repository":  "https://github.com/cloudbox/plugins",
			"category":    "development",
			"official":    true,
			"verified":    true,
		},
		{
			"name":        "cloudbox-backup-manager",
			"version":     "2.1.0",
			"description": "Automated backup management and scheduling for CloudBox projects",
			"author":      "CloudBox Development Team",
			"repository":  "https://github.com/cloudbox/official-plugins",
			"category":    "backup",
			"official":    true,
			"verified":    true,
		},
		{
			"name":        "cloudbox-analytics",
			"version":     "1.3.2",
			"description": "Advanced analytics and reporting for CloudBox applications",
			"author":      "CloudBox Community",
			"repository":  "https://github.com/cloudbox/community-plugins",
			"category":    "analytics",
			"official":    false,
			"verified":    true,
		},
	}

	for _, plugin := range allPlugins {
		include := true
		
		// Filter by search query
		if query != "" {
			queryLower := strings.ToLower(query)
			nameMatch := strings.Contains(strings.ToLower(plugin["name"].(string)), queryLower)
			descMatch := strings.Contains(strings.ToLower(plugin["description"].(string)), queryLower)
			if !nameMatch && !descMatch {
				include = false
			}
		}
		
		// Filter by category
		if category != "" && plugin["category"].(string) != category {
			include = false
		}
		
		// Filter by official status
		if official == "true" && !plugin["official"].(bool) {
			include = false
		} else if official == "false" && plugin["official"].(bool) {
			include = false
		}
		
		if include {
			results = append(results, plugin)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": results,
		"count":   len(results),
		"query":   query,
	})
}

// GetPluginDetails returns detailed information about a specific plugin
func (h *PluginHandler) GetPluginDetails(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	pluginName := c.Param("pluginName")
	repository := c.Query("repository")

	if pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Plugin name is required",
		})
		return
	}

	// In a real implementation, this would:
	// 1. Fetch plugin manifest from the specified repository
	// 2. Validate the plugin metadata
	// 3. Check security status and compatibility
	// 4. Get installation statistics and reviews
	
	// Mock detailed plugin information
	pluginDetails := map[string]interface{}{
		"name":         pluginName,
		"version":      "1.0.0",
		"description":  "Universal Script Runner for CloudBox - Database scripts en project setup",
		"author":       "CloudBox Development Team",
		"repository":   repository,
		"license":      "MIT",
		"homepage":     "https://github.com/cloudbox/plugins",
		"documentation": "https://docs.cloudbox.com/plugins/script-runner",
		"category":     "development",
		"tags":         []string{"database", "scripts", "automation"},
		"rating":       4.8,
		"reviews":      []map[string]interface{}{
			{
				"author": "developer1",
				"rating": 5,
				"comment": "Excellent plugin for database management",
				"date": "2024-08-15",
			},
		},
		"downloads":      1250,
		"install_count":  890,
		"verified":       true,
		"official":       true,
		"security_scan": map[string]interface{}{
			"last_scan": "2024-08-16T10:00:00Z",
			"status":    "clean",
			"issues":    0,
		},
		"compatibility": map[string]interface{}{
			"cloudbox_version": ">=1.0.0",
			"node_version":     ">=14.0.0",
			"supported_os":     []string{"linux", "darwin", "windows"},
		},
		"last_updated": "2024-08-16T12:00:00Z",
		"changelog": []map[string]interface{}{
			{
				"version": "1.0.0",
				"date": "2024-08-16",
				"changes": []string{
					"Initial release",
					"Added database script execution",
					"Added project setup automation",
				},
			},
		},
		"permissions": []string{
			"database:read",
			"database:write",
			"functions:deploy",
			"webhooks:create",
			"projects:manage",
		},
		"dependencies": map[string]string{
			"cloudbox-sdk": "^1.0.0",
		},
		"ui_screenshots": []string{
			"https://example.com/screenshots/plugin1.png",
			"https://example.com/screenshots/plugin2.png",
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugin":  pluginDetails,
	})
}

// InstallFromMarketplace installs a plugin directly from marketplace
func (h *PluginHandler) InstallFromMarketplace(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin installation"
		h.logPluginAction(c, "install_marketplace", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	var req struct {
		PluginName string `json:"plugin_name" binding:"required"`
		Repository string `json:"repository" binding:"required"`
		Version    string `json:"version,omitempty"`
		ProjectID  uint   `json:"project_id" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logPluginAction(c, "install_marketplace", "", "", "", userID, userEmail, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Use the existing InstallPlugin logic with marketplace data
	// This would typically fetch the plugin from the marketplace catalog
	// and then proceed with the standard installation flow
	
	// For now, redirect to standard installation
	installReq := struct {
		Repository string `json:"repository"`
		Version    string `json:"version,omitempty"`
		ProjectID  uint   `json:"project_id"`
	}{
		Repository: req.Repository,
		Version:    req.Version,
		ProjectID:  req.ProjectID,
	}

	// Mock the request body for InstallPlugin
	reqBody, _ := json.Marshal(installReq)
	c.Request.Body = ioutil.NopCloser(strings.NewReader(string(reqBody)))
	c.Request.Header.Set("Content-Type", "application/json")

	// Call existing InstallPlugin method
	h.InstallPlugin(c)
}

// GetPluginHealth returns health status for installed plugins
func (h *PluginHandler) GetPluginHealth(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	// Get project ID
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	if projectIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	projectID, err := strconv.ParseUint(projectIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Get plugin states for the project
	var states []models.PluginState
	err = h.db.Where("project_id = ?", projectID).Find(&states).Error
	if err != nil {
		log.Printf("Error fetching plugin states: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch plugin health data",
		})
		return
	}

	// Convert to health report format
	healthReport := make([]map[string]interface{}, 0, len(states))
	for _, state := range states {
		health := map[string]interface{}{
			"plugin_name":       state.PluginName,
			"status":           state.CurrentStatus,
			"health_status":    state.HealthStatus,
			"last_health_check": state.LastHealthCheck,
			"uptime_seconds":   state.UptimeSeconds,
			"cpu_usage":        state.CPUUsage,
			"memory_usage":     state.MemoryUsage,
			"port":             state.Port,
			"process_id":       state.ProcessID,
			"state_changed_at": state.StateChangedAt,
			"health_details":   state.HealthDetails,
		}
		healthReport = append(healthReport, health)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"health":  healthReport,
		"count":   len(healthReport),
	})
}

// UpdatePluginConfig updates plugin configuration
func (h *PluginHandler) UpdatePluginConfig(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin configuration"
		h.logPluginAction(c, "configure", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	pluginName := c.Param("pluginName")
	
	var req struct {
		ProjectID   uint                   `json:"project_id" binding:"required"`
		Config      map[string]interface{} `json:"config"`
		Environment map[string]interface{} `json:"environment"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logPluginAction(c, "configure", pluginName, "", "", userID, userEmail, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "configure", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err := h.db.Where("plugin_name = ? AND project_id = ?", pluginName, req.ProjectID).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "configure", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "configure", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Update configuration
	if req.Config != nil {
		installation.Config = req.Config
	}
	if req.Environment != nil {
		installation.Environment = req.Environment
	}
	installation.UpdatedAt = time.Now()

	err = h.db.Save(&installation).Error
	if err != nil {
		h.logPluginAction(c, "configure", pluginName, installation.Status, installation.Status, userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update plugin configuration",
		})
		return
	}

	// Success audit log
	h.logPluginAction(c, "configure", pluginName, installation.Status, installation.Status, userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin configuration updated successfully",
	})
}

// logPluginAction creates audit trail for all plugin operations
func (h *PluginHandler) logPluginAction(c *gin.Context, action, pluginName, oldStatus, newStatus, userID, userEmail string, success bool, errorMsg string) {
	// Extract client information
	clientIP := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	
	// Parse user ID
	userIDInt, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		userIDInt = 0
	}
	
	// Create audit log entry
	auditLog := models.PluginAuditLog{
		UserID:     uint(userIDInt),
		UserEmail:  userEmail,
		Action:     action,
		PluginName: pluginName,
		OldStatus:  oldStatus,
		NewStatus:  newStatus,
		IPAddress:  clientIP,
		UserAgent:  userAgent,
		Success:    success,
		ErrorMsg:   errorMsg,
		CreatedAt:  time.Now(),
	}
	
	// Log to database (if available)
	if h.db != nil {
		if err := h.db.Create(&auditLog).Error; err != nil {
			log.Printf("Failed to log plugin audit trail: %v", err)
		}
	}
	
	// Always log to system log for security monitoring
	log.Printf("PLUGIN_AUDIT: user=%s action=%s plugin=%s success=%v ip=%s", 
		userEmail, action, pluginName, success, clientIP)
}