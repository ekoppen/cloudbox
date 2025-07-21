package handlers

import (
	"net/http"

	"github.com/cloudbox/backend/internal/config"
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
	c.JSON(http.StatusOK, gin.H{"message": "Backups API not yet implemented"})
}

// CreateBackup creates a new backup
func (h *BackupHandler) CreateBackup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Create backup not yet implemented"})
}

// GetBackup returns a specific backup
func (h *BackupHandler) GetBackup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Get backup not yet implemented"})
}

// DeleteBackup deletes a backup
func (h *BackupHandler) DeleteBackup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Delete backup not yet implemented"})
}

// RestoreBackup restores from a backup
func (h *BackupHandler) RestoreBackup(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "Restore backup not yet implemented"})
}