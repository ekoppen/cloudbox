package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GitHubHandler handles GitHub repository management
type GitHubHandler struct {
	db            *gorm.DB
	cfg           *config.Config
	githubService *services.GitHubService
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(db *gorm.DB, cfg *config.Config) *GitHubHandler {
	return &GitHubHandler{
		db:            db,
		cfg:           cfg,
		githubService: services.NewGitHubService(db),
	}
}

// CreateGitHubRepositoryRequest represents a request to add a GitHub repository
type CreateGitHubRepositoryRequest struct {
	Name         string                 `json:"name" binding:"required"`
	FullName     string                 `json:"full_name" binding:"required"` // owner/repo
	CloneURL     string                 `json:"clone_url" binding:"required"`
	Branch       string                 `json:"branch"`
	IsPrivate    bool                   `json:"is_private"`
	Description  string                 `json:"description"`
	SDKVersion   string                 `json:"sdk_version"`
	AppPort      int                    `json:"app_port"`
	BuildCommand string                 `json:"build_command"`
	StartCommand string                 `json:"start_command"`
	Environment  map[string]interface{} `json:"environment"`
	SSHKeyID     *uint                  `json:"ssh_key_id"` // SSH key for private repository access
}

// UpdateGitHubRepositoryRequest represents a request to update a GitHub repository
type UpdateGitHubRepositoryRequest struct {
	Name         string                 `json:"name"`
	Branch       string                 `json:"branch"`
	Description  string                 `json:"description"`
	SDKVersion   string                 `json:"sdk_version"`
	AppPort      *int                   `json:"app_port"`
	BuildCommand string                 `json:"build_command"`
	StartCommand string                 `json:"start_command"`
	Environment  map[string]interface{} `json:"environment"`
	IsActive     *bool                  `json:"is_active"`
	SSHKeyID     *uint                  `json:"ssh_key_id"` // SSH key for private repository access
}

// ProjectAnalysisRequest represents a request to analyze a repository
type ProjectAnalysisRequest struct {
	RepoURL  string `json:"repo_url" binding:"required"`
	Branch   string `json:"branch"`
	SSHKeyID *uint  `json:"ssh_key_id"` // SSH key for private repository access
}

// ProjectAnalysisResponse represents the analysis result
type ProjectAnalysisResponse struct {
	ProjectType    string                 `json:"project_type"`    // react, vue, angular, next, nuxt, etc.
	Framework      string                 `json:"framework"`       // vite, webpack, etc.
	Language       string                 `json:"language"`        // javascript, typescript
	PackageManager string                 `json:"package_manager"` // npm, yarn, pnpm
	BuildCommand   string                 `json:"build_command"`
	StartCommand   string                 `json:"start_command"`
	DevCommand     string                 `json:"dev_command"`
	InstallCommand string                 `json:"install_command"`
	Port           int                    `json:"port"`
	Environment    map[string]interface{} `json:"environment"`
	HasDocker      bool                   `json:"has_docker"`
	DockerCommand  string                 `json:"docker_command"`
	Files          []string               `json:"files"` // Important files found
}

// ListGitHubRepositories returns all GitHub repositories for a project
func (h *GitHubHandler) ListGitHubRepositories(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var repositories []models.GitHubRepository
	if err := h.db.Where("project_id = ?", uint(projectID)).Preload("SSHKey").Find(&repositories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repositories"})
		return
	}

	c.JSON(http.StatusOK, repositories)
}

// CreateGitHubRepository adds a new GitHub repository
func (h *GitHubHandler) CreateGitHubRepository(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateGitHubRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate SSH key if provided
	if req.SSHKeyID != nil {
		var sshKey models.SSHKey
		if err := h.db.Where("id = ? AND project_id = ?", *req.SSHKeyID, uint(projectID)).First(&sshKey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SSH key not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify SSH key"})
			}
			return
		}
	}

	// Set defaults
	if req.Branch == "" {
		req.Branch = "main"
	}
	if req.AppPort == 0 {
		req.AppPort = 3000
	}
	if req.BuildCommand == "" {
		req.BuildCommand = "npm run build"
	}
	if req.StartCommand == "" {
		req.StartCommand = "npm start"
	}

	// Generate webhook secret
	webhookSecret, err := h.generateWebhookSecret()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate webhook secret"})
		return
	}

	// Create GitHub repository record
	repository := models.GitHubRepository{
		Name:          req.Name,
		FullName:      req.FullName,
		CloneURL:      req.CloneURL,
		Branch:        req.Branch,
		IsPrivate:     req.IsPrivate,
		Description:   req.Description,
		WebhookSecret: webhookSecret,
		SSHKeyID:      req.SSHKeyID,
		SDKVersion:    req.SDKVersion,
		AppPort:       req.AppPort,
		BuildCommand:  req.BuildCommand,
		StartCommand:  req.StartCommand,
		Environment:   req.Environment,
		ProjectID:     uint(projectID),
		IsActive:      true,
	}

	if err := h.db.Create(&repository).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create GitHub repository"})
		return
	}

	// Load with SSH key for response
	if err := h.db.Preload("SSHKey").First(&repository, repository.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load repository details"})
		return
	}

	c.JSON(http.StatusCreated, repository)
}

// GetGitHubRepository returns a specific GitHub repository
func (h *GitHubHandler) GetGitHubRepository(c *gin.Context) {
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

	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).Preload("SSHKey").First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repository"})
		}
		return
	}

	c.JSON(http.StatusOK, repository)
}

// UpdateGitHubRepository updates a GitHub repository
func (h *GitHubHandler) UpdateGitHubRepository(c *gin.Context) {
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

	var req UpdateGitHubRepositoryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repository"})
		}
		return
	}

	// Validate SSH key if provided
	if req.SSHKeyID != nil {
		var sshKey models.SSHKey
		if err := h.db.Where("id = ? AND project_id = ?", *req.SSHKeyID, uint(projectID)).First(&sshKey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SSH key not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify SSH key"})
			}
			return
		}
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Branch != "" {
		updates["branch"] = req.Branch
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.SDKVersion != "" {
		updates["sdk_version"] = req.SDKVersion
	}
	if req.AppPort != nil {
		updates["app_port"] = *req.AppPort
	}
	if req.BuildCommand != "" {
		updates["build_command"] = req.BuildCommand
	}
	if req.StartCommand != "" {
		updates["start_command"] = req.StartCommand
	}
	if req.Environment != nil {
		updates["environment"] = req.Environment
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	if req.SSHKeyID != nil {
		updates["ssh_key_id"] = *req.SSHKeyID
	}

	if err := h.db.Model(&repository).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update GitHub repository"})
		return
	}

	// Reload repository with SSH key
	if err := h.db.Preload("SSHKey").First(&repository, repository.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated repository"})
		return
	}

	c.JSON(http.StatusOK, repository)
}

// DeleteGitHubRepository deletes a GitHub repository
func (h *GitHubHandler) DeleteGitHubRepository(c *gin.Context) {
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

	// Check if repository is being used by any deployments
	var deploymentCount int64
	if err := h.db.Model(&models.Deployment{}).Where("github_repository_id = ?", uint(repoID)).Count(&deploymentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check repository usage"})
		return
	}

	if deploymentCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete repository that has active deployments"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).Delete(&models.GitHubRepository{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete GitHub repository"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "GitHub repository deleted successfully"})
}

// SyncRepository syncs repository information with GitHub
func (h *GitHubHandler) SyncRepository(c *gin.Context) {
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

	// Find repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repository"})
		}
		return
	}

	// Get access token from request header (optional for public repos)
	accessToken := c.GetHeader("X-GitHub-Token")

	// Sync with GitHub API
	if err := h.githubService.SyncRepository(&repository, accessToken); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to sync with GitHub API",
			"details": err.Error(),
		})
		return
	}

	// Reload repository to get updated data
	if err := h.db.First(&repository, repository.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated repository"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":    "Repository synced successfully",
		"repository": repository,
	})
}

// GetWebhookInfo returns webhook configuration information
func (h *GitHubHandler) GetWebhookInfo(c *gin.Context) {
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

	// Find repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repository"})
		}
		return
	}

	// Generate webhook URL
	webhookURL := fmt.Sprintf("%s/api/v1/deploy/webhook/%d", h.cfg.BaseURL, repository.ID)

	c.JSON(http.StatusOK, gin.H{
		"webhook_url":    webhookURL,
		"webhook_secret": repository.WebhookSecret,
		"events": []string{
			"push",
			"pull_request",
		},
		"content_type": "application/json",
	})
}

// ValidateRepository validates if a GitHub repository exists and is accessible
func (h *GitHubHandler) ValidateRepository(c *gin.Context) {
	var req struct {
		FullName string `json:"full_name" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get access token from request header (optional for public repos)
	accessToken := c.GetHeader("X-GitHub-Token")

	// Validate repository with GitHub API
	repoInfo, err := h.githubService.ValidateRepository(req.FullName, accessToken)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Repository validation failed",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"valid": true,
		"repository": repoInfo,
	})
}

// SearchRepositories searches for GitHub repositories
func (h *GitHubHandler) SearchRepositories(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	// Get access token from request header (optional)
	accessToken := c.GetHeader("X-GitHub-Token")

	// Search repositories
	repositories, err := h.githubService.SearchRepositories(query, accessToken, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to search repositories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"count": len(repositories),
	})
}

// GetRepositoryBranches returns all branches for a repository
func (h *GitHubHandler) GetRepositoryBranches(c *gin.Context) {
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

	// Find repository
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "GitHub repository not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch GitHub repository"})
		}
		return
	}

	// Get access token from request header (optional for public repos)
	accessToken := c.GetHeader("X-GitHub-Token")

	// Get branches from GitHub API
	branches, err := h.githubService.GetRepositoryBranches(&repository, accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch branches",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"branches": branches,
		"count": len(branches),
	})
}

// GetUserRepositories returns repositories for the authenticated user
func (h *GitHubHandler) GetUserRepositories(c *gin.Context) {
	// Get access token from request header (required for this endpoint)
	accessToken := c.GetHeader("X-GitHub-Token")
	if accessToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "GitHub access token required"})
		return
	}

	visibility := c.DefaultQuery("visibility", "all") // all, public, private
	limitStr := c.DefaultQuery("limit", "30")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 30
	}

	// Get user repositories
	repositories, err := h.githubService.GetUserRepositories(accessToken, visibility, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch user repositories",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"repositories": repositories,
		"count": len(repositories),
	})
}

// AnalyzeRepository analyzes a repository to detect project type and suggest build commands
func (h *GitHubHandler) AnalyzeRepository(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req ProjectAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default branch if not provided
	if req.Branch == "" {
		req.Branch = "main"
	}

	// Analyze repository
	analysis, err := h.analyzeRepository(uint(projectID), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze repository",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, analysis)
}

// analyzeRepository performs the actual repository analysis
func (h *GitHubHandler) analyzeRepository(projectID uint, req ProjectAnalysisRequest) (*ProjectAnalysisResponse, error) {
	// For now, we'll create a basic analysis based on common patterns
	// In a full implementation, this would clone/fetch the repository and analyze files
	
	analysis := &ProjectAnalysisResponse{
		ProjectType:    "unknown",
		Framework:      "unknown",
		Language:       "javascript",
		PackageManager: "npm",
		Port:           3000,
		Environment:    make(map[string]interface{}),
		Files:          []string{},
	}

	// Basic analysis based on repository URL and name patterns
	repoName := extractRepoName(req.RepoURL)
	
	// Detect common project types based on repository name
	switch {
	case containsIgnoreCase(repoName, "photoportfolio") || containsIgnoreCase(repoName, "portfolio"):
		analysis.ProjectType = "photoportfolio"
		analysis.Framework = "vite"
		analysis.Language = "typescript"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm run preview"
		analysis.DevCommand = "npm run dev"
		analysis.InstallCommand = "npm install"
		analysis.Port = 5173
		analysis.Environment["VITE_API_URL"] = "http://localhost:8080/p/project-slug/api"
		analysis.Files = []string{"package.json", "vite.config.ts", "src/main.tsx"}
		
	case containsIgnoreCase(repoName, "react") || containsIgnoreCase(repoName, "cra"):
		analysis.ProjectType = "react"
		analysis.Framework = "create-react-app"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm start"
		analysis.DevCommand = "npm start"
		analysis.InstallCommand = "npm install"
		analysis.Port = 3000
		
	case containsIgnoreCase(repoName, "next"):
		analysis.ProjectType = "nextjs"
		analysis.Framework = "nextjs"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm start"
		analysis.DevCommand = "npm run dev"
		analysis.InstallCommand = "npm install"
		analysis.Port = 3000
		
	case containsIgnoreCase(repoName, "vue"):
		analysis.ProjectType = "vue"
		analysis.Framework = "vite"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm run preview"
		analysis.DevCommand = "npm run dev"  
		analysis.InstallCommand = "npm install"
		analysis.Port = 5173
		
	case containsIgnoreCase(repoName, "nuxt"):
		analysis.ProjectType = "nuxt"
		analysis.Framework = "nuxt"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm start"
		analysis.DevCommand = "npm run dev"
		analysis.InstallCommand = "npm install"
		analysis.Port = 3000
		
	case containsIgnoreCase(repoName, "angular"):
		analysis.ProjectType = "angular"
		analysis.Framework = "angular-cli"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm start"
		analysis.DevCommand = "ng serve"
		analysis.InstallCommand = "npm install"
		analysis.Port = 4200
		
	case containsIgnoreCase(repoName, "svelte"):
		analysis.ProjectType = "svelte"
		analysis.Framework = "vite"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm run preview"
		analysis.DevCommand = "npm run dev"
		analysis.InstallCommand = "npm install"
		analysis.Port = 5173
		
	default:
		// Generic Node.js project
		analysis.ProjectType = "nodejs"
		analysis.Framework = "generic"
		analysis.BuildCommand = "npm run build"
		analysis.StartCommand = "npm start"
		analysis.DevCommand = "npm run dev"
		analysis.InstallCommand = "npm install"
		analysis.Port = 3000
	}

	// Check for Docker
	analysis.HasDocker = false // Would check for Dockerfile in real implementation
	if analysis.HasDocker {
		analysis.DockerCommand = "docker build -t " + repoName + " ."
	}

	return analysis, nil
}

// extractRepoName extracts repository name from URL
func extractRepoName(repoURL string) string {
	// Extract repo name from various URL formats
	// https://github.com/user/repo.git -> repo
	// git@github.com:user/repo.git -> repo
	
	if repoURL == "" {
		return ""
	}
	
	// Remove .git suffix
	if len(repoURL) > 4 && repoURL[len(repoURL)-4:] == ".git" {
		repoURL = repoURL[:len(repoURL)-4]
	}
	
	// Split by / and get last part
	parts := strings.Split(repoURL, "/")
	if len(parts) > 0 {
		return strings.ToLower(parts[len(parts)-1])
	}
	
	return ""
}

// containsIgnoreCase checks if a string contains a substring (case insensitive)
func containsIgnoreCase(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}

// generateWebhookSecret generates a random webhook secret
func (h *GitHubHandler) generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}