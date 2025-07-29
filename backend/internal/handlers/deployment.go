package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeploymentHandler handles deployment operations
type DeploymentHandler struct {
	db                *gorm.DB
	cfg               *config.Config
	deploymentService *services.DeploymentService
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(db *gorm.DB, cfg *config.Config) *DeploymentHandler {
	return &DeploymentHandler{
		db:                db, 
		cfg:               cfg,
		deploymentService: services.NewDeploymentService(db),
	}
}

// CreateDeploymentRequest represents a deployment creation request
type CreateDeploymentRequest struct {
	Name               string                 `json:"name" binding:"required"`
	Description        string                 `json:"description"`
	GitHubRepositoryID uint                   `json:"github_repository_id" binding:"required"`
	WebServerID        uint                   `json:"web_server_id" binding:"required"`
	Domain             string                 `json:"domain"`
	Subdomain          string                 `json:"subdomain"`
	Port               int                    `json:"port"`
	Environment        map[string]interface{} `json:"environment"`
	BuildCommand       string                 `json:"build_command"`
	StartCommand       string                 `json:"start_command"`
	Branch             string                 `json:"branch"`
	IsAutoDeployEnabled bool                  `json:"is_auto_deploy_enabled"`
}

// UpdateDeploymentRequest represents a deployment update request
type UpdateDeploymentRequest struct {
	Name               string                 `json:"name"`
	Description        string                 `json:"description"`
	Domain             string                 `json:"domain"`
	Subdomain          string                 `json:"subdomain"`
	Port               *int                   `json:"port"`
	Environment        map[string]interface{} `json:"environment"`
	BuildCommand       string                 `json:"build_command"`
	StartCommand       string                 `json:"start_command"`
	Branch             string                 `json:"branch"`
	IsAutoDeployEnabled *bool                 `json:"is_auto_deploy_enabled"`
}

// DeployRequest represents a manual deployment trigger request
type DeployRequest struct {
	CommitHash string `json:"commit_hash"`
	Branch     string `json:"branch"`
}

// ListDeployments returns all deployments for a project
func (h *DeploymentHandler) ListDeployments(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var deployments []models.Deployment
	if err := h.db.Where("project_id = ?", uint(projectID)).
		Find(&deployments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	// Manual loading of relations to avoid GORM naming issues
	for i := range deployments {
		// Load GitHub Repository
		if deployments[i].GitHubRepositoryID > 0 {
			var repo models.GitHubRepository
			if err := h.db.First(&repo, deployments[i].GitHubRepositoryID).Error; err == nil {
				deployments[i].GitHubRepository = repo
			}
		}
		
		// Load Web Server
		if deployments[i].WebServerID > 0 {
			var server models.WebServer
			if err := h.db.First(&server, deployments[i].WebServerID).Error; err == nil {
				deployments[i].WebServer = server
			}
		}
	}

	c.JSON(http.StatusOK, deployments)
}

// CreateDeployment creates a new deployment configuration
func (h *DeploymentHandler) CreateDeployment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify GitHub repository exists and belongs to the project
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", req.GitHubRepositoryID, uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify GitHub repository"})
		}
		return
	}

	// Verify web server exists and belongs to the project
	var webServer models.WebServer
	if err := h.db.Where("id = ? AND project_id = ?", req.WebServerID, uint(projectID)).First(&webServer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Web server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify web server"})
		}
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = repository.AppPort
	}
	if req.BuildCommand == "" {
		req.BuildCommand = repository.BuildCommand
	}
	if req.StartCommand == "" {
		req.StartCommand = repository.StartCommand
	}
	if req.Branch == "" {
		req.Branch = repository.Branch
	}

	// Generate initial version
	version := fmt.Sprintf("v1.0.0-%d", time.Now().Unix())

	// Auto-detect PhotoPortfolio template and setup environment
	if req.Environment == nil {
		req.Environment = make(map[string]interface{})
	}
	
	// Check if this is a PhotoPortfolio deployment
	isPhotoPortfolio := h.isPhotoPortfolioRepository(repository)
	if isPhotoPortfolio {
		// Auto-setup PhotoPortfolio template
		if err := h.setupPhotoPortfolioTemplate(uint(projectID)); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to setup PhotoPortfolio template: " + err.Error()})
			return
		}
		
		// Configure PhotoPortfolio environment variables
		h.configurePhotoPortfolioEnvironment(req.Environment, uint(projectID), req.Domain, req.Port)
	}

	// Create deployment configuration
	deployment := models.Deployment{
		Name:               req.Name,
		Description:        req.Description,
		Version:            version,
		Domain:             req.Domain,
		Subdomain:          req.Subdomain,
		Port:               req.Port,
		Environment:        req.Environment,
		BuildCommand:       req.BuildCommand,
		StartCommand:       req.StartCommand,
		Branch:             req.Branch,
		Status:             "configured",
		ProjectID:          uint(projectID),
		GitHubRepositoryID: req.GitHubRepositoryID,
		WebServerID:        req.WebServerID,
		IsAutoDeployEnabled: req.IsAutoDeployEnabled,
		TriggerBranch:      req.Branch,
	}

	if err := h.db.Create(&deployment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create deployment"})
		return
	}

	// Load relations for response
	if err := h.db.Preload("GitHubRepository").Preload("WebServer").First(&deployment, deployment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load deployment details"})
		return
	}

	c.JSON(http.StatusCreated, deployment)
}

// GetDeployment returns a specific deployment
func (h *DeploymentHandler) GetDeployment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	deploymentID, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", uint(deploymentID), uint(projectID)).
		Preload("GitHubRepository").
		Preload("WebServer").
		First(&deployment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployment"})
		}
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// UpdateDeployment updates a deployment configuration
func (h *DeploymentHandler) UpdateDeployment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	deploymentID, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req UpdateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find deployment
	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", uint(deploymentID), uint(projectID)).First(&deployment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployment"})
		}
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.Domain != "" {
		updates["domain"] = req.Domain
	}
	if req.Subdomain != "" {
		updates["subdomain"] = req.Subdomain
	}
	if req.Port != nil {
		updates["port"] = *req.Port
	}
	if req.Environment != nil {
		updates["environment"] = req.Environment
	}
	if req.BuildCommand != "" {
		updates["build_command"] = req.BuildCommand
	}
	if req.StartCommand != "" {
		updates["start_command"] = req.StartCommand
	}
	if req.Branch != "" {
		updates["branch"] = req.Branch
		updates["trigger_branch"] = req.Branch
	}
	if req.IsAutoDeployEnabled != nil {
		updates["is_auto_deploy_enabled"] = *req.IsAutoDeployEnabled
	}

	if err := h.db.Model(&deployment).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment"})
		return
	}

	// Reload with relations
	if err := h.db.Preload("GitHubRepository").Preload("WebServer").First(&deployment, deployment.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated deployment"})
		return
	}

	c.JSON(http.StatusOK, deployment)
}

// DeleteDeployment deletes a deployment
func (h *DeploymentHandler) DeleteDeployment(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	deploymentID, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(deploymentID), uint(projectID)).Delete(&models.Deployment{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete deployment"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Deployment deleted successfully"})
}

// Deploy triggers a manual deployment
func (h *DeploymentHandler) Deploy(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	deploymentID, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var req DeployRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find deployment with relations
	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", uint(deploymentID), uint(projectID)).
		Preload("GitHubRepository").
		Preload("WebServer").
		Preload("WebServer.SSHKey").
		First(&deployment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployment"})
		}
		return
	}

	// Set defaults
	if req.Branch == "" {
		req.Branch = deployment.Branch
	}

	// Update deployment status
	updates := map[string]interface{}{
		"status":     "pending",
		"commit_hash": req.CommitHash,
		"branch":     req.Branch,
		"build_logs": "",
		"deploy_logs": "",
		"error_logs": "",
	}

	if err := h.db.Model(&deployment).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update deployment status"})
		return
	}

	// Start real deployment process in background
	go h.executeRealDeployment(deployment, req.CommitHash, req.Branch)

	c.JSON(http.StatusOK, gin.H{
		"message": "Deployment started",
		"status":  "pending",
	})
}

// GetLogs returns deployment logs
func (h *DeploymentHandler) GetLogs(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	deploymentID, err := strconv.ParseUint(c.Param("deployment_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment ID"})
		return
	}

	var deployment models.Deployment
	if err := h.db.Where("id = ? AND project_id = ?", uint(deploymentID), uint(projectID)).First(&deployment).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Deployment not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployment"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"build_logs":  deployment.BuildLogs,
		"deploy_logs": deployment.DeployLogs,
		"error_logs":  deployment.ErrorLogs,
		"status":      deployment.Status,
	})
}

// executeRealDeployment performs a real deployment using the deployment service
func (h *DeploymentHandler) executeRealDeployment(deployment models.Deployment, commitHash, branch string) {
	// Execute real deployment
	result := h.deploymentService.ExecuteDeployment(deployment, commitHash, branch)
	
	// If deployment failed and we don't have detailed logs, add generic error
	if !result.Success && result.ErrorLogs == "" {
		result.ErrorLogs = "Deployment failed with unknown error"
	}
	
	// Log deployment result for debugging
	if result.Success {
		fmt.Printf("Deployment %d completed successfully in %dms (build: %dms, deploy: %dms)\n", 
			deployment.ID, result.BuildTime+result.DeployTime, result.BuildTime, result.DeployTime)
	} else {
		fmt.Printf("Deployment %d failed: %s\n", deployment.ID, result.ErrorLogs)
	}
}

// PhotoPortfolio Template Integration Functions

// isPhotoPortfolioRepository checks if a repository is a PhotoPortfolio project
func (h *DeploymentHandler) isPhotoPortfolioRepository(repository models.GitHubRepository) bool {
	// Check repository name patterns
	if repository.Name == "photoportfolio" || repository.Name == "photo-portfolio" {
		return true
	}
	
	// Check for PhotoPortfolio-specific files or configurations
	// This could be enhanced to check actual repository content via GitHub API
	
	// Check description for PhotoPortfolio keywords
	if repository.Description != "" {
		desc := repository.Description
		keywords := []string{"photoportfolio", "photo portfolio", "photography portfolio", "cloudbox photo"}
		for _, keyword := range keywords {
			if containsKeyword(desc, keyword) {
				return true
			}
		}
	}
	
	return false
}

// setupPhotoPortfolioTemplate sets up the PhotoPortfolio template for a project
func (h *DeploymentHandler) setupPhotoPortfolioTemplate(projectID uint) error {
	// Check if template is already set up
	var existingCollections []models.Collection
	if err := h.db.Where("project_id = ?", projectID).Find(&existingCollections).Error; err != nil {
		return fmt.Errorf("failed to check existing collections: %v", err)
	}
	
	// If collections already exist, skip setup
	if len(existingCollections) > 0 {
		fmt.Printf("PhotoPortfolio template already set up for project %d\n", projectID)
		return nil
	}
	
	// Create template handler to set up collections
	templateHandler := NewTemplateHandler(h.db, h.cfg)
	
	// Get PhotoPortfolio template definition
	template := templateHandler.getPhotoPortfolioTemplate()
	
	// Setup all collections
	for _, collectionTemplate := range template.Collections {
		_, err := templateHandler.setupCollection(projectID, collectionTemplate)
		if err != nil {
			return fmt.Errorf("failed to setup collection %s: %v", collectionTemplate.Name, err)
		}
	}
	
	fmt.Printf("PhotoPortfolio template successfully set up for project %d\n", projectID)
	return nil
}

// configurePhotoPortfolioEnvironment sets up environment variables for PhotoPortfolio
func (h *DeploymentHandler) configurePhotoPortfolioEnvironment(env map[string]interface{}, projectID uint, domain string, port int) {
	// Get project details
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		fmt.Printf("Warning: Could not load project %d for environment setup: %v\n", projectID, err)
		return
	}
	
	// Get project API key
	var apiKey models.APIKey
	if err := h.db.Where("project_id = ? AND is_active = true", projectID).First(&apiKey).Error; err != nil {
		fmt.Printf("Warning: Could not find API key for project %d: %v\n", projectID, err)
		return
	}
	
	// CloudBox connection settings
	cloudboxEndpoint := "http://localhost:8080" // This should come from config
	if h.cfg.BaseURL != "" {
		cloudboxEndpoint = h.cfg.BaseURL
	}
	
	// Set PhotoPortfolio environment variables
	env["CLOUDBOX_ENDPOINT"] = cloudboxEndpoint
	env["CLOUDBOX_PROJECT_SLUG"] = project.Slug
	env["CLOUDBOX_PROJECT_ID"] = fmt.Sprintf("%d", projectID)
	env["CLOUDBOX_API_KEY"] = apiKey.Key
	
	// API Configuration
	env["VITE_API_URL"] = fmt.Sprintf("%s/p/%s/api", cloudboxEndpoint, project.Slug)
	
	// Application Settings
	env["APP_TITLE"] = "PhotoPortfolio"
	env["APP_DESCRIPTION"] = "Professional Photography Portfolio powered by CloudBox"
	env["ANALYTICS_ENABLED"] = "true"
	
	// Docker Production Settings
	if domain != "" {
		env["DOMAIN"] = domain
	} else {
		env["DOMAIN"] = "localhost"
	}
	
	if port != 0 {
		env["WEB_PORT"] = fmt.Sprintf("%d", port)
	} else {
		env["WEB_PORT"] = "3000"
	}
	
	env["PROJECT_PREFIX"] = fmt.Sprintf("%s-prod", project.Slug)
	env["NETWORK_NAME"] = "cloudbox-network"
	
	// Build Arguments (for browser access)
	env["VITE_CLOUDBOX_ENDPOINT"] = cloudboxEndpoint
	env["VITE_CLOUDBOX_PROJECT_SLUG"] = project.Slug
	env["VITE_CLOUDBOX_PROJECT_ID"] = fmt.Sprintf("%d", projectID)
	env["VITE_CLOUDBOX_API_KEY"] = apiKey.Key
	env["VITE_API_URL"] = fmt.Sprintf("%s/p/%s/api", cloudboxEndpoint, project.Slug)
	
	fmt.Printf("PhotoPortfolio environment configured for project %d (%s)\n", projectID, project.Slug)
}

// containsKeyword checks if a string contains a substring (case-insensitive)
func containsKeyword(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// HandleWebhook handles GitHub webhook requests for automatic deployments
func (h *DeploymentHandler) HandleWebhook(c *gin.Context) {
	// Get repository ID from URL parameter
	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	// Find the GitHub repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ?", uint(repoID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository"})
		}
		return
	}

	// Log webhook details for debugging
	fmt.Printf("Webhook received for repo %d\n", repoID)
	fmt.Printf("Content-Type: %s\n", c.GetHeader("Content-Type"))
	fmt.Printf("X-GitHub-Event: %s\n", c.GetHeader("X-GitHub-Event"))
	fmt.Printf("X-Hub-Signature-256: %s\n", c.GetHeader("X-Hub-Signature-256"))

	// Check if this is a push event first (before parsing JSON)
	eventType := c.GetHeader("X-GitHub-Event")
	if eventType != "push" {
		// Only handle push events for now
		c.JSON(http.StatusOK, gin.H{"message": "Event ignored", "event": eventType})
		return
	}

	// Read raw body for debugging
	body, err := c.GetRawData()
	if err != nil {
		fmt.Printf("Failed to read webhook body: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	fmt.Printf("Webhook body length: %d bytes\n", len(body))
	if len(body) > 0 {
		previewLen := 200
		if len(body) < previewLen {
			previewLen = len(body)
		}
		fmt.Printf("Webhook body preview: %s\n", string(body[:previewLen]))
	}

	// Parse the webhook payload
	var payload map[string]interface{}
	if len(body) == 0 {
		fmt.Printf("Empty webhook body, creating default payload\n")
		payload = map[string]interface{}{}
	} else {
		if err := json.Unmarshal(body, &payload); err != nil {
			fmt.Printf("Failed to parse webhook JSON: %v\n", err)
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON payload", "details": err.Error()})
			return
		}
	}

	// Verify webhook secret if provided
	webhookSecret := c.GetHeader("X-Hub-Signature-256")
	if webhookSecret != "" {
		// In a full implementation, verify the webhook signature here
		// For now, we'll just acknowledge the webhook
		fmt.Printf("Webhook signature provided but not verified\n")
	}

	// Extract commit information from payload with fallbacks
	commitHash := "unknown"
	branch := repository.Branch // Default to repository's configured branch
	
	if ref, ok := payload["ref"].(string); ok {
		// Extract branch name from ref (e.g., "refs/heads/main" -> "main")
		if strings.HasPrefix(ref, "refs/heads/") {
			branch = strings.TrimPrefix(ref, "refs/heads/")
		}
	}
	
	if headCommit, ok := payload["head_commit"].(map[string]interface{}); ok {
		if id, ok := headCommit["id"].(string); ok {
			commitHash = id
		}
	}

	// If we couldn't extract commit info from payload, generate a fallback
	if commitHash == "unknown" {
		commitHash = fmt.Sprintf("webhook-%d", time.Now().Unix())
		fmt.Printf("No commit hash found in payload, using fallback: %s\n", commitHash)
	}

	fmt.Printf("Extracted branch: %s, commit: %s\n", branch, commitHash)

	// Check if the branch matches the repository's configured branch
	if branch != "" && branch != repository.Branch {
		fmt.Printf("Branch mismatch: got %s, expected %s\n", branch, repository.Branch)
		c.JSON(http.StatusOK, gin.H{
			"message": "Branch ignored", 
			"branch": branch,
			"configured_branch": repository.Branch,
		})
		return
	}

	// Update repository with pending commit information
	updates := map[string]interface{}{
		"pending_commit_hash":   commitHash,
		"pending_commit_branch": branch,
		"has_pending_update":    true,
		"last_sync_at":          time.Now(),
	}
	
	if err := h.db.Model(&repository).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository status"})
		return
	}

	// Send notification to project owner about available update
	fmt.Printf("Sending update notification for repository %d\n", repository.ID)
	if err := h.sendUpdateNotification(repository, commitHash, branch, payload); err != nil {
		// Log error but don't fail the webhook - GitHub expects 200 response
		fmt.Printf("Failed to send update notification for repository %d: %v\n", repository.ID, err)
	} else {
		fmt.Printf("Update notification sent successfully\n")
	}

	response := gin.H{
		"message": "Webhook processed successfully",
		"repository_id": repository.ID,
		"repository_name": repository.Name,
		"commit": commitHash,
		"branch": branch,
		"has_pending_update": true,
		"event_type": eventType,
	}

	fmt.Printf("Webhook response: %+v\n", response)
	c.JSON(http.StatusOK, response)
}

// sendUpdateNotification sends a notification message to the project owner about available updates
func (h *DeploymentHandler) sendUpdateNotification(repository models.GitHubRepository, commitHash, branch string, payload map[string]interface{}) error {
	// Extract commit information from payload
	commitMessage := ""
	commitAuthor := ""
	
	if headCommit, ok := payload["head_commit"].(map[string]interface{}); ok {
		if message, ok := headCommit["message"].(string); ok {
			commitMessage = message
		}
		if author, ok := headCommit["author"].(map[string]interface{}); ok {
			if name, ok := author["name"].(string); ok {
				commitAuthor = name
			}
		}
	}
	
	// Truncate commit message if too long
	if len(commitMessage) > 100 {
		commitMessage = commitMessage[:97] + "..."
	}
	
	// Find or create a system notifications channel for this project
	channelID, err := h.getOrCreateNotificationChannel(repository.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get notification channel: %v", err)
	}
	
	// Create notification message
	messageContent := fmt.Sprintf("🔄 **Update Available for %s**\n\n", repository.Name)
	messageContent += fmt.Sprintf("**Branch:** %s\n", branch)
	messageContent += fmt.Sprintf("**Commit:** `%s`\n", commitHash[:8])
	if commitAuthor != "" {
		messageContent += fmt.Sprintf("**Author:** %s\n", commitAuthor)
	}
	if commitMessage != "" {
		messageContent += fmt.Sprintf("**Message:** %s\n", commitMessage)
	}
	messageContent += "\n✨ Ready to deploy when you are!"
	
	// Create message record
	message := models.Message{
		ID:        generateUUID(),
		Content:   messageContent,
		Type:      "system",
		ChannelID: channelID,
		UserID:    "system",
		ProjectID: repository.ProjectID,
		Metadata: map[string]interface{}{
			"type":              "github_update",
			"repository_id":     repository.ID,
			"repository_name":   repository.Name,
			"commit_hash":       commitHash,
			"branch":           branch,
			"can_deploy":       true,
			"github_url":       fmt.Sprintf("https://github.com/%s/commit/%s", repository.FullName, commitHash),
		},
	}
	
	if err := h.db.Create(&message).Error; err != nil {
		return fmt.Errorf("failed to create notification message: %v", err)
	}
	
	// Update channel activity
	if err := h.db.Model(&models.Channel{}).Where("id = ?", channelID).Updates(map[string]interface{}{
		"last_activity":  time.Now(),
		"message_count":  gorm.Expr("message_count + 1"),
	}).Error; err != nil {
		return fmt.Errorf("failed to update channel activity: %v", err)
	}
	
	return nil
}

// getOrCreateNotificationChannel finds or creates a system notification channel for the project
func (h *DeploymentHandler) getOrCreateNotificationChannel(projectID uint) (string, error) {
	// Look for existing notifications channel
	var channel models.Channel
	err := h.db.Where("project_id = ? AND name = ? AND type = ?", projectID, "System Notifications", "system").First(&channel).Error
	
	if err == nil {
		return channel.ID, nil
	}
	
	if err != gorm.ErrRecordNotFound {
		return "", fmt.Errorf("failed to query notification channel: %v", err)
	}
	
	// Create new notifications channel
	channel = models.Channel{
		ID:          generateUUID(),
		Name:        "System Notifications",
		Description: "Automatic notifications from CloudBox system",
		Type:        "system",
		ProjectID:   projectID,
		CreatedBy:   "system",
		IsActive:    true,
		Settings: map[string]interface{}{
			"read_only": true,
			"system_channel": true,
		},
		LastActivity: time.Now(),
	}
	
	if err := h.db.Create(&channel).Error; err != nil {
		return "", fmt.Errorf("failed to create notification channel: %v", err)
	}
	
	return channel.ID, nil
}

// generateUUID generates a UUID string
func generateUUID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix())
}

// DeployPendingUpdate deploys a pending update for a repository
func (h *DeploymentHandler) DeployPendingUpdate(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	repoID, err := strconv.ParseUint(c.Param("repo_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	// Find the GitHub repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository"})
		}
		return
	}

	// Check if there's a pending update
	if !repository.HasPendingUpdate || repository.PendingCommitHash == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No pending updates available"})
		return
	}

	// Find deployments that use this repository
	var deployments []models.Deployment
	if err := h.db.Where("github_repository_id = ?", repository.ID).
		Preload("GitHubRepository").Preload("WebServer").Find(&deployments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	if len(deployments) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No deployments found for this repository"})
		return
	}

	// Update repository to mark update as deployed
	updates := map[string]interface{}{
		"last_commit_hash":      repository.PendingCommitHash,
		"pending_commit_hash":   "",
		"pending_commit_branch": "",
		"has_pending_update":    false,
		"last_sync_at":          time.Now(),
	}
	
	if err := h.db.Model(&repository).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update repository status"})
		return
	}

	// Start deployments for all associated deployments
	deployedCount := 0
	for _, deployment := range deployments {
		// Start deployment in background
		go h.executeRealDeployment(deployment, repository.PendingCommitHash, repository.PendingCommitBranch)
		deployedCount++
	}

	// Send deployment started notification
	if err := h.sendDeploymentStartedNotification(repository, repository.PendingCommitHash, deployedCount); err != nil {
		fmt.Printf("Failed to send deployment notification for repository %d: %v\n", repository.ID, err)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("Started %d deployments", deployedCount),
		"deployments_started": deployedCount,
		"commit": repository.PendingCommitHash,
		"repository_id": repository.ID,
	})
}

// sendDeploymentStartedNotification sends a notification when deployment is manually started
func (h *DeploymentHandler) sendDeploymentStartedNotification(repository models.GitHubRepository, commitHash string, deploymentCount int) error {
	// Find the system notifications channel
	channelID, err := h.getOrCreateNotificationChannel(repository.ProjectID)
	if err != nil {
		return fmt.Errorf("failed to get notification channel: %v", err)
	}
	
	// Create deployment started message
	messageContent := fmt.Sprintf("🚀 **Deployment Started for %s**\n\n", repository.Name)
	messageContent += fmt.Sprintf("**Commit:** `%s`\n", commitHash[:8])
	messageContent += fmt.Sprintf("**Deployments:** %d environment(s)\n", deploymentCount)
	messageContent += "\n⏳ Deployment in progress..."
	
	message := models.Message{
		ID:        generateUUID(),
		Content:   messageContent,
		Type:      "system",
		ChannelID: channelID,
		UserID:    "system",
		ProjectID: repository.ProjectID,
		Metadata: map[string]interface{}{
			"type":              "deployment_started",
			"repository_id":     repository.ID,
			"repository_name":   repository.Name,
			"commit_hash":       commitHash,
			"deployment_count":  deploymentCount,
		},
	}
	
	if err := h.db.Create(&message).Error; err != nil {
		return fmt.Errorf("failed to create deployment notification: %v", err)
	}
	
	// Update channel activity
	if err := h.db.Model(&models.Channel{}).Where("id = ?", channelID).Updates(map[string]interface{}{
		"last_activity":  time.Now(),
		"message_count":  gorm.Expr("message_count + 1"),
	}).Error; err != nil {
		return fmt.Errorf("failed to update channel activity: %v", err)
	}
	
	return nil
}