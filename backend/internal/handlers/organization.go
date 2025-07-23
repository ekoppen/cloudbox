package handlers

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// OrganizationHandler handles organization requests
type OrganizationHandler struct {
	db           *gorm.DB
	cfg          *config.Config
	auditService *services.AuditService
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(db *gorm.DB, cfg *config.Config) *OrganizationHandler {
	return &OrganizationHandler{
		db:           db,
		cfg:          cfg,
		auditService: services.NewAuditService(db),
	}
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
		Name          string `json:"name" binding:"required"`
		Description   string `json:"description"`
		Color         string `json:"color"`
		Website       string `json:"website"`
		Email         string `json:"email"`
		Phone         string `json:"phone"`
		ContactPerson string `json:"contact_person"`
		LogoURL       string `json:"logo_url"`
		Address       string `json:"address"`
		City          string `json:"city"`
		Country       string `json:"country"`
		PostalCode    string `json:"postal_code"`
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
		Name:          req.Name,
		Description:   req.Description,
		Color:         color,
		Website:       req.Website,
		Email:         req.Email,
		Phone:         req.Phone,
		ContactPerson: req.ContactPerson,
		LogoURL:       req.LogoURL,
		Address:       req.Address,
		City:          req.City,
		Country:       req.Country,
		PostalCode:    req.PostalCode,
		UserID:        userID,
	}

	if err := h.db.Create(&organization).Error; err != nil {
		// Log failed creation
		h.auditService.LogOrganizationCreation(c, 0, req.Name, false, fmt.Sprintf("Database error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create organization"})
		return
	}

	// Log successful creation
	h.auditService.LogOrganizationCreation(c, organization.ID, organization.Name, true, "")

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
		Name          *string `json:"name"`
		Description   *string `json:"description"`
		Color         *string `json:"color"`
		IsActive      *bool   `json:"is_active"`
		Website       *string `json:"website"`
		Email         *string `json:"email"`
		Phone         *string `json:"phone"`
		ContactPerson *string `json:"contact_person"`
		LogoURL       *string `json:"logo_url"`
		Address       *string `json:"address"`
		City          *string `json:"city"`
		Country       *string `json:"country"`
		PostalCode    *string `json:"postal_code"`
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

	// Store original values for change tracking
	original := organization

	// Update fields and track changes
	updates := make(map[string]interface{})
	var changedFields []string

	if req.Name != nil && *req.Name != organization.Name {
		updates["name"] = *req.Name
		changedFields = append(changedFields, "name")
	}
	if req.Description != nil && *req.Description != organization.Description {
		updates["description"] = *req.Description
		changedFields = append(changedFields, "description")
	}
	if req.Color != nil && *req.Color != organization.Color {
		updates["color"] = *req.Color
		changedFields = append(changedFields, "color")
	}
	if req.IsActive != nil && *req.IsActive != organization.IsActive {
		updates["is_active"] = *req.IsActive
		changedFields = append(changedFields, "is_active")
	}
	if req.Website != nil && *req.Website != organization.Website {
		updates["website"] = *req.Website
		changedFields = append(changedFields, "website")
	}
	if req.Email != nil && *req.Email != organization.Email {
		updates["email"] = *req.Email
		changedFields = append(changedFields, "email")
	}
	if req.Phone != nil && *req.Phone != organization.Phone {
		updates["phone"] = *req.Phone
		changedFields = append(changedFields, "phone")
	}
	if req.ContactPerson != nil && *req.ContactPerson != organization.ContactPerson {
		updates["contact_person"] = *req.ContactPerson
		changedFields = append(changedFields, "contact_person")
	}
	if req.LogoURL != nil && *req.LogoURL != organization.LogoURL {
		updates["logo_url"] = *req.LogoURL
		changedFields = append(changedFields, "logo_url")
	}
	if req.Address != nil && *req.Address != organization.Address {
		updates["address"] = *req.Address
		changedFields = append(changedFields, "address")
	}
	if req.City != nil && *req.City != organization.City {
		updates["city"] = *req.City
		changedFields = append(changedFields, "city")
	}
	if req.Country != nil && *req.Country != organization.Country {
		updates["country"] = *req.Country
		changedFields = append(changedFields, "country")
	}
	if req.PostalCode != nil && *req.PostalCode != organization.PostalCode {
		updates["postal_code"] = *req.PostalCode
		changedFields = append(changedFields, "postal_code")
	}

	// If no changes, return current organization
	if len(changedFields) == 0 {
		c.JSON(http.StatusOK, organization)
		return
	}

	// Apply updates
	if err := h.db.Model(&organization).Updates(updates).Error; err != nil {
		// Log failed update
		h.auditService.LogOrganizationUpdate(c, uint(orgID), organization.Name, changedFields, false, fmt.Sprintf("Database error: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update organization"})
		return
	}

	// Log successful update
	h.auditService.LogOrganizationUpdate(c, uint(orgID), organization.Name, changedFields, true, "")

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
			// Log failed deletion - organization not found
			h.auditService.LogOrganizationDeletion(c, uint(orgID), "Unknown", 0, false, "Organization not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		} else {
			// Log failed deletion - database error
			h.auditService.LogOrganizationDeletion(c, uint(orgID), "Unknown", 0, false, fmt.Sprintf("Database error: %v", err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization"})
		}
		return
	}

	// Count projects in this organization
	var projectCount int64
	tx.Model(&models.Project{}).Where("organization_id = ?", orgID).Count(&projectCount)

	// Move all projects out of this organization (set organization_id to NULL)
	if err := tx.Model(&models.Project{}).Where("organization_id = ?", orgID).Update("organization_id", nil).Error; err != nil {
		tx.Rollback()
		// Log failed deletion - project update error
		h.auditService.LogOrganizationDeletion(c, uint(orgID), organization.Name, int(projectCount), false, fmt.Sprintf("Failed to update projects: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update projects"})
		return
	}

	// Delete organization
	if err := tx.Delete(&organization).Error; err != nil {
		tx.Rollback()
		// Log failed deletion - organization delete error
		h.auditService.LogOrganizationDeletion(c, uint(orgID), organization.Name, int(projectCount), false, fmt.Sprintf("Failed to delete organization: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete organization"})
		return
	}

	tx.Commit()

	// Log successful deletion
	h.auditService.LogOrganizationDeletion(c, uint(orgID), organization.Name, int(projectCount), true, "")

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