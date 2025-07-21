package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// DeploymentHandler handles deployment operations
type DeploymentHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewDeploymentHandler creates a new deployment handler
func NewDeploymentHandler(db *gorm.DB, cfg *config.Config) *DeploymentHandler {
	return &DeploymentHandler{db: db, cfg: cfg}
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

	// TODO: Start deployment process in background
	// For now, simulate deployment process
	go h.simulateDeployment(deployment, req.CommitHash, req.Branch)

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

// simulateDeployment simulates the deployment process
func (h *DeploymentHandler) simulateDeployment(deployment models.Deployment, commitHash, branch string) {
	// Simulate build process
	time.Sleep(2 * time.Second)
	h.db.Model(&deployment).Updates(map[string]interface{}{
		"status":     "building",
		"build_logs": "Starting build process...\nCloning repository...\nInstalling dependencies...\n",
	})

	time.Sleep(3 * time.Second)
	h.db.Model(&deployment).Updates(map[string]interface{}{
		"status":     "deploying",
		"build_logs": "Starting build process...\nCloning repository...\nInstalling dependencies...\nBuild completed successfully!\n",
		"deploy_logs": "Connecting to server...\nUploading files...\n",
	})

	time.Sleep(2 * time.Second)
	now := time.Now()
	h.db.Model(&deployment).Updates(map[string]interface{}{
		"status":      "deployed",
		"deployed_at": &now,
		"deploy_logs": "Connecting to server...\nUploading files...\nStarting application...\nDeployment completed successfully!\n",
		"build_time":  int64(3000),  // 3 seconds
		"deploy_time": int64(2000),  // 2 seconds
		"file_count":  int64(42),
		"total_size":  int64(1024 * 1024), // 1 MB
	})
}