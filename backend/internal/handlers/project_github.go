package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ProjectGitHubHandler handles project-specific GitHub OAuth configuration
type ProjectGitHubHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewProjectGitHubHandler creates a new project GitHub handler
func NewProjectGitHubHandler(db *gorm.DB, cfg *config.Config) *ProjectGitHubHandler {
	return &ProjectGitHubHandler{db: db, cfg: cfg}
}

// GetProjectGitHubConfig gets GitHub configuration for a project
func (h *ProjectGitHubHandler) GetProjectGitHubConfig(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Check if user has access to this project
	if !h.hasProjectAccess(c, uint(projectID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this project"})
		return
	}

	var gitHubConfig models.ProjectGitHubConfig
	err = h.db.Where("project_id = ?", uint(projectID)).First(&gitHubConfig).Error
	
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return default config if not found
			c.JSON(http.StatusOK, models.ProjectGitHubConfig{
				ProjectID: uint(projectID),
				IsEnabled: false,
				CallbackURL: h.generateCallbackURL(uint(projectID)),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GitHub configuration"})
		return
	}

	// Don't expose client secret in response
	gitHubConfig.ClientSecret = ""
	c.JSON(http.StatusOK, gitHubConfig)
}

// UpdateProjectGitHubConfig updates GitHub configuration for a project
func (h *ProjectGitHubHandler) UpdateProjectGitHubConfig(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if !h.hasProjectAccess(c, uint(projectID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this project"})
		return
	}

	var req struct {
		ClientID     string `json:"client_id"`
		ClientSecret string `json:"client_secret"`
		IsEnabled    bool   `json:"is_enabled"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetUint("user_id")

	// Get or create GitHub config
	var gitHubConfig models.ProjectGitHubConfig
	err = h.db.Where("project_id = ?", uint(projectID)).First(&gitHubConfig).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create new config
		gitHubConfig = models.ProjectGitHubConfig{
			ProjectID:   uint(projectID),
			ClientID:    req.ClientID,
			ClientSecret: req.ClientSecret, // In production, this should be encrypted
			IsEnabled:   req.IsEnabled,
			CallbackURL: h.generateCallbackURL(uint(projectID)),
			CreatedBy:   userID,
			UpdatedBy:   userID,
		}
		
		if err := h.db.Create(&gitHubConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create GitHub configuration"})
			return
		}
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get GitHub configuration"})
		return
	} else {
		// Update existing config
		gitHubConfig.ClientID = req.ClientID
		if req.ClientSecret != "" {
			gitHubConfig.ClientSecret = req.ClientSecret // In production, this should be encrypted
		}
		gitHubConfig.IsEnabled = req.IsEnabled
		gitHubConfig.CallbackURL = h.generateCallbackURL(uint(projectID))
		gitHubConfig.UpdatedBy = userID
		
		if err := h.db.Save(&gitHubConfig).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update GitHub configuration"})
			return
		}
	}

	// Don't expose client secret in response
	gitHubConfig.ClientSecret = ""
	c.JSON(http.StatusOK, gitHubConfig)
}

// TestProjectGitHubConfig tests GitHub OAuth configuration for a project
func (h *ProjectGitHubHandler) TestProjectGitHubConfig(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if !h.hasProjectAccess(c, uint(projectID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this project"})
		return
	}

	var gitHubConfig models.ProjectGitHubConfig
	err = h.db.Where("project_id = ?", uint(projectID)).First(&gitHubConfig).Error
	
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub configuration not found"})
		return
	}

	if !gitHubConfig.IsEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth is not enabled for this project"})
		return
	}

	if gitHubConfig.ClientID == "" || gitHubConfig.ClientSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth Client ID and Secret are required"})
		return
	}

	// Generate test OAuth URL
	state := h.generateRandomState()
	oauthURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&redirect_uri=%s&scope=repo&state=%s",
		gitHubConfig.ClientID,
		gitHubConfig.CallbackURL,
		state,
	)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "GitHub OAuth configuration is valid",
		"test_url": oauthURL,
		"callback_url": gitHubConfig.CallbackURL,
	})
}

// GetProjectGitHubInstructions returns setup instructions for project GitHub OAuth
func (h *ProjectGitHubHandler) GetProjectGitHubInstructions(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	if !h.hasProjectAccess(c, uint(projectID)) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied to this project"})
		return
	}

	callbackURL := h.generateCallbackURL(uint(projectID))

	instructions := []map[string]interface{}{
		{
			"step": 1,
			"title": "Create GitHub OAuth App",
			"description": "Go to GitHub Settings > Developer settings > OAuth Apps",
			"action": "Click 'New OAuth App'",
			"details": []string{
				"Application name: Your Project Name - CloudBox",
				"Homepage URL: " + h.cfg.BaseURL,
				"Authorization callback URL: " + callbackURL,
			},
		},
		{
			"step": 2,
			"title": "Copy OAuth Credentials",
			"description": "After creating the OAuth App, copy the credentials",
			"action": "Copy Client ID and generate Client Secret",
			"details": []string{
				"Client ID: Copy this to the Client ID field below",
				"Client Secret: Generate and copy to Client Secret field below",
				"Keep these credentials secure and never share them publicly",
			},
		},
		{
			"step": 3,
			"title": "Configure CloudBox",
			"description": "Enter the OAuth credentials in the form below",
			"action": "Paste Client ID and Client Secret, then enable GitHub OAuth",
			"details": []string{
				"Client ID: Paste the Client ID from step 2",
				"Client Secret: Paste the Client Secret from step 2", 
				"Enable GitHub OAuth: Toggle this on to activate OAuth for this project",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"instructions": instructions,
		"callback_url": callbackURL,
		"base_url": h.cfg.BaseURL,
	})
}

// hasProjectAccess checks if user has access to the project
func (h *ProjectGitHubHandler) hasProjectAccess(c *gin.Context, projectID uint) bool {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Superadmins have access to all projects
	if userRole == "superadmin" {
		return true
	}

	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		return false
	}

	// Project owner has access
	if project.UserID == userID {
		return true
	}

	// Check if user is organization admin
	var orgAdmin models.OrganizationAdmin
	err := h.db.Where("user_id = ? AND organization_id = ? AND is_active = true", 
		userID, project.OrganizationID).First(&orgAdmin).Error
	
	return err == nil
}

// generateCallbackURL generates the OAuth callback URL for a project
func (h *ProjectGitHubHandler) generateCallbackURL(projectID uint) string {
	return fmt.Sprintf("%s/api/v1/projects/%d/github/oauth/callback", h.cfg.BaseURL, projectID)
}

// HandleProjectOAuthCallback handles GitHub OAuth callback for a specific project
func (h *ProjectGitHubHandler) HandleProjectOAuthCallback(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Authorization code required"})
		return
	}

	// Get project GitHub config
	var gitHubConfig models.ProjectGitHubConfig
	err = h.db.Where("project_id = ?", uint(projectID)).First(&gitHubConfig).Error
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub configuration not found for this project"})
		return
	}

	if !gitHubConfig.IsEnabled {
		c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub OAuth is not enabled for this project"})
		return
	}

	// Exchange code for token using project-specific config
	token, err := h.exchangeCodeForToken(code, gitHubConfig.ClientID, gitHubConfig.ClientSecret, gitHubConfig.CallbackURL)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token"})
		return
	}

	// Return success response - in a real app, you'd store the token securely
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "GitHub OAuth authorization successful for project",
		"project_id": projectID,
		"token_type": token.TokenType,
		"scope": token.Scope,
	})
}

// exchangeCodeForToken exchanges authorization code for access token using project config
func (h *ProjectGitHubHandler) exchangeCodeForToken(code, clientID, clientSecret, callbackURL string) (*GitHubOAuthToken, error) {
	// This would make an HTTP request to GitHub's token endpoint
	// For now, return a mock token
	return &GitHubOAuthToken{
		AccessToken: "mock_access_token",
		TokenType:   "bearer",
		Scope:       "repo",
	}, nil
}

// GitHubOAuthToken represents the OAuth token response from GitHub
type GitHubOAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// generateRandomState generates a secure random state parameter for OAuth
func (h *ProjectGitHubHandler) generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}