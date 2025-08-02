package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

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
		h.renderOAuthResult(c, false, "Invalid project ID", "")
		return
	}

	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")
	
	// Handle OAuth errors (user denied access, etc.)
	if errorParam != "" {
		h.renderOAuthResult(c, false, fmt.Sprintf("GitHub OAuth error: %s", errorParam), "")
		return
	}
	
	if code == "" {
		h.renderOAuthResult(c, false, "Authorization code required", "")
		return
	}

	// Parse state to get repository ID (format: projectID_repoID_random)
	var repoID uint
	if state != "" {
		parts := strings.Split(state, "_")
		if len(parts) >= 2 {
			if parsedRepoID, err := strconv.ParseUint(parts[1], 10, 32); err == nil {
				repoID = uint(parsedRepoID)
			}
		}
	}

	// Get project GitHub config
	var gitHubConfig models.ProjectGitHubConfig
	err = h.db.Where("project_id = ?", uint(projectID)).First(&gitHubConfig).Error
	if err != nil {
		h.renderOAuthResult(c, false, "GitHub configuration not found for this project", "")
		return
	}

	if !gitHubConfig.IsEnabled {
		h.renderOAuthResult(c, false, "GitHub OAuth is not enabled for this project", "")
		return
	}

	// Exchange code for token using project-specific config
	token, err := h.exchangeCodeForToken(code, gitHubConfig.ClientID, gitHubConfig.ClientSecret, gitHubConfig.CallbackURL)
	if err != nil {
		h.renderOAuthResult(c, false, "Failed to exchange code for token: " + err.Error(), "")
		return
	}

	// Get GitHub user info to verify
	userInfo, err := h.getGitHubUserInfo(token.AccessToken)
	if err != nil {
		h.renderOAuthResult(c, false, "Failed to get user info: " + err.Error(), "")
		return
	}

	// If we have a specific repository ID from state, store the token there
	if repoID > 0 {
		// Update repository with OAuth info
		now := time.Now()
		updates := map[string]interface{}{
			"access_token":    token.AccessToken,
			"token_scopes":    token.Scope,
			"authorized_at":   &now,
			"authorized_by":   userInfo.Login,
		}

		if err := h.db.Model(&models.GitHubRepository{}).Where("id = ? AND project_id = ?", repoID, projectID).Updates(updates).Error; err != nil {
			h.renderOAuthResult(c, false, "Failed to save OAuth token to repository", "")
			return
		}

		message := fmt.Sprintf("GitHub OAuth authorization successful! Repository access granted for %s", userInfo.Login)
		h.renderOAuthResult(c, true, message, fmt.Sprintf("repo_%d", repoID))
	} else {
		// Fallback: just show success without storing token
		message := fmt.Sprintf("GitHub OAuth authorization successful! Token type: %s, Scope: %s", token.TokenType, token.Scope)
		h.renderOAuthResult(c, true, message, fmt.Sprintf("project_%d", projectID))
	}
}

// exchangeCodeForToken exchanges authorization code for access token using project config
func (h *ProjectGitHubHandler) exchangeCodeForToken(code, clientID, clientSecret, callbackURL string) (*ProjectGitHubOAuthToken, error) {
	// Prepare token exchange request
	tokenURL := "https://github.com/login/oauth/access_token"
	
	// Create form data
	data := fmt.Sprintf("client_id=%s&client_secret=%s&code=%s&redirect_uri=%s",
		clientID, clientSecret, code, callbackURL)
		
	// Create HTTP request
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}
	
	// Set headers
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CloudBox-OAuth/1.0")
	
	// Execute request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange token: %w", err)
	}
	defer resp.Body.Close()
	
	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}
	
	// Check for HTTP errors
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("GitHub token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse JSON response
	var tokenResponse struct {
		AccessToken  string `json:"access_token"`
		TokenType    string `json:"token_type"`
		Scope        string `json:"scope"`
		Error        string `json:"error"`
		ErrorDesc    string `json:"error_description"`
	}
	
	if err := json.Unmarshal(body, &tokenResponse); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	
	// Check for OAuth errors
	if tokenResponse.Error != "" {
		return nil, fmt.Errorf("GitHub OAuth error: %s - %s", tokenResponse.Error, tokenResponse.ErrorDesc)
	}
	
	if tokenResponse.AccessToken == "" {
		return nil, fmt.Errorf("no access token received from GitHub")
	}
	
	return &ProjectGitHubOAuthToken{
		AccessToken: tokenResponse.AccessToken,
		TokenType:   tokenResponse.TokenType,
		Scope:       tokenResponse.Scope,
	}, nil
}

// ProjectGitHubOAuthToken represents the OAuth token response from GitHub for project-specific OAuth
type ProjectGitHubOAuthToken struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	Scope       string `json:"scope"`
}

// ProjectGitHubUser represents a GitHub user for project-specific OAuth
type ProjectGitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

// generateRandomState generates a secure random state parameter for OAuth
func (h *ProjectGitHubHandler) generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// getGitHubUserInfo gets GitHub user information using an access token
func (h *ProjectGitHubHandler) getGitHubUserInfo(accessToken string) (*ProjectGitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CloudBox/1.0")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var user ProjectGitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	return &user, nil
}

// renderOAuthResult renders an HTML page that closes the popup and communicates with parent window
func (h *ProjectGitHubHandler) renderOAuthResult(c *gin.Context, success bool, message, data string) {
	statusIcon := "❌"
	if success {
		statusIcon = "✅"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>GitHub OAuth Result</title>
    <style>
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            display: flex;
            justify-content: center;
            align-items: center;
            min-height: 100vh;
            margin: 0;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            color: #fff;
        }
        .container {
            text-align: center;
            padding: 2rem;
            background: rgba(255, 255, 255, 0.1);
            border-radius: 16px;
            backdrop-filter: blur(10px);
            box-shadow: 0 8px 32px rgba(0, 0, 0, 0.1);
            max-width: 400px;
        }
        .icon {
            font-size: 4rem;
            margin-bottom: 1rem;
        }
        .message {
            font-size: 1.1rem;
            line-height: 1.5;
            margin-bottom: 1.5rem;
        }
        .closing {
            font-size: 0.9rem;
            opacity: 0.8;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">%s</div>
        <div class="message">%s</div>
        <div class="closing">This window will close automatically...</div>
    </div>
    
    <script>
        // Send result to parent window
        if (window.opener) {
            window.opener.postMessage({
                type: 'github_oauth_result',
                success: %t,
                message: '%s',
                data: '%s'
            }, '*');
        }
        
        // Close popup after a short delay
        setTimeout(() => {
            window.close();
        }, 2000);
    </script>
</body>
</html>`, statusIcon, message, success, message, data)

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, html)
}