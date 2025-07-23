package handlers

import (
	"net/http"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// BackupHandler handles backup-related requests
type BackupHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(db *gorm.DB, cfg *config.Config) *BackupHandler {
	return &BackupHandler{db: db, cfg: cfg}
}

// ListBackups returns all backups for the authenticated user
func (h *BackupHandler) ListBackups(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup functionality is not yet implemented",
		"code":  "NOT_IMPLEMENTED",
	})
}

// CreateBackup creates a new backup
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup functionality is not yet implemented",
		"code":  "NOT_IMPLEMENTED",
	})
}

// GetBackup returns a specific backup
func (h *BackupHandler) GetBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup functionality is not yet implemented",
		"code":  "NOT_IMPLEMENTED",
	})
}

// DeleteBackup deletes a backup
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup functionality is not yet implemented",
		"code":  "NOT_IMPLEMENTED",
	})
}

// RestoreBackup restores from a backup
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Backup functionality is not yet implemented",
		"code":  "NOT_IMPLEMENTED",
	})
}