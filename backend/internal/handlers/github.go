package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	ProjectType       string                 `json:"project_type"`       // react, vue, angular, next, nuxt, etc.
	Framework         string                 `json:"framework"`          // vite, webpack, etc.
	Language          string                 `json:"language"`           // javascript, typescript
	PackageManager    string                 `json:"package_manager"`    // npm, yarn, pnpm
	BuildCommand      string                 `json:"build_command"`
	StartCommand      string                 `json:"start_command"`
	DevCommand        string                 `json:"dev_command"`
	InstallCommand    string                 `json:"install_command"`
	Port              int                    `json:"port"`
	Environment       map[string]interface{} `json:"environment"`
	HasDocker         bool                   `json:"has_docker"`
	DockerCommand     string                 `json:"docker_command"`
	HasScripts        bool                   `json:"has_scripts"`        // Install/deployment scripts available
	InstallScript     string                 `json:"install_script"`     // Path to install script (if available)
	DeployScript      string                 `json:"deploy_script"`      // Path to deploy script (if available)
	HasDockerCompose  bool                   `json:"has_docker_compose"` // Docker Compose + Dockerfile available
	Files             []string               `json:"files"`              // Important files found
	FileContents      map[string]string      `json:"-"`                  // File contents for analysis (not exposed in JSON)
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

	// Use repository's stored access token for GitHub API calls
	accessToken := repository.AccessToken

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

	// Use repository's stored access token for GitHub API calls
	accessToken := repository.AccessToken

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

	// For this endpoint, we analyze without a specific repository ID
	// This is used for analyzing repositories before adding them
	// Get token from header for temporary analysis
	accessToken := c.GetHeader("X-GitHub-Token")
	if accessToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-GitHub-Token header required for repository analysis"})
		return
	}
	
	// Create temporary repository object with the token
	tempRepo := &models.GitHubRepository{
		AccessToken: accessToken,
	}
	
	analysis, err := h.analyzeRepository(uint(projectID), req, tempRepo)
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
func (h *GitHubHandler) analyzeRepository(projectID uint, req ProjectAnalysisRequest, repository *models.GitHubRepository) (*ProjectAnalysisResponse, error) {
	analysis := &ProjectAnalysisResponse{
		ProjectType:    "unknown",
		Framework:      "unknown",
		Language:       "javascript",
		PackageManager: "npm",
		Port:           3000,
		Environment:    make(map[string]interface{}),
		Files:          []string{},
	}

	// Try to analyze repository contents via GitHub API
	repoName := extractRepoName(req.RepoURL)
	
	// If we have actual repository access, analyze the contents
	if err := h.analyzeRepositoryContents(analysis, req, repository); err != nil {
		// Fall back to basic analysis based on repository name patterns
		h.analyzeByName(analysis, repoName)
	}

	// Check for Docker
	analysis.HasDocker = false // Would check for Dockerfile in real implementation
	if analysis.HasDocker {
		analysis.DockerCommand = "docker build -t " + repoName + " ."
	}

	return analysis, nil
}

// analyzeAndSaveRepository performs analysis and saves to database
func (h *GitHubHandler) analyzeAndSaveRepository(projectID uint, repoID uint, req ProjectAnalysisRequest, repository *models.GitHubRepository) (*models.RepositoryAnalysis, error) {
	// Perform the analysis
	analysisResp, err := h.analyzeRepository(projectID, req, repository)
	if err != nil {
		return nil, err
	}

	// Generate install options
	installOptions := h.generateInstallOptions(analysisResp)

	// Create repository analysis record
	repoAnalysis := &models.RepositoryAnalysis{
		GitHubRepositoryID: repoID,
		ProjectID:          projectID,
		AnalyzedAt:         time.Now(),
		AnalyzedBranch:     req.Branch,
		AnalysisStatus:     "completed",
		
		// Project detection
		ProjectType:    analysisResp.ProjectType,
		Framework:      analysisResp.Framework,
		Language:       analysisResp.Language,
		PackageManager: analysisResp.PackageManager,
		
		// Build configuration
		BuildCommand:   analysisResp.BuildCommand,
		StartCommand:   analysisResp.StartCommand,
		DevCommand:     analysisResp.DevCommand,
		InstallCommand: analysisResp.InstallCommand,
		TestCommand:    "", // Will be detected later
		
		// Runtime configuration
		Port:        analysisResp.Port,
		Environment: analysisResp.Environment,
		
		// Docker support
		HasDocker:     analysisResp.HasDocker,
		DockerCommand: analysisResp.DockerCommand,
		DockerPort:    0, // Will be detected later
		
		// Dependencies and features
		Dependencies:    []string{}, // Will be populated from package.json analysis
		DevDependencies: []string{}, // Will be populated from package.json analysis
		Scripts:         []string{}, // Will be populated from package.json analysis
		
		// File structure
		ImportantFiles:   analysisResp.Files,
		ConfigFiles:      []string{}, // Will be detected later
		EnvironmentFiles: []string{}, // Will be detected later
		
		// Installation options
		InstallOptions: installOptions,
		
		// Analysis insights
		Insights:     h.generateInsights(analysisResp),
		Warnings:     []string{}, // Will be populated later
		Requirements: []string{}, // Will be populated later
		
		// Performance metrics
		EstimatedBuildTime: 0,    // Will be calculated later
		EstimatedSize:      0,    // Will be calculated later
		Complexity:         h.calculateComplexity(analysisResp),
		
		// Analysis errors
		AnalysisErrors: []string{}, // Will store any errors encountered
	}

	// Save or update the analysis
	var existingAnalysis models.RepositoryAnalysis
	result := h.db.Where("github_repository_id = ?", repoID).First(&existingAnalysis)
	
	if result.Error == gorm.ErrRecordNotFound {
		// Create new analysis
		if err := h.db.Create(repoAnalysis).Error; err != nil {
			return nil, err
		}
	} else if result.Error != nil {
		return nil, result.Error
	} else {
		// Update existing analysis
		repoAnalysis.ID = existingAnalysis.ID
		if err := h.db.Save(repoAnalysis).Error; err != nil {
			return nil, err
		}
	}

	return repoAnalysis, nil
}

// analyzeRepositoryContents analyzes repository contents via GitHub API
func (h *GitHubHandler) analyzeRepositoryContents(analysis *ProjectAnalysisResponse, req ProjectAnalysisRequest, repository *models.GitHubRepository) error {
	// Extract owner and repo from URL
	owner, repo, err := h.parseGitHubURL(req.RepoURL)
	if err != nil {
		return err
	}

	// Try to fetch key files to analyze project structure
	files := []string{"package.json", "Dockerfile", "docker-compose.yml", "vite.config.ts", "vite.config.js", 
					  "next.config.js", "nuxt.config.js", "angular.json", "svelte.config.js", ".env.example", "README.md",
					  "scripts/install.sh", "scripts/deploy.sh", "scripts/setup.sh", "install.sh", "deploy.sh"}
	
	fileContents := make(map[string]string)
	
	for _, file := range files {
		content, err := h.fetchFileContentWithToken(repository.AccessToken, owner, repo, file, req.Branch)
		if err == nil && content != "" {
			fileContents[file] = content
			analysis.Files = append(analysis.Files, file)
		}
	}
	
	// Store file contents in analysis for later use in port analysis
	analysis.FileContents = fileContents

	// Analyze package.json if it exists
	if packageJSON, exists := fileContents["package.json"]; exists {
		h.analyzePackageJSON(analysis, packageJSON)
	}

	// Check for deployment options (prioritized order)
	// 1. Scripts folder (highest priority)
	analysis.HasScripts = h.hasInstallScripts(fileContents)
	if analysis.HasScripts {
		// Get the actual script paths found in the repository
		analysis.InstallScript, analysis.DeployScript = h.getAvailableScripts(fileContents)
		// Scripts take precedence - set Docker to false
		analysis.HasDocker = false
		if analysis.InstallScript != "" {
			analysis.DockerCommand = "bash " + analysis.InstallScript
		}
	}
	
	analysis.HasDockerCompose = h.hasDockerCompose(fileContents)
	
	if analysis.HasDockerCompose && !analysis.HasScripts {
		// 2. Docker Compose + Dockerfile (second priority)
		analysis.HasDocker = true
		analysis.DockerCommand = "docker-compose up -d"
	} else if _, hasDockerfile := fileContents["Dockerfile"]; hasDockerfile {
		// 3. Only Dockerfile (third priority)
		analysis.HasDocker = true
		analysis.DockerCommand = "docker build -t " + repo + " ."
	}

	// Analyze environment variables from .env.example
	if envExample, exists := fileContents[".env.example"]; exists {
		h.analyzeEnvironmentVariables(analysis, envExample)
	}

	// Analyze README for additional insights
	if readme, exists := fileContents["README.md"]; exists {
		h.analyzeReadme(analysis, readme)
	}

	return nil
}

// hasInstallScripts checks if repository has install/deployment scripts
func (h *GitHubHandler) hasInstallScripts(fileContents map[string]string) bool {
	scriptFiles := []string{"scripts/install.sh", "scripts/deploy.sh", "scripts/setup.sh", "install.sh", "deploy.sh"}
	for _, script := range scriptFiles {
		if _, exists := fileContents[script]; exists {
			return true
		}
	}
	return false
}

// getAvailableScripts returns the actual script paths found in the repository
func (h *GitHubHandler) getAvailableScripts(fileContents map[string]string) (installScript, deployScript string) {
	// Priority order for install scripts
	installScripts := []string{"scripts/install.sh", "install.sh", "scripts/setup.sh"}
	deployScripts := []string{"scripts/deploy.sh", "deploy.sh", "scripts/start.sh", "start.sh"}
	
	for _, script := range installScripts {
		if _, exists := fileContents[script]; exists {
			installScript = script
			break
		}
	}
	
	for _, script := range deployScripts {
		if _, exists := fileContents[script]; exists {
			deployScript = script
			break
		}
	}
	
	return installScript, deployScript
}

// analyzeScriptPortRequirements analyzes script files for port requirements
func (h *GitHubHandler) analyzeScriptPortRequirements(fileContents map[string]string, scriptPath string) []models.PortRequirement {
	var requirements []models.PortRequirement
	
	if scriptPath == "" {
		return requirements
	}
	
	scriptContent, exists := fileContents[scriptPath]
	if !exists {
		return requirements
	}
	
	// Simplified port patterns focused on essential ports only
	// Prioritize single main application port to avoid duplicates
	portPatterns := map[string]models.PortRequirement{
		// Main application port (highest priority)
		"PORT": {Name: "Application", Port: 3000, Protocol: "tcp", Required: true, Description: "Main application port", Variable: "PORT"},
		
		// Alternative web server patterns (only if PORT is not found)
		"WEB_PORT": {Name: "Web Server", Port: 8080, Protocol: "tcp", Required: true, Description: "Web server port", Variable: "WEB_PORT"},
		"HTTP_PORT": {Name: "HTTP Server", Port: 8080, Protocol: "tcp", Required: true, Description: "HTTP server port", Variable: "HTTP_PORT"},
		
		// API patterns (only for explicit API services)
		"API_PORT": {Name: "API Server", Port: 4000, Protocol: "tcp", Required: false, Description: "API server port", Variable: "API_PORT"},
		
		// Database patterns (only if explicitly referenced in scripts)
		"DB_PORT": {Name: "Database", Port: 5432, Protocol: "tcp", Required: false, Description: "External database port", Variable: "DB_PORT"},
		"MONGODB_PORT": {Name: "MongoDB", Port: 27017, Protocol: "tcp", Required: false, Description: "MongoDB port", Variable: "MONGODB_PORT"},
	}
	
	lines := strings.Split(scriptContent, "\n")
	foundPorts := make(map[string]bool)
	hasMainPort := false // Track if we found the main application port
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		
		// Skip comments and empty lines
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		// Look for port variable assignments and usage with priority logic
		for variable, template := range portPatterns {
			if foundPorts[variable] {
				continue // Already found this port
			}
			
			// Check for various patterns where ports might be defined or used
			patterns := []string{
				variable + "=",           // WEB_PORT=80
				variable + " =",          // WEB_PORT = 80  
				"${" + variable + "}",    // ${WEB_PORT}
				"$" + variable,           // $WEB_PORT
				":" + variable,           // Port: WEB_PORT
				variable + ":",           // WEB_PORT: 80
			}
			
			for _, pattern := range patterns {
				if strings.Contains(strings.ToUpper(line), strings.ToUpper(pattern)) {
					// Priority logic: avoid duplicate web server ports
					if variable == "PORT" {
						hasMainPort = true
					} else if (variable == "WEB_PORT" || variable == "HTTP_PORT") && hasMainPort {
						// Skip alternative web server ports if main PORT already found
						break
					}
					
					// Found a port reference, try to extract the default value
					port := h.extractPortFromLine(line, variable)
					if port > 0 {
						requirement := template
						requirement.Port = port
						requirements = append(requirements, requirement)
						foundPorts[variable] = true
						
						if variable == "PORT" {
							hasMainPort = true
						}
						break
					} else {
						// Found reference but no default port, use template defaults
						requirements = append(requirements, template)
						foundPorts[variable] = true
						
						if variable == "PORT" {
							hasMainPort = true
						}
						break
					}
				}
			}
		}
	}
	
	// Filter out unnecessary ports for portfolio apps
	if h.isPortfolioApp(scriptContent) {
		filteredRequirements := []models.PortRequirement{}
		for _, req := range requirements {
			// Portfolio apps only need frontend webserver port
			// API is provided by CloudBox itself, no separate API server needed
			if req.Variable == "PORT" || req.Variable == "WEB_PORT" || req.Variable == "HTTP_PORT" {
				filteredRequirements = append(filteredRequirements, req)
			}
			// Skip API_PORT - CloudBox provides the API
			// Skip database ports - portfolio apps typically use CloudBox data API
		}
		return filteredRequirements
	}
	
	return requirements
}

// isPortfolioApp checks if this appears to be a simple portfolio/static site app
func (h *GitHubHandler) isPortfolioApp(scriptContent string) bool {
	portfolioIndicators := []string{
		"portfolio", "photoportfolio", "static", "vite", "npm run preview", 
		"serve", "http-server", "live-server",
	}
	
	scriptLower := strings.ToLower(scriptContent)
	for _, indicator := range portfolioIndicators {
		if strings.Contains(scriptLower, indicator) {
			return true
		}
	}
	
	return false
}

// extractPortFromLine tries to extract a port number from a script line
func (h *GitHubHandler) extractPortFromLine(line, variable string) int {
	// Remove comments
	if commentIndex := strings.Index(line, "#"); commentIndex >= 0 {
		line = line[:commentIndex]
	}
	
	// Look for patterns like WEB_PORT=80, WEB_PORT: 80, etc.
	patterns := []string{
		variable + "=",
		variable + " =",
		variable + ":",
		variable + " :",
	}
	
	for _, pattern := range patterns {
		if index := strings.Index(strings.ToUpper(line), strings.ToUpper(pattern)); index >= 0 {
			// Extract everything after the pattern
			remaining := strings.TrimSpace(line[index+len(pattern):])
			
			// Extract the first number found
			var numStr strings.Builder
			for _, char := range remaining {
				if char >= '0' && char <= '9' {
					numStr.WriteRune(char)
				} else if numStr.Len() > 0 {
					break // Stop at first non-digit after finding digits
				}
			}
			
			if numStr.Len() > 0 {
				if port, err := strconv.Atoi(numStr.String()); err == nil && port > 0 && port <= 65535 {
					return port
				}
			}
		}
	}
	
	return 0
}

// getDefaultPortForVariable returns sensible defaults for common port variables
func (h *GitHubHandler) getDefaultPortForVariable(variable string) int {
	defaults := map[string]int{
		"WEB_PORT":     80,
		"HTTP_PORT":    80,
		"PORT":         3000,
		"API_PORT":     4000,
		"SERVER_PORT":  8000,
		"DB_PORT":      5432,
		"MONGO_PORT":   27017,
		"MONGODB_PORT": 27017,
		"MYSQL_PORT":   3306,
		"POSTGRES_PORT": 5432,
		"REDIS_PORT":   6379,
		"DEV_PORT":     3000,
		"PREVIEW_PORT": 4173,
	}
	
	return defaults[variable]
}

// analyzeScriptPortRequirementsFromMap is a wrapper that accepts the file contents map
func (h *GitHubHandler) analyzeScriptPortRequirementsFromMap(fileContents map[string]string, scriptPath string) []models.PortRequirement {
	return h.analyzeScriptPortRequirements(fileContents, scriptPath)
}

// mergePortRequirements combines port requirements from multiple sources, avoiding duplicates
func (h *GitHubHandler) mergePortRequirements(existing, additional []models.PortRequirement) []models.PortRequirement {
	existingVars := make(map[string]bool)
	result := make([]models.PortRequirement, len(existing))
	copy(result, existing)
	
	// Track existing variables
	for _, req := range existing {
		existingVars[req.Variable] = true
	}
	
	// Add new requirements that don't conflict
	for _, req := range additional {
		if !existingVars[req.Variable] {
			result = append(result, req)
			existingVars[req.Variable] = true
		}
	}
	
	return result
}

// hasDockerCompose checks if repository has both docker-compose.yml and Dockerfile
func (h *GitHubHandler) hasDockerCompose(fileContents map[string]string) bool {
	_, hasCompose := fileContents["docker-compose.yml"]
	_, hasDockerfile := fileContents["Dockerfile"]
	return hasCompose && hasDockerfile
}

// analyzeByName performs basic analysis based on repository name
func (h *GitHubHandler) analyzeByName(analysis *ProjectAnalysisResponse, repoName string) {
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

// parseGitHubURL extracts owner and repo from GitHub URL
func (h *GitHubHandler) parseGitHubURL(repoURL string) (owner, repo string, err error) {
	// Handle various GitHub URL formats:
	// https://github.com/owner/repo
	// https://github.com/owner/repo.git
	// git@github.com:owner/repo.git
	
	// Remove .git suffix
	if strings.HasSuffix(repoURL, ".git") {
		repoURL = repoURL[:len(repoURL)-4]
	}
	
	// Handle SSH format
	if strings.HasPrefix(repoURL, "git@github.com:") {
		repoURL = strings.TrimPrefix(repoURL, "git@github.com:")
		parts := strings.Split(repoURL, "/")
		if len(parts) == 2 {
			return parts[0], parts[1], nil
		}
	}
	
	// Handle HTTPS format
	if strings.Contains(repoURL, "github.com/") {
		parts := strings.Split(repoURL, "github.com/")
		if len(parts) > 1 {
			pathParts := strings.Split(parts[1], "/")
			if len(pathParts) >= 2 {
				return pathParts[0], pathParts[1], nil
			}
		}
	}
	
	return "", "", fmt.Errorf("invalid GitHub URL format")
}

// fetchFileContent fetches file content from GitHub repository via GitHub API
func (h *GitHubHandler) fetchFileContent(owner, repo, file, branch string) (string, error) {
	// Try to get repository-specific access token first
	var githubRepo models.GitHubRepository
	if err := h.db.Where("full_name = ?", owner+"/"+repo).First(&githubRepo).Error; err == nil && githubRepo.AccessToken != "" {
		// Use repository-specific token
		return h.fetchFileContentWithToken(githubRepo.AccessToken, owner, repo, file, branch)
	}
	
	// Fall back to global token if available
	globalToken := os.Getenv("GITHUB_TOKEN")
	if globalToken != "" {
		return h.fetchFileContentWithToken(globalToken, owner, repo, file, branch)
	}
	
	return "", fmt.Errorf("No GitHub access token available for repository %s/%s. Please authorize repository access.", owner, repo)
}

// fetchFileContentWithToken fetches file content using a specific GitHub token
func (h *GitHubHandler) fetchFileContentWithToken(githubToken, owner, repo, file, branch string) (string, error) {

	// Construct GitHub API URL
	url := fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", owner, repo, file, branch)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	
	// Add authorization header
	req.Header.Set("Authorization", "token "+githubToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")
	
	// Make HTTP request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch file: %w", err)
	}
	defer resp.Body.Close()
	
	// Handle 404 (file not found)
	if resp.StatusCode == 404 {
		return "", fmt.Errorf("file not found: %s", file)
	}
	
	// Handle other errors
	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("GitHub API error %d: %s", resp.StatusCode, string(body))
	}
	
	// Parse response
	var response struct {
		Content  string `json:"content"`
		Encoding string `json:"encoding"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	
	// Decode base64 content
	if response.Encoding == "base64" {
		decoded, err := base64.StdEncoding.DecodeString(response.Content)
		if err != nil {
			return "", fmt.Errorf("failed to decode base64 content: %w", err)
		}
		return string(decoded), nil
	}
	
	return response.Content, nil
}

// GitHubAuthorizeRepository initiates OAuth flow for repository access
func (h *GitHubHandler) GitHubAuthorizeRepository(c *gin.Context) {
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

	// Get repository to validate ownership
	var repo models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", repoID, projectID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Get project-specific OAuth configuration
	var gitHubConfig models.ProjectGitHubConfig
	if err := h.db.Where("project_id = ?", projectID).First(&gitHubConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured for this project"})
		return
	}
	
	if !gitHubConfig.IsEnabled || gitHubConfig.ClientID == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not enabled or configured for this project"})
		return
	}

	// Generate state parameter for security
	state := fmt.Sprintf("%d_%d_%s", projectID, repoID, generateRandomString(16))
	
	// Build GitHub OAuth URL
	scope := "repo" // Full repository access
	if !repo.IsPrivate {
		scope = "public_repo" // Only public repositories
	}
	
	// Use project-specific callback URL
	callbackURL := gitHubConfig.CallbackURL
	
	authURL := fmt.Sprintf(
		"https://github.com/login/oauth/authorize?client_id=%s&scope=%s&state=%s&redirect_uri=%s", 
		gitHubConfig.ClientID,
		scope,
		state,
		callbackURL,
	)

	c.JSON(http.StatusOK, gin.H{
		"auth_url": authURL,
		"state":    state,
	})
}

// GitHubOAuthCallback handles the OAuth callback from GitHub
func (h *GitHubHandler) GitHubOAuthCallback(c *gin.Context) {
	code := c.Query("code")
	state := c.Query("state")
	
	if code == "" || state == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Missing code or state parameter"})
		return
	}

	// Parse state to get project and repo IDs
	parts := strings.Split(state, "_")
	if len(parts) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid state parameter"})
		return
	}

	projectID, err := strconv.ParseUint(parts[0], 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID in state"})
		return
	}

	repoID, err := strconv.ParseUint(parts[1], 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID in state"})
		return
	}

	// Get project-specific OAuth configuration for token exchange
	var gitHubConfig models.ProjectGitHubConfig
	if err := h.db.Where("project_id = ?", projectID).First(&gitHubConfig).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub OAuth not configured for this project"})
		return
	}

	// Exchange code for access token using project-specific credentials
	token, err := h.exchangeCodeForTokenWithProjectConfig(code, gitHubConfig.ClientID, gitHubConfig.ClientSecret)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to exchange code for token: " + err.Error()})
		return
	}

	// Get GitHub user info to verify
	userInfo, err := h.getGitHubUserInfo(token.AccessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info: " + err.Error()})
		return
	}

	// Update repository with OAuth info
	now := time.Now()
	updates := map[string]interface{}{
		"access_token":    token.AccessToken,
		"refresh_token":   token.RefreshToken,
		"token_scopes":    token.Scope,
		"authorized_at":   &now,
		"authorized_by":   userInfo.Login,
	}

	if token.ExpiresIn > 0 {
		expiresAt := now.Add(time.Duration(token.ExpiresIn) * time.Second)
		updates["token_expires_at"] = &expiresAt
	}

	if err := h.db.Model(&models.GitHubRepository{}).Where("id = ? AND project_id = ?", repoID, projectID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save OAuth token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Repository authorized successfully",
		"authorized_by": userInfo.Login,
		"scopes":        token.Scope,
	})
}

// UpdateRepositoryToken updates the Personal Access Token for a repository
func (h *GitHubHandler) UpdateRepositoryToken(c *gin.Context) {
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

	var req struct {
		AccessToken string `json:"access_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get repository
	var repo models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", repoID, projectID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Validate token by testing GitHub API access
	userInfo, err := h.getGitHubUserInfo(req.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Personal Access Token: " + err.Error()})
		return
	}

	// Update repository with PAT info
	now := time.Now()
	updates := map[string]interface{}{
		"access_token":  req.AccessToken,
		"authorized_at": &now,
		"authorized_by": userInfo.Login,
		"token_scopes":  "fine-grained-pat", // PAT identifier
	}

	if err := h.db.Model(&models.GitHubRepository{}).Where("id = ? AND project_id = ?", repoID, projectID).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save Personal Access Token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "Personal Access Token saved successfully",
		"authorized_by": userInfo.Login,
	})
}

// TestRepositoryAccess tests if we can access the repository with current authorization
func (h *GitHubHandler) TestRepositoryAccess(c *gin.Context) {
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

	// Get repository
	var repo models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", repoID, projectID).First(&repo).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Test access by trying to fetch package.json
	owner, repoName, err := h.parseGitHubURL("https://github.com/" + repo.FullName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository URL"})
		return
	}

	// Use the repository's stored access token directly if available
	var content string
	if repo.AccessToken != "" {
		content, err = h.fetchFileContentWithToken(repo.AccessToken, owner, repoName, "package.json", repo.Branch)
	} else {
		content, err = h.fetchFileContent(owner, repoName, "package.json", repo.Branch)
	}
	
	testResult := map[string]interface{}{
		"repository":     repo.FullName,
		"branch":         repo.Branch,
		"has_auth":       repo.AccessToken != "",
		"authorized_by":  repo.AuthorizedBy,
		"authorized_at":  repo.AuthorizedAt,
		"token_scopes":   repo.TokenScopes,
	}

	if err != nil {
		testResult["access_test"] = "failed"
		testResult["error"] = err.Error()
		testResult["needs_auth"] = repo.AccessToken == ""
		c.JSON(http.StatusOK, testResult)
		return
	}

	testResult["access_test"] = "success"
	testResult["sample_file"] = "package.json found"
	testResult["file_size"] = len(content)

	c.JSON(http.StatusOK, testResult)
}

// Helper functions for OAuth flow

type GitHubOAuthToken struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
}

type GitHubUser struct {
	Login string `json:"login"`
	ID    int    `json:"id"`
	Name  string `json:"name"`
}

func (h *GitHubHandler) exchangeCodeForToken(code string) (*GitHubOAuthToken, error) {
	clientID := h.getSystemSetting("github_oauth_client_id")
	clientSecret := h.getSystemSetting("github_oauth_client_secret")
	
	data := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	
	jsonData, _ := json.Marshal(data)
	
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var token GitHubOAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, err
	}
	
	if token.AccessToken == "" {
		return nil, fmt.Errorf("no access token received")
	}
	
	return &token, nil
}

// exchangeCodeForTokenWithProjectConfig exchanges authorization code for access token using project-specific OAuth config
func (h *GitHubHandler) exchangeCodeForTokenWithProjectConfig(code, clientID, clientSecret string) (*GitHubOAuthToken, error) {
	data := map[string]string{
		"client_id":     clientID,
		"client_secret": clientSecret,
		"code":          code,
	}
	
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}
	
	req, err := http.NewRequest("POST", "https://github.com/login/oauth/access_token", strings.NewReader(string(jsonData)))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()
	
	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}
	
	var token GitHubOAuthToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	
	// Check for OAuth errors
	if token.AccessToken == "" {
		return nil, fmt.Errorf("no access token received from GitHub")
	}
	
	return &token, nil
}

func (h *GitHubHandler) getGitHubUserInfo(accessToken string) (*GitHubUser, error) {
	req, err := http.NewRequest("GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Authorization", "token "+accessToken)
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	var user GitHubUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, err
	}
	
	return &user, nil
}

func generateRandomString(length int) string {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(bytes)[:length]
}

// Helper functions for system settings
func (h *GitHubHandler) getSystemSetting(key string) string {
	var setting models.SystemSetting
	if err := h.db.Where("key = ?", key).First(&setting).Error; err != nil {
		return ""
	}
	return setting.Value
}

func (h *GitHubHandler) generateOAuthCallbackURL() string {
	domain := h.getSystemSetting("site_domain")
	protocol := h.getSystemSetting("site_protocol")
	
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
		// For production, assume backend is on same domain with port 8080 or standard ports
		if !strings.Contains(domain, ":") {
			if protocol == "https" {
				domain = domain + ":443"
			} else {
				domain = domain + ":8080"
			}
		}
	}

	return protocol + "://" + domain + "/api/v1/github/oauth/callback"
}

// analyzePackageJSON analyzes package.json to determine project type and dependencies
func (h *GitHubHandler) analyzePackageJSON(analysis *ProjectAnalysisResponse, packageJSON string) {
	// Parse package.json content (basic string matching for now)
	
	// Detect React
	if strings.Contains(packageJSON, "\"react\"") {
		analysis.ProjectType = "react"
		analysis.Language = "javascript"
		if strings.Contains(packageJSON, "\"typescript\"") {
			analysis.Language = "typescript"
		}
	}
	
	// Detect Next.js
	if strings.Contains(packageJSON, "\"next\"") {
		analysis.ProjectType = "nextjs"
		analysis.Framework = "nextjs"
		analysis.Port = 3000
	}
	
	// Detect Vue
	if strings.Contains(packageJSON, "\"vue\"") {
		analysis.ProjectType = "vue"
		analysis.Framework = "vite"
		analysis.Port = 5173
	}
	
	// Detect Nuxt
	if strings.Contains(packageJSON, "\"nuxt\"") {
		analysis.ProjectType = "nuxt"
		analysis.Framework = "nuxt"
		analysis.Port = 3000
	}
	
	// Detect Angular
	if strings.Contains(packageJSON, "\"@angular/core\"") {
		analysis.ProjectType = "angular"
		analysis.Framework = "angular-cli"
		analysis.Port = 4200
	}
	
	// Detect Svelte
	if strings.Contains(packageJSON, "\"svelte\"") {
		analysis.ProjectType = "svelte"
		analysis.Framework = "vite"
		analysis.Port = 5173
	}
	
	// Detect Vite
	if strings.Contains(packageJSON, "\"vite\"") {
		analysis.Framework = "vite"
	}
	
	// Detect package manager
	if strings.Contains(packageJSON, "\"packageManager\": \"yarn") {
		analysis.PackageManager = "yarn"
		analysis.InstallCommand = "yarn install"
	} else if strings.Contains(packageJSON, "\"packageManager\": \"pnpm") {
		analysis.PackageManager = "pnpm"
		analysis.InstallCommand = "pnpm install"
	}
	
	// Extract scripts
	if strings.Contains(packageJSON, "\"build\":") {
		analysis.BuildCommand = "npm run build"
		if analysis.PackageManager == "yarn" {
			analysis.BuildCommand = "yarn build"
		} else if analysis.PackageManager == "pnpm" {
			analysis.BuildCommand = "pnpm build"
		}
	}
	
	if strings.Contains(packageJSON, "\"start\":") {
		analysis.StartCommand = "npm start"
		if analysis.PackageManager == "yarn" {
			analysis.StartCommand = "yarn start"
		} else if analysis.PackageManager == "pnpm" {
			analysis.StartCommand = "pnpm start"
		}
	}
	
	if strings.Contains(packageJSON, "\"dev\":") {
		analysis.DevCommand = "npm run dev"
		if analysis.PackageManager == "yarn" {
			analysis.DevCommand = "yarn dev"
		} else if analysis.PackageManager == "pnpm" {
			analysis.DevCommand = "pnpm dev"
		}
	}
}

// analyzeEnvironmentVariables extracts environment variables from .env.example
func (h *GitHubHandler) analyzeEnvironmentVariables(analysis *ProjectAnalysisResponse, envContent string) {
	lines := strings.Split(envContent, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			
			// Detect common environment variable patterns
			switch {
			case strings.Contains(key, "API_URL") || strings.Contains(key, "BACKEND_URL"):
				analysis.Environment[key] = "http://localhost:8080/p/project-slug/api"
			case strings.Contains(key, "PORT"):
				if portStr := strings.Trim(value, "\"'"); portStr != "" {
					if port, err := strconv.Atoi(portStr); err == nil {
						analysis.Port = port
					}
				}
				analysis.Environment[key] = value
			default:
				analysis.Environment[key] = value
			}
		}
	}
}

// analyzeReadme extracts insights from README file
func (h *GitHubHandler) analyzeReadme(analysis *ProjectAnalysisResponse, readme string) {
	readmeLower := strings.ToLower(readme)
	
	// Look for deployment instructions
	if strings.Contains(readmeLower, "docker") {
		analysis.HasDocker = true
	}
	
	// Look for framework mentions
	if strings.Contains(readmeLower, "next.js") || strings.Contains(readmeLower, "nextjs") {
		analysis.ProjectType = "nextjs"
		analysis.Framework = "nextjs"
	} else if strings.Contains(readmeLower, "vue.js") || strings.Contains(readmeLower, "vuejs") {
		analysis.ProjectType = "vue"
		analysis.Framework = "vite"
	} else if strings.Contains(readmeLower, "react") {
		analysis.ProjectType = "react"
	}
	
	// Look for port mentions
	portPattern := regexp.MustCompile(`(?i)port\s*:?\s*(\d+)`)
	if matches := portPattern.FindStringSubmatch(readme); len(matches) > 1 {
		if port, err := strconv.Atoi(matches[1]); err == nil {
			analysis.Port = port
		}
	}
}

// generateInstallOptions creates different installation options based on analysis
// Priority order: Scripts > Docker Compose > Docker > Package Manager
func (h *GitHubHandler) generateInstallOptions(analysis *ProjectAnalysisResponse) []models.InstallOption {
	var options []models.InstallOption

	// 1. Scripts option (highest priority)
	if analysis.HasScripts && analysis.InstallScript != "" {
		// Use the actual script paths found in the repository
		installCmd := "bash " + analysis.InstallScript
		deployCmd := "bash " + analysis.DeployScript
		
		// Fallback to install script if deploy script not found
		if analysis.DeployScript == "" && analysis.InstallScript != "" {
			deployCmd = "bash " + analysis.InstallScript
		}
		
		// Analyze script for port requirements
		portRequirements := h.analyzeScriptPortRequirementsFromMap(analysis.FileContents, analysis.InstallScript)
		if analysis.DeployScript != "" {
			// Also check deploy script for additional port requirements
			deployPortReqs := h.analyzeScriptPortRequirementsFromMap(analysis.FileContents, analysis.DeployScript)
			portRequirements = h.mergePortRequirements(portRequirements, deployPortReqs)
		}
		
		options = append(options, models.InstallOption{
			Name:         "scripts",
			Command:      installCmd,
			BuildCommand: installCmd,
			StartCommand: deployCmd,
			DevCommand:   deployCmd,
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: true,
			Description:  "Deploy using custom install scripts (recommended - repository provides optimized deployment)",
			PortRequirements: portRequirements,
		})
	}

	// 2. Docker Compose option (second priority)
	if analysis.HasDockerCompose {
		options = append(options, models.InstallOption{
			Name:         "docker-compose",
			Command:      "docker-compose up -d",
			BuildCommand: "docker-compose build",
			StartCommand: "docker-compose up -d",
			DevCommand:   "docker-compose up",
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: !analysis.HasScripts,
			Description:  "Deploy using Docker Compose (full stack deployment with services)",
		})
	}

	// 3. Docker option (third priority)
	if analysis.HasDocker && !analysis.HasDockerCompose {
		options = append(options, models.InstallOption{
			Name:         "docker",
			Command:      "docker build -t app .",
			BuildCommand: "docker build -t app .",
			StartCommand: "docker run -p " + strconv.Itoa(analysis.Port) + ":" + strconv.Itoa(analysis.Port) + " app",
			DevCommand:   "docker run -p " + strconv.Itoa(analysis.Port) + ":" + strconv.Itoa(analysis.Port) + " app",
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: !analysis.HasScripts && !analysis.HasDockerCompose,
			Description:  "Deploy using Docker container",
		})
	}

	// 4. Package manager options (lowest priority)
	hasHigherPriorityOption := analysis.HasScripts || analysis.HasDockerCompose || analysis.HasDocker
	
	switch analysis.PackageManager {
	case "npm":
		options = append(options, models.InstallOption{
			Name:         "npm",
			Command:      "npm install",
			BuildCommand: analysis.BuildCommand,
			StartCommand: analysis.StartCommand,
			DevCommand:   analysis.DevCommand,
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: !hasHigherPriorityOption,
			Description:  "Standard npm installation (manual deployment)",
		})
		
	case "yarn":
		options = append(options, models.InstallOption{
			Name:         "yarn",
			Command:      "yarn install",
			BuildCommand: strings.Replace(analysis.BuildCommand, "npm run", "yarn", 1),
			StartCommand: strings.Replace(analysis.StartCommand, "npm", "yarn", 1),
			DevCommand:   strings.Replace(analysis.DevCommand, "npm run", "yarn", 1),
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: !hasHigherPriorityOption,
			Description:  "Fast, reliable, and secure yarn installation (manual deployment)",
		})
		
		// Also add npm as alternative
		options = append(options, models.InstallOption{
			Name:         "npm",
			Command:      "npm install",
			BuildCommand: analysis.BuildCommand,
			StartCommand: analysis.StartCommand,
			DevCommand:   analysis.DevCommand,
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: false,
			Description:  "Alternative npm installation (manual deployment)",
		})
		
	case "pnpm":
		options = append(options, models.InstallOption{
			Name:         "pnpm",
			Command:      "pnpm install",
			BuildCommand: strings.Replace(analysis.BuildCommand, "npm run", "pnpm", 1),
			StartCommand: strings.Replace(analysis.StartCommand, "npm", "pnpm", 1),
			DevCommand:   strings.Replace(analysis.DevCommand, "npm run", "pnpm", 1),
			Port:         analysis.Port,
			Environment:  analysis.Environment,
			IsRecommended: !hasHigherPriorityOption,
			Description:  "Fast, disk space efficient pnpm installation (manual deployment)",
		})
		
		// Also add npm and yarn as alternatives
		options = append(options, 
			models.InstallOption{
				Name:         "npm",
				Command:      "npm install",
				BuildCommand: analysis.BuildCommand,
				StartCommand: analysis.StartCommand,
				DevCommand:   analysis.DevCommand,
				Port:         analysis.Port,
				Environment:  analysis.Environment,
				IsRecommended: false,
				Description:  "Alternative npm installation (manual deployment)",
			},
			models.InstallOption{
				Name:         "yarn",
				Command:      "yarn install",
				BuildCommand: strings.Replace(analysis.BuildCommand, "npm run", "yarn", 1),
				StartCommand: strings.Replace(analysis.StartCommand, "npm", "yarn", 1),
				DevCommand:   strings.Replace(analysis.DevCommand, "npm run", "yarn", 1),
				Port:         analysis.Port,
				Environment:  analysis.Environment,
				IsRecommended: false,
				Description:  "Alternative yarn installation (manual deployment)",
			})
	}

	// Custom deployment option
	if len(options) > 0 {
		options = append(options, models.InstallOption{
			Name:         "custom",
			Command:      "",
			BuildCommand: "",
			StartCommand: "",
			DevCommand:   "",
			Port:         analysis.Port,
			Environment:  make(map[string]interface{}),
			IsRecommended: false,
			Description:  "Custom deployment configuration",
		})
	}

	return options
}

// generateInsights creates helpful insights based on analysis
func (h *GitHubHandler) generateInsights(analysis *ProjectAnalysisResponse) []string {
	var insights []string

	// Framework-specific insights
	switch analysis.ProjectType {
	case "react":
		insights = append(insights, "React application detected - make sure to set REACT_APP_ environment variables")
		if analysis.Framework == "vite" {
			insights = append(insights, "Vite detected - very fast build times expected")
		}
	case "vue":
		insights = append(insights, "Vue.js application detected - configure VITE_ environment variables")
	case "nextjs":
		insights = append(insights, "Next.js application detected - supports both SSR and static generation")
		insights = append(insights, "Make sure to set up proper environment variables for API routes")
	case "angular":
		insights = append(insights, "Angular application detected - requires Node.js 16+ for building")
	case "svelte":
		insights = append(insights, "Svelte application detected - very lightweight runtime")
	}

	// Docker insights
	if analysis.HasDocker {
		insights = append(insights, "Dockerfile found - container deployment recommended")
		insights = append(insights, "Make sure to expose the correct port in your Docker configuration")
	}

	// Package manager insights
	switch analysis.PackageManager {
	case "yarn":
		insights = append(insights, "Yarn detected - faster installation than npm")
	case "pnpm":
		insights = append(insights, "pnpm detected - most disk space efficient package manager")
	}

	// Port insights
	if analysis.Port != 3000 && analysis.Port != 5173 && analysis.Port != 4200 {
		insights = append(insights, fmt.Sprintf("Custom port %d detected - make sure to configure your reverse proxy", analysis.Port))
	}

	// Environment insights
	if len(analysis.Environment) > 0 {
		insights = append(insights, "Environment variables detected - review and configure them for your deployment")
	}

	return insights
}

// calculateComplexity estimates project complexity (1-10 scale)
func (h *GitHubHandler) calculateComplexity(analysis *ProjectAnalysisResponse) int {
	complexity := 1

	// Base complexity by project type
	switch analysis.ProjectType {
	case "angular":
		complexity += 3 // Angular is more complex
	case "nextjs":
		complexity += 2 // Next.js has SSR complexity
	case "react", "vue":
		complexity += 1 // Modern frameworks
	}

	// Framework complexity
	if analysis.Framework == "webpack" {
		complexity += 1 // Webpack config can be complex
	}

	// Docker adds complexity
	if analysis.HasDocker {
		complexity += 1
	}

	// Environment variables add complexity
	if len(analysis.Environment) > 5 {
		complexity += 1
	}

	// TypeScript adds slight complexity
	if analysis.Language == "typescript" {
		complexity += 1
	}

	// Cap at 10
	if complexity > 10 {
		complexity = 10
	}

	return complexity
}

// GetRepositoryAnalysis returns the stored analysis for a repository
func (h *GitHubHandler) GetRepositoryAnalysis(c *gin.Context) {
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

	// Find repository analysis
	var analysis models.RepositoryAnalysis
	if err := h.db.Where("github_repository_id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&analysis).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Repository analysis not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch repository analysis"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"analysis": analysis})
}

// ReAnalyzeRepository performs a fresh analysis of a repository and saves it
func (h *GitHubHandler) ReAnalyzeRepository(c *gin.Context) {
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

	// Get branch from query parameter or use repository default
	branch := c.DefaultQuery("branch", repository.Branch)

	// Create analysis request
	req := ProjectAnalysisRequest{
		RepoURL:  repository.CloneURL,
		Branch:   branch,
		SSHKeyID: repository.SSHKeyID,
	}

	// Perform analysis and save to database
	analysis, err := h.analyzeAndSaveRepository(uint(projectID), uint(repoID), req, &repository)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze repository",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Repository analyzed successfully",
		"analysis": analysis,
	})
}

// AnalyzeAndSaveRepository analyzes a repository for an existing GitHub repository record
func (h *GitHubHandler) AnalyzeAndSaveRepository(c *gin.Context) {
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

	var req ProjectAnalysisRequest
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

	// Use repository URL if not provided
	if req.RepoURL == "" {
		req.RepoURL = repository.CloneURL
	}

	// Set default branch if not provided
	if req.Branch == "" {
		req.Branch = repository.Branch
	}

	// Perform analysis and save to database
	analysis, err := h.analyzeAndSaveRepository(uint(projectID), uint(repoID), req, &repository)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to analyze repository",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Repository analyzed and saved successfully",
		"analysis": analysis,
	})
}

// CIPCheckRequest represents a request to check CIP compliance
type CIPCheckRequest struct {
	RepoURL string `json:"repo_url" binding:"required"`
	Branch  string `json:"branch"`
}

// CIPCheckResponse represents the CIP compliance check response
type CIPCheckResponse struct {
	IsCIPCompliant bool   `json:"is_cip_compliant"`
	Message        string `json:"message"`
	Details        *struct {
		HasCloudBoxJSON bool     `json:"has_cloudbox_json"`
		HasInstallScript bool    `json:"has_install_script"`
		HasStartScript   bool    `json:"has_start_script"`
		HasStatusScript  bool    `json:"has_status_script"`
		HasStopScript    bool    `json:"has_stop_script"`
		MissingScripts   []string `json:"missing_scripts,omitempty"`
	} `json:"details,omitempty"`
}

// CheckCIPCompliance checks if a repository is CIP compliant
func (h *GitHubHandler) CheckCIPCompliance(c *gin.Context) {
	var req CIPCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project ID and repo ID from URL
	projectID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	repoID, err := strconv.Atoi(c.Param("repo_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid repository ID"})
		return
	}

	// Get repository from database
	var repository models.GitHubRepository
	if err := h.db.Where("id = ? AND project_id = ?", repoID, projectID).First(&repository).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Set default branch if not provided
	branch := req.Branch
	if branch == "" {
		branch = repository.Branch
	}
	if branch == "" {
		branch = "main"
	}

	// Check CIP compliance by looking for cloudbox.json and required scripts
	compliance, err := h.checkCIPCompliance(req.RepoURL, branch, &repository)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check CIP compliance",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, compliance)
}

// checkCIPCompliance performs the actual CIP compliance check
func (h *GitHubHandler) checkCIPCompliance(repoURL, branch string, repository *models.GitHubRepository) (*CIPCheckResponse, error) {
	response := &CIPCheckResponse{
		IsCIPCompliant: false,
		Details: &struct {
			HasCloudBoxJSON bool     `json:"has_cloudbox_json"`
			HasInstallScript bool    `json:"has_install_script"`
			HasStartScript   bool    `json:"has_start_script"`
			HasStatusScript  bool    `json:"has_status_script"`
			HasStopScript    bool    `json:"has_stop_script"`
			MissingScripts   []string `json:"missing_scripts,omitempty"`
		}{},
	}

	// Check for cloudbox.json in repository
	// Remove .git suffix from repository full name for API calls
	repoName := strings.TrimSuffix(repository.FullName, ".git")
	cloudboxJSONURL := fmt.Sprintf("https://api.github.com/repos/%s/contents/cloudbox.json?ref=%s", repoName, branch)
	
	// Extensive debug logging
	log.Printf("=== CIP Check Debug Info ===")
	log.Printf("Original FullName: %s", repository.FullName)
	log.Printf("Cleaned RepoName: %s", repoName)
	log.Printf("Branch: %s", branch)
	log.Printf("Request URL: %s", cloudboxJSONURL)
	log.Printf("Has Access Token: %t", repository.AccessToken != "")
	if repository.AccessToken != "" {
		log.Printf("Token prefix: %s...", repository.AccessToken[:20])
	}
	
	req, err := http.NewRequest("GET", cloudboxJSONURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add GitHub token if available
	if repository.AccessToken != "" {
		req.Header.Set("Authorization", "token "+repository.AccessToken)
		log.Printf("Authorization header set with token")
	} else {
		log.Printf("WARNING: No access token available - API call may fail for private repos")
	}

	// Add User-Agent header for GitHub API
	req.Header.Set("User-Agent", "CloudBox-CIP-Check/1.0")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("HTTP request failed: %v", err)
		return nil, fmt.Errorf("failed to check cloudbox.json at %s: %w", cloudboxJSONURL, err)
	}
	defer resp.Body.Close()

	// Debug logging
	log.Printf("HTTP Response: Status=%d, Headers=%v", resp.StatusCode, resp.Header)
	fmt.Printf("CIP Check: URL=%s, Status=%d, Repo=%s, Branch=%s\n", cloudboxJSONURL, resp.StatusCode, repoName, branch)

	if resp.StatusCode == http.StatusOK {
		response.Details.HasCloudBoxJSON = true
		
		// Parse cloudbox.json to check for required scripts
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read cloudbox.json: %w", err)
		}

		// GitHub API returns content in base64
		var gitHubResponse struct {
			Content string `json:"content"`
		}
		
		if err := json.Unmarshal(body, &gitHubResponse); err != nil {
			response.Message = "Failed to parse GitHub API response"
			return response, nil
		}
		
		// Decode base64 content
		decodedContent, err := base64.StdEncoding.DecodeString(strings.ReplaceAll(gitHubResponse.Content, "\n", ""))
		if err != nil {
			response.Message = "Failed to decode cloudbox.json content"
			return response, nil
		}
		
		// Now parse the actual cloudbox.json content
		var cloudboxConfig struct {
			Scripts map[string]string `json:"scripts"`
		}
		
		if err := json.Unmarshal(decodedContent, &cloudboxConfig); err != nil {
			response.Message = "cloudbox.json found but invalid JSON format"
			return response, nil
		}

		// Check for required scripts
		requiredScripts := []string{"install", "start", "status", "stop"}
		var missingScripts []string

		for _, script := range requiredScripts {
			if scriptPath, exists := cloudboxConfig.Scripts[script]; exists && scriptPath != "" {
				switch script {
				case "install":
					response.Details.HasInstallScript = true
				case "start":
					response.Details.HasStartScript = true
				case "status":
					response.Details.HasStatusScript = true
				case "stop":
					response.Details.HasStopScript = true
				}
			} else {
				missingScripts = append(missingScripts, script)
			}
		}

		response.Details.MissingScripts = missingScripts

		if len(missingScripts) == 0 {
			response.IsCIPCompliant = true
			response.Message = "Repository is CIP compliant with all required scripts"
		} else {
			response.Message = fmt.Sprintf("cloudbox.json found but missing scripts: %s", strings.Join(missingScripts, ", "))
		}
	} else if resp.StatusCode == http.StatusNotFound {
		// Read response body for more details
		body, _ := io.ReadAll(resp.Body)
		response.Message = "No cloudbox.json found - not CIP compliant"
		log.Printf("CIP Check: cloudbox.json not found at %s (404 error)", cloudboxJSONURL)
		log.Printf("GitHub API 404 response body: %s", string(body))
	} else if resp.StatusCode == http.StatusUnauthorized {
		response.Message = "Access denied - check GitHub token permissions"
	} else if resp.StatusCode == http.StatusForbidden {
		response.Message = "Forbidden - GitHub API rate limit or insufficient permissions"
	} else {
		// Read response body for more details
		body, _ := io.ReadAll(resp.Body)
		response.Message = fmt.Sprintf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return response, nil
}