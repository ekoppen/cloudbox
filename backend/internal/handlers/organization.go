package handlers

import (
	"net/http"
	"strconv"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrganizationHandler handles organization requests
type OrganizationHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(db *gorm.DB, cfg *config.Config) *OrganizationHandler {
	return &OrganizationHandler{db: db, cfg: cfg}
}

// ListOrganizations returns all organizations for the current user
func (h *OrganizationHandler) ListOrganizations(c *gin.Context) {
	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var organizations []models.Organization
	query := h.db.Model(&models.Organization{})

	// Super admins can see all organizations, others only their own
	if userRole != "superadmin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.Find(&organizations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organizations"})
		return
	}

	c.JSON(http.StatusOK, organizations)
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c *gin.Context) {
	userID := c.GetUint("user_id")

	var req struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		Color       string `json:"color"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default color if not provided
	color := req.Color
	if color == "" {
		color = "#3B82F6"
	}

	organization := models.Organization{
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		UserID:      userID,
	}

	if err := h.db.Create(&organization).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	c.JSON(http.StatusCreated, organization)
}

// GetOrganization returns a specific organization
func (h *OrganizationHandler) GetOrganization(c *gin.Context) {
	orgID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var organization models.Organization
	query := h.db.Where("id = ?", orgID)

	// Super admins can see any organization, others only their own
	if userRole != "superadmin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization"})
		}
		return
	}

	c.JSON(http.StatusOK, organization)
}

// UpdateOrganization updates an organization
func (h *OrganizationHandler) UpdateOrganization(c *gin.Context) {
	orgID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
		Color       *string `json:"color"`
		IsActive    *bool   `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find organization
	var organization models.Organization
	query := h.db.Where("id = ?", orgID)

	// Super admins can update any organization, others only their own
	if userRole != "superadmin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization"})
		}
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	
	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.db.Model(&organization).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	// Return updated organization
	h.db.Where("id = ?", orgID).First(&organization)
	c.JSON(http.StatusOK, organization)
}

// DeleteOrganization deletes an organization
func (h *OrganizationHandler) DeleteOrganization(c *gin.Context) {
	orgID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find organization
	var organization models.Organization
	query := tx.Where("id = ?", orgID)

	// Super admins can delete any organization, others only their own
	if userRole != "superadmin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&organization).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization"})
		}
		return
	}

	// Move all projects out of this organization (set organization_id to NULL)
	if err := tx.Model(&models.Project{}).Where("organization_id = ?", orgID).Update("organization_id", nil).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update projects"})
		return
	}

	// Delete organization
	if err := tx.Delete(&organization).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
		return
	}

	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Organization deleted successfully"})
}

// GetOrganizationProjects returns all projects in an organization
func (h *OrganizationHandler) GetOrganizationProjects(c *gin.Context) {
	orgID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}

	userID := c.GetUint("user_id")
	userRole := c.GetString("user_role")

	// Verify organization access
	var organization models.Organization
	query := h.db.Where("id = ?", orgID)

	if userRole != "superadmin" {
		query = query.Where("user_id = ?", userID)
	}

	if err := query.First(&organization).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization"})
		}
		return
	}

	// Get projects in organization
	var projects []models.Project
	projectQuery := h.db.Where("organization_id = ?", orgID)

	if userRole != "superadmin" {
		projectQuery = projectQuery.Where("user_id = ?", userID)
	}

	if err := projectQuery.Find(&projects).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch projects"})
		return
	}

	c.JSON(http.StatusOK, projects)
}