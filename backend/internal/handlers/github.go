package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GitHubHandler handles GitHub repository management
type GitHubHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewGitHubHandler creates a new GitHub handler
func NewGitHubHandler(db *gorm.DB, cfg *config.Config) *GitHubHandler {
	return &GitHubHandler{db: db, cfg: cfg}
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
}

// ListGitHubRepositories returns all GitHub repositories for a project
func (h *GitHubHandler) ListGitHubRepositories(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var repositories []models.GitHubRepository
	if err := h.db.Where("project_id = ?", uint(projectID)).Find(&repositories).Error; err != nil {
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
	if err := h.db.Where("id = ? AND project_id = ?", uint(repoID), uint(projectID)).First(&repository).Error; err != nil {
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

	if err := h.db.Model(&repository).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update GitHub repository"})
		return
	}

	// Reload repository
	if err := h.db.First(&repository, repository.ID).Error; err != nil {
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

	// TODO: Implement actual GitHub API sync
	// For now, just update the sync timestamp
	now := time.Now()
	if err := h.db.Model(&repository).Update("last_sync_at", now).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update sync timestamp"})
		return
	}

	repository.LastSyncAt = &now

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

// generateWebhookSecret generates a random webhook secret
func (h *GitHubHandler) generateWebhookSecret() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}