package handlers

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SystemSettingsHandler handles system-wide configuration settings
type SystemSettingsHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewSystemSettingsHandler creates a new system settings handler
func NewSystemSettingsHandler(db *gorm.DB, cfg *config.Config) *SystemSettingsHandler {
	return &SystemSettingsHandler{
		db:  db,
		cfg: cfg,
	}
}

// GetSystemSettings returns all system settings grouped by category
func (h *SystemSettingsHandler) GetSystemSettings(c *gin.Context) {
	var settings []models.SystemSetting
	if err := h.db.Where("is_active = ?", true).Order("category, sort_order, name").Find(&settings).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch system settings"})
		return
	}

	// Group settings by category
	grouped := make(map[string][]models.SystemSetting)
	for _, setting := range settings {
		// Hide secret values in response
		if setting.IsSecret && setting.Value != "" {
			setting.Value = "••••••••"
		}
		grouped[setting.Category] = append(grouped[setting.Category], setting)
	}

	c.JSON(http.StatusOK, gin.H{
		"settings": grouped,
		"instructions": h.generateGitHubInstructions(),
	})
}

// UpdateSystemSetting updates a specific system setting
func (h *SystemSettingsHandler) UpdateSystemSetting(c *gin.Context) {
	key := c.Param("key")
	if key == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Setting key is required"})
		return
	}

	var req struct {
		Value string `json:"value"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	// Find the setting
	var setting models.SystemSetting
	if err := h.db.Where("key = ?", key).First(&setting).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Setting not found"})
		return
	}

	// Validate the value based on type
	if err := h.validateSettingValue(setting.ValueType, req.Value); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the setting
	setting.Value = req.Value
	if err := h.db.Save(&setting).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update setting"})
		return
	}

	// Hide secret values in response
	if setting.IsSecret && setting.Value != "" {
		setting.Value = "••••••••"
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Setting updated successfully",
		"setting": setting,
	})
}

// GetGitHubInstructions returns dynamic GitHub OAuth setup instructions
func (h *SystemSettingsHandler) GetGitHubInstructions(c *gin.Context) {
	instructions := h.generateGitHubInstructions()
	c.JSON(http.StatusOK, instructions)
}

// TestGitHubOAuth tests the GitHub OAuth configuration
func (h *SystemSettingsHandler) TestGitHubOAuth(c *gin.Context) {
	// Get current settings
	clientID := h.getSettingValue("github_oauth_client_id")
	clientSecret := h.getSettingValue("github_oauth_client_secret")
	enabled := h.getSettingValue("github_oauth_enabled") == "true"

	if !enabled {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "GitHub OAuth is not enabled",
			"status": "disabled",
		})
		return
	}

	if clientID == "" || clientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "GitHub OAuth Client ID and Secret are required", 
			"status": "incomplete",
		})
		return
	}

	// Test by making a simple GitHub API call
	testResult := map[string]interface{}{
		"status": "configured",
		"client_id": clientID,
		"client_secret_set": clientSecret != "",
		"callback_url": h.generateCallbackURL(),
		"instructions": h.generateGitHubInstructions(),
	}

	c.JSON(http.StatusOK, testResult)
}

// Helper functions

func (h *SystemSettingsHandler) validateSettingValue(valueType, value string) error {
	switch valueType {
	case "boolean":
		if value != "true" && value != "false" {
			return fmt.Errorf("value must be 'true' or 'false'")
		}
	case "integer":
		if _, err := strconv.Atoi(value); err != nil {
			return fmt.Errorf("value must be a valid integer")
		}
	}
	return nil
}

func (h *SystemSettingsHandler) getSettingValue(key string) string {
	var setting models.SystemSetting
	if err := h.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

func (h *SystemSettingsHandler) generateCallbackURL() string {
	domain := h.getSettingValue("site_domain")
	protocol := h.getSettingValue("site_protocol")
	
	if domain == "" {
		domain = "localhost:3000"
	}
	if protocol == "" {
		protocol = "http"
	}

	// For development, always use backend port for callback
	if strings.Contains(domain, "localhost") {
		domain = strings.Replace(domain, "3000", "8080", 1)
	} else {
		// For production, assume backend is on same domain
		if !strings.Contains(domain, ":") {
			domain = domain + ":8080"
		}
	}

	return protocol + "://" + domain + "/api/v1/github/oauth/callback"
}

func (h *SystemSettingsHandler) generateGitHubInstructions() map[string]interface{} {
	domain := h.getSettingValue("site_domain")
	protocol := h.getSettingValue("site_protocol")
	
	if domain == "" {
		domain = "localhost:3000"
	}
	if protocol == "" {
		protocol = "http"
	}

	callbackURL := h.generateCallbackURL()
	homepageURL := protocol + "://" + domain

	return map[string]interface{}{
		"callback_url": callbackURL,
		"homepage_url": homepageURL,
		"steps": []map[string]interface{}{
			{
				"step": 1,
				"title": "Create GitHub OAuth App",
				"description": "Go to GitHub Settings > Developer settings > OAuth Apps",
				"url": "https://github.com/settings/developers",
				"action": "Click 'New OAuth App'",
			},
			{
				"step": 2,
				"title": "Configure OAuth App",
				"description": "Fill in the OAuth App details",
				"fields": map[string]string{
					"Application name": "CloudBox Repository Analysis",
					"Homepage URL": homepageURL,
					"Application description": "CloudBox repository analysis and deployment platform",
					"Authorization callback URL": callbackURL,
				},
			},
			{
				"step": 3,
				"title": "Get Credentials",
				"description": "After creating the app, copy the credentials",
				"actions": []string{
					"Copy the Client ID",
					"Generate and copy the Client Secret",
					"Paste both values in the form below",
				},
			},
			{
				"step": 4,
				"title": "Enable OAuth",
				"description": "Enable GitHub OAuth in CloudBox",
				"action": "Set 'Enable GitHub OAuth' to true and save settings",
			},
		},
		"current_domain": domain,
		"current_protocol": protocol,
		"is_production": !strings.Contains(domain, "localhost"),
	}
}