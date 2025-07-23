package services

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// AuditService handles audit logging
type AuditService struct {
	db *gorm.DB
}

// NewAuditService creates a new audit service
func NewAuditService(db *gorm.DB) *AuditService {
	return &AuditService{db: db}
}

// LogAction logs an audit trail entry
func (s *AuditService) LogAction(c *gin.Context, action models.AuditLogAction, resource, resourceID, description string, success bool, errorMsg string, metadata interface{}) error {
	// Get actor information from context
	actorID := c.GetUint("user_id")
	actorName := c.GetString("user_name")
	actorRole := c.GetString("user_role")

	// If name is not in context, try to get it from user_email
	if actorName == "" {
		actorName = c.GetString("user_email")
	}

	// Serialize metadata to JSON
	var metadataJSON string
	if metadata != nil {
		if jsonBytes, err := json.Marshal(metadata); err == nil {
			metadataJSON = string(jsonBytes)
		}
	}

	// Get project ID if applicable
	var projectID *uint
	if projectParam := c.Param("id"); projectParam != "" {
		if id, err := strconv.ParseUint(projectParam, 10, 32); err == nil {
			pid := uint(id)
			projectID = &pid
		}
	}

	// Create audit log entry
	auditLog := models.AuditLog{
		Action:      action,
		Resource:    resource,
		ResourceID:  resourceID,
		Description: description,
		ActorID:     actorID,
		ActorName:   actorName,
		ActorRole:   actorRole,
		IPAddress:   c.ClientIP(),
		UserAgent:   c.GetHeader("User-Agent"),
		Method:      c.Request.Method,
		Path:        c.Request.URL.Path,
		Metadata:    metadataJSON,
		ProjectID:   projectID,
		Success:     success,
		ErrorMsg:    errorMsg,
	}

	// Save to database
	return s.db.Create(&auditLog).Error
}

// LogProjectDeletion logs project deletion with additional context
func (s *AuditService) LogProjectDeletion(c *gin.Context, projectID uint, projectName string, success bool, errorMsg string) error {
	metadata := map[string]interface{}{
		"project_name": projectName,
		"deleted_by_superadmin": c.GetString("user_role") == "superadmin",
	}

	return s.LogAction(
		c,
		models.AuditActionProjectDelete,
		"project",
		fmt.Sprintf("%d", projectID),
		fmt.Sprintf("Project '%s' deleted", projectName),
		success,
		errorMsg,
		metadata,
	)
}

// LogProjectCreation logs project creation
func (s *AuditService) LogProjectCreation(c *gin.Context, projectID uint, projectName string, success bool, errorMsg string) error {
	metadata := map[string]interface{}{
		"project_name": projectName,
	}

	return s.LogAction(
		c,
		models.AuditActionProjectCreate,
		"project",
		fmt.Sprintf("%d", projectID),
		fmt.Sprintf("Project '%s' created", projectName),
		success,
		errorMsg,
		metadata,
	)
}

// LogLogin logs user login
func (s *AuditService) LogLogin(c *gin.Context, userID uint, userName, userRole string, success bool, errorMsg string) error {
	metadata := map[string]interface{}{
		"user_role": userRole,
	}

	return s.LogAction(
		c,
		models.AuditActionLogin,
		"user",
		fmt.Sprintf("%d", userID),
		fmt.Sprintf("User '%s' logged in", userName),
		success,
		errorMsg,
		metadata,
	)
}

// GetAuditLogs retrieves audit logs with filtering
func (s *AuditService) GetAuditLogs(action string, resource string, actorID uint, limit int, offset int) ([]models.AuditLog, int64, error) {
	var logs []models.AuditLog
	var total int64

	query := s.db.Model(&models.AuditLog{})

	// Apply filters
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resource != "" {
		query = query.Where("resource = ?", resource)
	}
	if actorID > 0 {
		query = query.Where("actor_id = ?", actorID)
	}

	// Get total count
	query.Count(&total)

	// Get paginated results
	err := query.Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&logs).Error

	return logs, total, err
}