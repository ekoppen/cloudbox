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
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// ProjectHandler handles project-related requests
type ProjectHandler struct {
	db           *gorm.DB
	cfg          *config.Config
	auditService *services.AuditService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(db *gorm.DB, cfg *config.Config) *ProjectHandler {
	return &ProjectHandler{
		db:           db,
		cfg:          cfg,
		auditService: services.NewAuditService(db),
	}
}

// CreateProjectRequest represents a project creation request
type CreateProjectRequest struct {
	Name           string `json:"name" binding:"required"`
	Description    string `json:"description"`
	OrganizationID *uint  `json:"organization_id"`
}

// ListProjects returns all projects for the authenticated user
// Superadmins can see all projects, regular admins only see their own
func (h *ProjectHandler) ListProjects(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	
	var projects []models.Project
	var query *gorm.DB
	
	if userRole == "superadmin" {
		// Superadmin can see all projects
		query = h.db.Preload("User")
	} else {
		// Regular admin can only see their own projects
		query = h.db.Where("user_id = ?", userID)
	}
	
	if err := query.Preload("Organization").Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}

// CreateProject creates a new project
func (h *ProjectHandler) CreateProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Generate unique slug from name
	slug := generateSlug(req.Name)
	
	// Ensure slug is unique
	var count int64
	h.db.Model(&models.Project{}).Where("slug = ?", slug).Count(&count)
	if count > 0 {
		slug = fmt.Sprintf("%s-%d", slug, count+1)
	}

	// Validate organization if provided
	if req.OrganizationID != nil {
		var organization models.Organization
		if err := h.db.Where("id = ? AND user_id = ?", *req.OrganizationID, userID).First(&organization).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization or access denied"})
			return
		}
	}

	// Create project
	project := models.Project{
		Name:           req.Name,
		Description:    req.Description,
		Slug:           slug,
		UserID:         userID,
		OrganizationID: req.OrganizationID,
		IsActive:       true,
	}

	if err := h.db.Create(&project).Error; err != nil {
		// Log failed creation
		h.auditService.LogProjectCreation(c, 0, req.Name, false, fmt.Sprintf("Database error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create project"})
		return
	}

	// Create default CORS config
	corsConfig := models.CORSConfig{
		ProjectID:        project.ID,
		AllowedOrigins:   pq.StringArray{"*"},
		AllowedMethods:   pq.StringArray{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowedHeaders:   pq.StringArray{"*"},
		AllowCredentials: false,
		MaxAge:           3600,
	}
	h.db.Create(&corsConfig)

	// Log successful creation
	h.auditService.LogProjectCreation(c, project.ID, project.Name, true, "")

	c.JSON(http.StatusCreated, project)
}

// GetProject returns a specific project
func (h *ProjectHandler) GetProject(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	project, canAccess := h.canAccessProject(c, uint(projectID))
	if !canAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Load related data
	if err := h.db.Preload("APIKeys").Preload("CORSConfig").First(&project, project.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load project details"})
		return
	}

	c.JSON(http.StatusOK, project)
}

// UpdateProject updates a project
func (h *ProjectHandler) UpdateProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateProjectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update project
	result := h.db.Model(&models.Project{}).
		Where("id = ? AND user_id = ?", uint(projectID), userID).
		Updates(models.Project{
			Name:        req.Name,
			Description: req.Description,
		})

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update project"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Project updated successfully"})
}

// DeleteProject deletes a project
func (h *ProjectHandler) DeleteProject(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// First, get the project details for audit logging
	var project models.Project
	if err := h.db.Where("id = ?", uint(projectID)).First(&project).Error; err != nil {
		// Log failed attempt
		h.auditService.LogProjectDeletion(c, uint(projectID), "Unknown Project", false, "Project not found")
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var result *gorm.DB
	if userRole == "superadmin" {
		// Superadmin can delete any project
		result = h.db.Where("id = ?", uint(projectID)).Delete(&models.Project{})
	} else {
		// Regular admin can only delete their own projects
		result = h.db.Where("id = ? AND user_id = ?", uint(projectID), userID).Delete(&models.Project{})
	}

	if result.Error != nil {
		// Log failed deletion
		h.auditService.LogProjectDeletion(c, uint(projectID), project.Name, false, fmt.Sprintf("Database error: %v", result.Error))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete project"})
		return
	}

	if result.RowsAffected == 0 {
		// Log failed deletion due to permissions
		h.auditService.LogProjectDeletion(c, uint(projectID), project.Name, false, "Project not found or insufficient permissions")
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Log successful deletion
	h.auditService.LogProjectDeletion(c, uint(projectID), project.Name, true, "")

	c.JSON(http.StatusOK, gin.H{"message": "Project deleted successfully"})
}

// API Key related handlers

// ListAPIKeys returns all API keys for a project
func (h *ProjectHandler) ListAPIKeys(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project access
	_, canAccess := h.canAccessProject(c, uint(projectID))
	if !canAccess {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var apiKeys []models.APIKey
	if err := h.db.Where("project_id = ?", projectID).Find(&apiKeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch API keys"})
		return
	}

	c.JSON(http.StatusOK, apiKeys)
}

// CreateAPIKey creates a new API key for a project
func (h *ProjectHandler) CreateAPIKey(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req struct {
		Name        string   `json:"name" binding:"required"`
		Permissions []string `json:"permissions"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", uint(projectID), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Generate API key
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	apiKey := hex.EncodeToString(keyBytes)

	// Create API key record
	key := models.APIKey{
		Name:        req.Name,
		Key:         apiKey,
		KeyHash:     apiKey, // In production, hash this
		ProjectID:   uint(projectID),
		Permissions: pq.StringArray(req.Permissions),
		IsActive:    true,
	}

	if err := h.db.Create(&key).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create API key"})
		return
	}

	c.JSON(http.StatusCreated, key)
}

// DeleteAPIKey deletes an API key
func (h *ProjectHandler) DeleteAPIKey(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", uint(projectID), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(keyID), projectID).Delete(&models.APIKey{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete API key"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "API key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "API key deleted successfully"})
}

// CORS related handlers

// GetCORSConfig returns CORS configuration for a project
func (h *ProjectHandler) GetCORSConfig(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", uint(projectID), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	var corsConfig models.CORSConfig
	if err := h.db.Where("project_id = ?", projectID).First(&corsConfig).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "CORS config not found"})
		return
	}

	c.JSON(http.StatusOK, corsConfig)
}

// UpdateCORSConfig updates CORS configuration for a project
func (h *ProjectHandler) UpdateCORSConfig(c *gin.Context) {
	userID := c.GetUint("user_id")
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req models.CORSConfig
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify project ownership
	var project models.Project
	if err := h.db.Where("id = ? AND user_id = ?", uint(projectID), userID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Update CORS config
	result := h.db.Model(&models.CORSConfig{}).
		Where("project_id = ?", projectID).
		Updates(req)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update CORS config"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "CORS config updated successfully"})
}

// Project API data handlers (for project-specific APIs)

// GetData handles GET requests to project data
func (h *ProjectHandler) GetData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Data API not yet implemented"})
}

// CreateData handles POST requests to project data
func (h *ProjectHandler) CreateData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Data API not yet implemented"})
}

// UpdateData handles PUT requests to project data
func (h *ProjectHandler) UpdateData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Data API not yet implemented"})
}

// DeleteData handles DELETE requests to project data
func (h *ProjectHandler) DeleteData(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Data API not yet implemented"})
}

// Helper functions

// canAccessProject checks if user can access a project based on role
func (h *ProjectHandler) canAccessProject(c *gin.Context, projectID uint) (models.Project, bool) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")
	
	var project models.Project
	var query *gorm.DB
	
	if userRole == "superadmin" {
		// Superadmin can access any project
		query = h.db.Where("id = ?", projectID)
	} else {
		// Regular admin can only access their own projects
		query = h.db.Where("id = ? AND user_id = ?", projectID, userID)
	}
	
	err := query.First(&project).Error
	if err != nil {
		return project, false
	}
	
	return project, true
}

// generateSlug creates a URL-friendly slug from a name
func generateSlug(name string) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove special characters (basic implementation)
	var result strings.Builder
	for _, r := range slug {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '-' {
			result.WriteRune(r)
		}
	}
	return result.String()
}