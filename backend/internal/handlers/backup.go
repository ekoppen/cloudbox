package handlers

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BackupHandler handles backup-related requests
type BackupHandler struct {
	db            *gorm.DB
	cfg           *config.Config
	backupService *services.BackupService
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(db *gorm.DB, cfg *config.Config) *BackupHandler {
	// Create backup directory if not specified in config
	backupDir := "/var/lib/cloudbox/backups"
	if cfg.BackupDir != "" {
		backupDir = cfg.BackupDir
	}
	
	backupService := services.NewBackupService(db, backupDir)
	
	return &BackupHandler{
		db:            db,
		cfg:           cfg,
		backupService: backupService,
	}
}

// CreateBackupRequest represents a request to create a backup
type CreateBackupRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"` // manual, automatic
}

// RestoreBackupRequest represents a request to restore from backup
type RestoreBackupRequest struct {
	TargetProjectID uint `json:"target_project_id"`
}

// ListBackups returns all backups for a project
func (h *BackupHandler) ListBackups(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	backups, err := h.backupService.ListBackups(uint(projectID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch backups"})
		return
	}

	c.JSON(http.StatusOK, backups)
}

// CreateBackup creates a new backup
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Type == "" {
		req.Type = "manual"
	}

	// Create backup using service
	backup, err := h.backupService.CreateBackup(uint(projectID), req.Type)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to create backup: %v", err)})
		return
	}

	c.JSON(http.StatusCreated, backup)
}

// GetBackup returns a specific backup
func (h *BackupHandler) GetBackup(c *gin.Context) {
	backupID, err := strconv.ParseUint(c.Param("backup_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	backup, err := h.backupService.GetBackup(uint(backupID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
		return
	}

	c.JSON(http.StatusOK, backup)
}

// DeleteBackup deletes a backup
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	backupID, err := strconv.ParseUint(c.Param("backup_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	if err := h.backupService.DeleteBackup(uint(backupID)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to delete backup: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Backup deleted successfully"})
}

// RestoreBackup restores from a backup
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	backupID, err := strconv.ParseUint(c.Param("backup_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid backup ID"})
		return
	}

	var req RestoreBackupRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// If no target project specified, restore to original project
	if req.TargetProjectID == 0 {
		// Get backup to find original project ID
		backup, err := h.backupService.GetBackup(uint(backupID))
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Backup not found"})
			return
		}
		req.TargetProjectID = backup.ProjectID
	}

	if err := h.backupService.RestoreBackup(uint(backupID), req.TargetProjectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Failed to restore backup: %v", err)})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":           "Backup restore completed successfully",
		"target_project_id": req.TargetProjectID,
	})
}