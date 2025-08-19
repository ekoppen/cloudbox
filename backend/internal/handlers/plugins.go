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

	// No fallback plugins - return empty array for clean installations

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
	
	// Note: Success audit logging happens at the end of the function
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := fmt.Sprintf("Insufficient privileges. Required: admin/superadmin, current: %s", userRole)
		h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	// Get project ID (optional for admin - if not provided, show all plugins)
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	var installations []models.PluginInstallation
	var err error

	if projectIDStr == "" {
		// Admin view: show only system-wide plugin controls (project_id = 0)
		err = h.db.Where("project_id = ?", 0).Find(&installations).Error
	} else {
		// Project-specific view
		projectID, parseErr := strconv.ParseUint(projectIDStr, 10, 32)
		if parseErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "Invalid project ID",
			})
			return
		}
		err = h.db.Where("project_id = ?", projectID).Find(&installations).Error
	}
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

	// No fallback plugins - return empty array for clean installations

	// Success audit log
	h.logPluginAction(c, "list_all", "", "", "", userID, userEmail, true, "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": plugins,
	})
}

// EnablePlugin enables a plugin system-wide or for a specific project
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

	// Get project ID (optional for admin operations)
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	// If no project ID provided, enable for all projects (admin operation)
	if projectIDStr == "" {
		// Enable plugin system-wide
		var installations []models.PluginInstallation
		err := h.db.Where("plugin_name = ?", pluginName).Find(&installations).Error
		if err != nil {
			h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
			return
		}

		if len(installations) == 0 {
			errMsg := "Plugin not installed in any project"
			h.logPluginAction(c, "enable", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}

		// Update all installations
		now := time.Now()
		for _, installation := range installations {
			installation.Status = "enabled"
			installation.LastEnabledAt = &now
			installation.UpdatedAt = now
			
			if err := h.db.Save(&installation).Error; err != nil {
				h.logPluginAction(c, "enable", pluginName, "", "", userID, userEmail, false, err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to enable plugin",
				})
				return
			}

			// Update plugin state for each project
			h.updatePluginState(installation.ProjectID, pluginName, "enabled", userID)
		}

		// Success audit log
		h.logPluginAction(c, "enable", pluginName, "disabled", "enabled", userID, userEmail, true, fmt.Sprintf("Enabled for %d projects", len(installations)))
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Plugin enabled successfully in %d projects", len(installations)),
		})
		return
	}

	// Project-specific enable
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

// DisablePlugin disables a plugin system-wide or for a specific project
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

	// Get project ID (optional for admin operations)
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	// If no project ID provided, disable for all projects (admin operation)
	if projectIDStr == "" {
		// Disable plugin system-wide
		var installations []models.PluginInstallation
		err := h.db.Where("plugin_name = ?", pluginName).Find(&installations).Error
		if err != nil {
			h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
			return
		}

		if len(installations) == 0 {
			errMsg := "Plugin not installed in any project"
			h.logPluginAction(c, "disable", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}

		// Update all installations
		now := time.Now()
		for _, installation := range installations {
			installation.Status = "disabled"
			installation.LastDisabledAt = &now
			installation.UpdatedAt = now
			
			if err := h.db.Save(&installation).Error; err != nil {
				h.logPluginAction(c, "disable", pluginName, "", "", userID, userEmail, false, err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to disable plugin",
				})
				return
			}

			// Update plugin state for each project
			h.updatePluginState(installation.ProjectID, pluginName, "disabled", userID)
		}

		// Success audit log
		h.logPluginAction(c, "disable", pluginName, "enabled", "disabled", userID, userEmail, true, fmt.Sprintf("Disabled for %d projects", len(installations)))
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Plugin disabled successfully in %d projects", len(installations)),
		})
		return
	}

	// Project-specific disable
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

// UninstallPlugin removes a plugin installation system-wide or for a specific project
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

	// Get project ID (optional for admin operations)
	projectIDStr := c.GetString("project_id")
	if projectIDStr == "" {
		projectIDStr = c.Query("project_id")
	}

	// If no project ID provided, uninstall from all projects (admin operation)
	if projectIDStr == "" {
		// Uninstall plugin system-wide
		var installations []models.PluginInstallation
		err := h.db.Where("plugin_name = ?", pluginName).Find(&installations).Error
		if err != nil {
			h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, err.Error())
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Database error",
			})
			return
		}

		if len(installations) == 0 {
			errMsg := "Plugin not installed in any project"
			h.logPluginAction(c, "uninstall", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}

		// Delete all installations
		for _, installation := range installations {
			// Delete installation record
			if err := h.db.Delete(&installation).Error; err != nil {
				h.logPluginAction(c, "uninstall", pluginName, "", "", userID, userEmail, false, err.Error())
				c.JSON(http.StatusInternalServerError, gin.H{
					"success": false,
					"error":   "Failed to uninstall plugin",
				})
				return
			}

			// Delete plugin state record
			h.db.Where("plugin_name = ? AND project_id = ?", pluginName, installation.ProjectID).Delete(&models.PluginState{})
		}

		// Clean up plugin files (if they exist)
		pluginPath := filepath.Join("./plugins", pluginName)
		if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
			os.RemoveAll(pluginPath)
		}

		// Success audit log
		h.logPluginAction(c, "uninstall", pluginName, "installed", "uninstalled", userID, userEmail, true, fmt.Sprintf("Uninstalled from %d projects", len(installations)))
		
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"message": fmt.Sprintf("Plugin uninstalled successfully from %d projects", len(installations)),
		})
		return
	}

	// Project-specific uninstall
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

	// Check if plugin is still used by other projects
	var count int64
	h.db.Model(&models.PluginInstallation{}).Where("plugin_name = ?", pluginName).Count(&count)
	
	// Only delete plugin files if not used by any other project
	if count == 0 {
		pluginPath := filepath.Join("./plugins", pluginName)
		if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
			os.RemoveAll(pluginPath)
		}
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
	err := h.db.Unscoped().Where("is_active = ?", true).Find(&repositories).Error
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
	// Get plugin info from registry
	var registryPlugin models.PluginRegistry
	h.db.Where("name = ?", installation.PluginName).First(&registryPlugin)
	
	// Use registry info if available, otherwise fallback to basic info
	description := "Installed plugin"
	author := "Unknown"
	pluginType := "dashboard-plugin"
	uiConfig := installation.Config // Project-specific config
	
	if registryPlugin.ID != 0 {
		description = registryPlugin.Description
		author = registryPlugin.Author
		pluginType = registryPlugin.Type
		// Use UI config from registry for navigation, not project config
		if registryPlugin.UIConfig != nil {
			uiConfig = registryPlugin.UIConfig
		}
	}

	return Plugin{
		PluginConfig: PluginConfig{
			Name:        installation.PluginName,
			Version:     installation.PluginVersion,
			Description: description,
			Author:      author,
			Type:        pluginType,
			Main:        "index.js",
			UI:          uiConfig,
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
	err := h.db.Unscoped().Where("is_active = ?", true).Find(&repositories).Error
	if err != nil {
		log.Printf("Error fetching approved repositories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch marketplace data",
		})
		return
	}

	// Fetch marketplace plugins from database
	var marketplacePlugins []models.PluginMarketplace
	err = h.db.Unscoped().Where("status = ?", "available").Order("download_count DESC, install_count DESC").Find(&marketplacePlugins).Error
	if err != nil {
		log.Printf("Error fetching marketplace plugins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch marketplace plugins",
		})
		return
	}

	// Convert to response format
	marketplace := make([]map[string]interface{}, 0, len(marketplacePlugins))
	for _, plugin := range marketplacePlugins {
		marketplaceEntry := map[string]interface{}{
			"name":         plugin.Name,
			"version":      plugin.Version,
			"description":  plugin.Description,
			"author":       plugin.Author,
			"repository":   plugin.Repository,
			"type":         plugin.Type,
			"license":      plugin.License,
			"downloads":    plugin.DownloadCount,
			"installs":     plugin.InstallCount,
			"verified":     plugin.IsVerified,
			"approved":     plugin.IsApproved,
			"status":       plugin.Status,
			"permissions":  plugin.Permissions,
			"dependencies": plugin.Dependencies,
			"ui_config":    plugin.UIConfig,
			"main_file":    plugin.MainFile,
			"checksum":     plugin.Checksum,
			"signature":    plugin.Signature,
			"registry_source": plugin.RegistrySource,
			"source_metadata": plugin.SourceMetadata,
			"published_at":    plugin.PublishedAt,
			"deprecated_at":   plugin.DeprecatedAt,
		}
		marketplace = append(marketplace, marketplaceEntry)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": marketplace,
		"repositories": len(repositories),
	})
}

// AddPluginToMarketplace adds a new plugin to the marketplace (superadmin only)
func (h *PluginHandler) AddPluginToMarketplace(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Only superadmins can add plugins to marketplace
	if userRole != "superadmin" {
		errMsg := "Superadmin access required to add plugins to marketplace"
		h.logPluginAction(c, "add_to_marketplace", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	var req struct {
		Name         string                 `json:"name" binding:"required"`
		Version      string                 `json:"version" binding:"required"`
		Description  string                 `json:"description" binding:"required"`
		Author       string                 `json:"author" binding:"required"`
		Repository   string                 `json:"repository" binding:"required"`
		License      string                 `json:"license"`
		Tags         []string               `json:"tags"`
		Permissions  []string               `json:"permissions"`
		Dependencies map[string]string      `json:"dependencies"`
		IsVerified   bool                   `json:"verified"`
		IsApproved   bool                   `json:"approved"`
		Type         string                 `json:"type"`
		MainFile     string                 `json:"main_file"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logPluginAction(c, "add_to_marketplace", "", "", "", userID, userEmail, false, "Invalid request data: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(req.Name); err != nil {
		h.logPluginAction(c, "add_to_marketplace", req.Name, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Validate GitHub repository format
	repo, err := h.validator.ValidateGitHubRepository(req.Repository)
	if err != nil {
		h.logPluginAction(c, "add_to_marketplace", req.Name, "", "", userID, userEmail, false, "Repository validation failed: "+err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Repository validation failed: " + err.Error(),
		})
		return
	}

	// Check if plugin already exists in marketplace
	var existingPlugin models.PluginMarketplace
	err = h.db.Unscoped().Table("plugin_registry").Where("name = ? AND repository = ?", req.Name, req.Repository).First(&existingPlugin).Error
	if err == nil {
		h.logPluginAction(c, "add_to_marketplace", req.Name, "existing", "existing", userID, userEmail, false, "Plugin already exists in marketplace")
		c.JSON(http.StatusConflict, gin.H{
			"success": false,
			"error":   "Plugin already exists in marketplace",
		})
		return
	} else if err != gorm.ErrRecordNotFound {
		h.logPluginAction(c, "add_to_marketplace", req.Name, "", "", userID, userEmail, false, "Database error: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Set defaults
	if req.License == "" {
		req.License = "MIT"
	}
	if req.Type == "" {
		req.Type = "dashboard-plugin"
	}
	if req.MainFile == "" {
		req.MainFile = "index.js"
	}
	if req.Tags == nil {
		req.Tags = []string{}
	}
	if req.Permissions == nil {
		req.Permissions = []string{}
	}
	if req.Dependencies == nil {
		req.Dependencies = make(map[string]string)
	}

	// Create marketplace entry
	now := time.Now()
	marketplacePlugin := models.PluginMarketplace{
		Name:         req.Name,
		Version:      req.Version,
		Description:  req.Description,
		Author:       req.Author,
		Repository:   req.Repository,
		License:      req.License,
		Type:         req.Type,
		Permissions:  req.Permissions,
		Dependencies: convertToInterface(req.Dependencies),
		IsVerified:   req.IsVerified,
		IsApproved:   req.IsApproved,
		Status:       "available",
		MainFile:     req.MainFile,
		RegistrySource: "manual",
		SourceMetadata: map[string]interface{}{
			"added_by":    userEmail,
			"added_by_id": userID,
			"repo_owner":  repo.Owner,
			"repo_name":   repo.Name,
		},
		PublishedAt:  &now,
	}

	err = h.db.Create(&marketplacePlugin).Error
	if err != nil {
		h.logPluginAction(c, "add_to_marketplace", req.Name, "", "", userID, userEmail, false, "Failed to add plugin to marketplace: "+err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to add plugin to marketplace",
		})
		return
	}

	// Success audit log
	h.logPluginAction(c, "add_to_marketplace", req.Name, "", "available", userID, userEmail, true, "Plugin added to marketplace successfully")
	
	log.Printf("Plugin %s added to marketplace by %s", req.Name, userEmail)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin added to marketplace successfully",
		"plugin":  marketplacePlugin.Name,
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
	verified := c.Query("verified")
	featured := c.Query("featured")
	
	// Build database query
	dbQuery := h.db.Unscoped().Table("plugin_registry").Where("status = ?", "available")
	
	// Apply search query filter - search in name, description, and author
	if query != "" {
		queryLower := strings.ToLower(query)
		searchPattern := "%" + queryLower + "%"
		
		// Search in available columns: name, description, author, and repository
		dbQuery = dbQuery.Where(
			"LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR LOWER(author) LIKE ? OR LOWER(repository) LIKE ?", 
			searchPattern, searchPattern, searchPattern, searchPattern,
		)
	}
	
	// Note: category and official filters removed as these columns don't exist in plugin_registry table
	
	// Apply verified status filter
	if verified == "true" {
		dbQuery = dbQuery.Where("is_verified = ?", true)
	} else if verified == "false" {
		dbQuery = dbQuery.Where("is_verified = ?", false)
	}
	
	// Apply approved status filter (replacing featured)
	if featured == "true" {
		dbQuery = dbQuery.Where("is_approved = ?", true)
	} else if featured == "false" {
		dbQuery = dbQuery.Where("is_approved = ?", false)
	}
	
	// Fetch filtered marketplace plugins from database
	var marketplacePlugins []models.PluginMarketplace
	err := dbQuery.Order("download_count DESC, install_count DESC").Find(&marketplacePlugins).Error
	if err != nil {
		log.Printf("Error searching marketplace plugins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to search marketplace plugins",
		})
		return
	}

	// Convert to response format
	results := make([]map[string]interface{}, 0, len(marketplacePlugins))
	for _, plugin := range marketplacePlugins {
		pluginEntry := map[string]interface{}{
			"name":         plugin.Name,
			"version":      plugin.Version,
			"description":  plugin.Description,
			"author":       plugin.Author,
			"repository":   plugin.Repository,
			"type":         plugin.Type,
			"license":      plugin.License,
			"downloads":    plugin.DownloadCount,
			"installs":     plugin.InstallCount,
			"verified":     plugin.IsVerified,
			"approved":     plugin.IsApproved,
			"status":       plugin.Status,
			"permissions":  plugin.Permissions,
			"dependencies": plugin.Dependencies,
			"ui_config":    plugin.UIConfig,
			"main_file":    plugin.MainFile,
			"checksum":     plugin.Checksum,
			"signature":    plugin.Signature,
			"registry_source": plugin.RegistrySource,
			"source_metadata": plugin.SourceMetadata,
			"published_at":    plugin.PublishedAt,
			"deprecated_at":   plugin.DeprecatedAt,
		}
		results = append(results, pluginEntry)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": results,
		"count":   len(results),
		"filters": map[string]interface{}{
			"query":    query,
			"category": category,
			"official": official,
			"verified": verified,
			"featured": featured,
		},
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

	// Fetch plugin details from marketplace database
	var marketplacePlugin models.PluginMarketplace
	query := h.db.Unscoped().Table("plugin_registry").Where("name = ? AND status = ?", pluginName, "available")
	
	// If repository is specified, also filter by repository
	if repository != "" {
		query = query.Where("repository = ?", repository)
	}
	
	err := query.First(&marketplacePlugin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Plugin not found in marketplace",
			})
			return
		}
		log.Printf("Error fetching plugin details: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch plugin details",
		})
		return
	}

	// Get installation statistics (handle missing table gracefully)
	var installCount int64
	err = h.db.Model(&models.PluginInstallation{}).Where("plugin_name = ?", pluginName).Count(&installCount).Error
	if err != nil {
		// If table doesn't exist or other error, default to 0 installations
		installCount = 0
		log.Printf("Warning: Could not get installation count for plugin %s: %v", pluginName, err)
	}
	
	// Build detailed plugin response from database data
	pluginDetails := map[string]interface{}{
		"name":         marketplacePlugin.Name,
		"version":      marketplacePlugin.Version,
		"description":  marketplacePlugin.Description,
		"author":       marketplacePlugin.Author,
		"repository":   marketplacePlugin.Repository,
		"license":      marketplacePlugin.License,
		"type":         marketplacePlugin.Type,
		"downloads":    marketplacePlugin.DownloadCount,
		"install_count": installCount,
		"installs":     marketplacePlugin.InstallCount,
		"verified":     marketplacePlugin.IsVerified,
		"approved":     marketplacePlugin.IsApproved,
		"permissions":  marketplacePlugin.Permissions,
		"dependencies": marketplacePlugin.Dependencies,
		"ui_config":    marketplacePlugin.UIConfig,
		"main_file":    marketplacePlugin.MainFile,
		"checksum":     marketplacePlugin.Checksum,
		"signature":    marketplacePlugin.Signature,
		"registry_source": marketplacePlugin.RegistrySource,
		"source_metadata": marketplacePlugin.SourceMetadata,
		"published_at":    marketplacePlugin.PublishedAt,
		"deprecated_at":   marketplacePlugin.DeprecatedAt,
		"status":       marketplacePlugin.Status,
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

// ==================================================
// PROJECT-LEVEL PLUGIN MANAGEMENT API ENDPOINTS
// ==================================================

// GetAvailablePlugins returns plugins that are enabled by superadmin for a project
func (h *PluginHandler) GetAvailablePlugins(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Check if user is admin for the project
	if userRole != "admin" && userRole != "superadmin" {
		h.logPluginAction(c, "list_available", "", "", "", userID, userEmail, false, "Admin access required")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	// Verify project exists and user has access
	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get plugins that are enabled by admin (have system-wide installation with enabled status)
	var availablePlugins []models.PluginRegistry
	err = h.db.Unscoped().Raw(`
		SELECT DISTINCT pr.* 
		FROM plugin_registry pr 
		WHERE pr.name IN (
			SELECT DISTINCT plugin_name 
			FROM plugin_installations 
			WHERE deleted_at IS NULL 
			AND status = 'enabled'
			AND project_id = 0
		) 
		AND pr.is_approved = true 
		ORDER BY pr.name
	`).Scan(&availablePlugins).Error
	
	log.Printf("DEBUG: Found %d available plugins", len(availablePlugins))
	for i, plugin := range availablePlugins {
		log.Printf("DEBUG: Plugin %d: %s (approved: %v, status: %s)", i+1, plugin.Name, plugin.IsApproved, plugin.Status)
	}
	
	if err != nil {
		log.Printf("Error fetching available plugins: %v", err)
		h.logPluginAction(c, "list_available", "", "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch available plugins",
		})
		return
	}

	// Get project's current plugin installations to mark status
	var projectInstallations []models.PluginInstallation
	h.db.Where("project_id = ? AND deleted_at IS NULL", projectIDInt).Find(&projectInstallations)
	
	// Create map for quick lookup
	installationMap := make(map[string]models.PluginInstallation)
	for _, installation := range projectInstallations {
		installationMap[installation.PluginName] = installation
	}

	// Convert to response format
	plugins := make([]map[string]interface{}, 0, len(availablePlugins))
	for _, plugin := range availablePlugins {
		installation, isInstalled := installationMap[plugin.Name]
		
		pluginEntry := map[string]interface{}{
			"name":         plugin.Name,
			"version":      plugin.Version,
			"description":  plugin.Description,
			"author":       plugin.Author,
			"repository":   plugin.Repository,
			"license":      plugin.License,
			"type":         plugin.Type,
			"permissions":  plugin.Permissions,
			"dependencies": plugin.Dependencies,
			"ui_config":    plugin.UIConfig,
			"is_installed": isInstalled,
			"is_enabled":   isInstalled && installation.Status == "enabled",
			"status":       "available", // All returned plugins are available for this project
		}
		
		// Add installation-specific info if installed
		if isInstalled {
			pluginEntry["installed_at"] = installation.InstalledAt
			pluginEntry["installation_status"] = installation.Status
			pluginEntry["config"] = installation.Config
		}
		
		plugins = append(plugins, pluginEntry)
	}

	h.logPluginAction(c, "list_available", "", "", "", userID, userEmail, true, "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": plugins,
		"count":   len(plugins),
	})
}

// GetInstalledPlugins returns installed plugins for a specific project
func (h *PluginHandler) GetInstalledPlugins(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	// Check if user is admin for the project
	if userRole != "admin" && userRole != "superadmin" {
		h.logPluginAction(c, "list_installed", "", "", "", userID, userEmail, false, "Admin access required")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	projectID := c.Param("id")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get installed plugins for this project
	var installations []models.PluginInstallation
	err = h.db.Where("project_id = ?", projectIDInt).Find(&installations).Error
	if err != nil {
		log.Printf("Error fetching installed plugins: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch installed plugins",
		})
		return
	}

	// Convert to response format
	plugins := make([]Plugin, 0, len(installations))
	for _, installation := range installations {
		plugin := h.convertInstallationToPlugin(installation)
		plugins = append(plugins, plugin)
	}

	h.logPluginAction(c, "list_installed", "", "", "", userID, userEmail, true, "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"plugins": plugins,
		"count":   len(plugins),
	})
}

// InstallPluginToProject installs a plugin to a specific project
func (h *PluginHandler) InstallPluginToProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin installation"
		h.logPluginAction(c, "install_project", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "install_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Check if plugin exists in marketplace
	var marketplacePlugin models.PluginMarketplace
	err = h.db.Unscoped().Table("plugin_registry").Where("name = ? AND status = ?", pluginName, "available").First(&marketplacePlugin).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			h.logPluginAction(c, "install_project", pluginName, "not_found", "not_found", userID, userEmail, false, "Plugin not found in marketplace")
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Plugin not found in marketplace",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Check if already installed
	var existingInstallation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&existingInstallation).Error
	if err == nil {
		h.logPluginAction(c, "install_project", pluginName, existingInstallation.Status, existingInstallation.Status, userID, userEmail, false, "Plugin already installed")
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
		PluginVersion:    marketplacePlugin.Version,
		ProjectID:        uint(projectIDInt),
		Status:           "disabled", // Disabled by default for security
		InstallationPath: fmt.Sprintf("./plugins/%s", pluginName),
		InstalledBy:      uint(userIDInt),
		InstalledAt:      time.Now(),
		Config:          make(map[string]interface{}),
		Environment:     make(map[string]interface{}),
	}

	err = h.db.Create(&installation).Error
	if err != nil {
		h.logPluginAction(c, "install_project", pluginName, "uninstalled", "uninstalled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to create installation record",
		})
		return
	}

	// Create plugin state record
	state := models.PluginState{
		PluginName:     pluginName,
		ProjectID:      uint(projectIDInt),
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
	h.recordPluginDownload(pluginName, marketplacePlugin.Version, uint(projectIDInt), uint(userIDInt), marketplacePlugin.Repository, "completed", c)

	// Success audit log
	h.logPluginAction(c, "install_project", pluginName, "uninstalled", "disabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin installed successfully (disabled by default)",
		"plugin":  pluginName,
		"status":  "disabled",
	})
}

// EnablePluginForProject enables a plugin for a specific project
func (h *PluginHandler) EnablePluginForProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "enable_project", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "enable_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "enable_project", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "enable_project", pluginName, "", "", userID, userEmail, false, err.Error())
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
		h.logPluginAction(c, "enable_project", pluginName, currentStatus, "disabled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to enable plugin",
		})
		return
	}

	// Update plugin state
	h.updatePluginState(uint(projectIDInt), pluginName, "enabled", userID)

	// Success audit log
	h.logPluginAction(c, "enable_project", pluginName, currentStatus, "enabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin enabled successfully",
	})
}

// DisablePluginForProject disables a plugin for a specific project
func (h *PluginHandler) DisablePluginForProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "disable_project", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "disable_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "disable_project", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "disable_project", pluginName, "", "", userID, userEmail, false, err.Error())
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
		h.logPluginAction(c, "disable_project", pluginName, currentStatus, "enabled", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to disable plugin",
		})
		return
	}

	// Update plugin state
	h.updatePluginState(uint(projectIDInt), pluginName, "disabled", userID)

	// Success audit log
	h.logPluginAction(c, "disable_project", pluginName, currentStatus, "disabled", userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin disabled successfully",
	})
}

// UninstallPluginFromProject removes a plugin from a specific project
func (h *PluginHandler) UninstallPluginFromProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin operations"
		h.logPluginAction(c, "uninstall_project", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "uninstall_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "uninstall_project", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "uninstall_project", pluginName, "", "", userID, userEmail, false, err.Error())
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
		h.logPluginAction(c, "uninstall_project", pluginName, currentStatus, currentStatus, userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to uninstall plugin",
		})
		return
	}

	// Delete plugin state record
	h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).Delete(&models.PluginState{})

	// Clean up plugin files (if they exist)
	pluginPath := filepath.Join("./plugins", pluginName)
	if _, err := os.Stat(pluginPath); !os.IsNotExist(err) {
		os.RemoveAll(pluginPath)
	}

	// Success audit log
	h.logPluginAction(c, "uninstall_project", pluginName, currentStatus, "uninstalled", userID, userEmail, true, "")
	
	log.Printf("Plugin %s uninstalled from project %s by %s", pluginName, projectID, userEmail)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin uninstalled successfully",
	})
}

// UpdatePluginConfigForProject updates plugin configuration for a specific project
func (h *PluginHandler) UpdatePluginConfigForProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		errMsg := "Admin access required for plugin configuration"
		h.logPluginAction(c, "configure_project", "", "", "", userID, userEmail, false, errMsg)
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   errMsg,
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	var req struct {
		Config      map[string]interface{} `json:"config"`
		Environment map[string]interface{} `json:"environment"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logPluginAction(c, "configure_project", pluginName, "", "", userID, userEmail, false, "Invalid request data")
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "configure_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get current installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "configure_project", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "configure_project", pluginName, "", "", userID, userEmail, false, err.Error())
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
		h.logPluginAction(c, "configure_project", pluginName, installation.Status, installation.Status, userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update plugin configuration",
		})
		return
	}

	// Success audit log
	h.logPluginAction(c, "configure_project", pluginName, installation.Status, installation.Status, userID, userEmail, true, "")
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Plugin configuration updated successfully",
	})
}

// GetPluginStatusForProject returns plugin status for a specific project
func (h *PluginHandler) GetPluginStatusForProject(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")
	
	if userRole != "admin" && userRole != "superadmin" {
		h.logPluginAction(c, "status_project", "", "", "", userID, userEmail, false, "Admin access required")
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	projectID := c.Param("id")
	pluginName := c.Param("plugin_name")

	if projectID == "" || pluginName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and plugin name are required",
		})
		return
	}

	// Security: Validate plugin name
	if err := h.validatePluginName(pluginName); err != nil {
		h.logPluginAction(c, "status_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid project ID",
		})
		return
	}

	// Verify project exists and user has access
	var project models.Project
	err = h.db.First(&project, projectIDInt).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "Project not found",
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get plugin installation
	var installation models.PluginInstallation
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&installation).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			errMsg := "Plugin not installed"
			h.logPluginAction(c, "status_project", pluginName, "not_installed", "not_installed", userID, userEmail, false, errMsg)
			c.JSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   errMsg,
			})
			return
		}
		h.logPluginAction(c, "status_project", pluginName, "", "", userID, userEmail, false, err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Database error",
		})
		return
	}

	// Get plugin state
	var state models.PluginState
	err = h.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectIDInt).First(&state).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		log.Printf("Error fetching plugin state: %v", err)
	}

	// Build status response
	status := map[string]interface{}{
		"plugin_name":       installation.PluginName,
		"plugin_version":    installation.PluginVersion,
		"installation_status": installation.Status,
		"installed_at":      installation.InstalledAt,
		"last_enabled_at":   installation.LastEnabledAt,
		"last_disabled_at":  installation.LastDisabledAt,
		"config":           installation.Config,
		"environment":      installation.Environment,
		"error_message":    installation.ErrorMessage,
		"last_error_at":    installation.LastErrorAt,
	}

	if err == nil {
		status["current_status"] = state.CurrentStatus
		status["process_id"] = state.ProcessID
		status["port"] = state.Port
		status["last_health_check"] = state.LastHealthCheck
		status["health_status"] = state.HealthStatus
		status["health_details"] = state.HealthDetails
		status["cpu_usage"] = state.CPUUsage
		status["memory_usage"] = state.MemoryUsage
		status["uptime_seconds"] = state.UptimeSeconds
		status["state_changed_at"] = state.StateChangedAt
	}

	h.logPluginAction(c, "status_project", pluginName, "", "", userID, userEmail, true, "")

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"status":  status,
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

// Helper function to convert map[string]string to map[string]interface{}
func convertToInterface(deps map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range deps {
		result[k] = v
	}
	return result
}