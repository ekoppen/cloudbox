package handlers

import (
	"fmt"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// PublicFileHandler handles public file serving requests
type PublicFileHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewPublicFileHandler creates a new public file handler
func NewPublicFileHandler(db *gorm.DB, cfg *config.Config) *PublicFileHandler {
	return &PublicFileHandler{db: db, cfg: cfg}
}

// ServePublicFile serves public files with security validation
// GET /public/{project_slug}/{bucket_name}/{file_path}
func (h *PublicFileHandler) ServePublicFile(c *gin.Context) {
	projectSlug := c.Param("project_slug")
	bucketName := c.Param("bucket_name")
	filePath := c.Param("file_path")

	// Security: Clean the file path to prevent directory traversal
	filePath = filepath.Clean(filePath)
	if strings.Contains(filePath, "..") || strings.HasPrefix(filePath, "/") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file path"})
		return
	}

	// 1. Validate project (fail fast)
	project, err := h.validateActiveProject(projectSlug)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found or inactive"})
		return
	}

	// 2. Validate bucket (fail fast)
	bucket, err := h.validatePublicBucket(project.ID, bucketName)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Bucket not found or not public"})
		return
	}

	// 3. Validate and serve file (fail fast)
	file, err := h.validateFile(project.ID, bucketName, filePath)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// 4. Serve file with proper headers
	h.serveFileWithCache(c, file, bucket)
}

// validateActiveProject checks if project exists and is active
func (h *PublicFileHandler) validateActiveProject(projectSlug string) (*models.Project, error) {
	var project models.Project
	if err := h.db.Where("slug = ? AND is_active = ?", projectSlug, true).First(&project).Error; err != nil {
		return nil, err
	}
	return &project, nil
}

// validatePublicBucket checks if bucket exists and is public
func (h *PublicFileHandler) validatePublicBucket(projectID uint, bucketName string) (*models.Bucket, error) {
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ? AND is_public = ?", projectID, bucketName, true).First(&bucket).Error; err != nil {
		return nil, err
	}
	return &bucket, nil
}

// validateFile checks if file exists and belongs to project/bucket
func (h *PublicFileHandler) validateFile(projectID uint, bucketName, filePath string) (*models.File, error) {
	var file models.File
	if err := h.db.Where("project_id = ? AND bucket_name = ? AND file_path = ?", projectID, bucketName, filePath).First(&file).Error; err != nil {
		return nil, err
	}
	return &file, nil
}

// serveFileWithCache serves the file with proper caching headers
func (h *PublicFileHandler) serveFileWithCache(c *gin.Context, file *models.File, bucket *models.Bucket) {
	// Build full file path
	fullPath := filepath.Join("./uploads", fmt.Sprintf("%d", file.ProjectID), file.BucketName, file.FilePath)

	// Check if file exists on disk
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		log.Printf("File not found on disk: %s - %v", fullPath, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}

	// Open the file
	fileHandle, err := os.Open(fullPath)
	if err != nil {
		log.Printf("Failed to open file: %s - %v", fullPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read file"})
		return
	}
	defer fileHandle.Close()

	// Set MIME type
	contentType := mime.TypeByExtension(filepath.Ext(file.OriginalName))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Set security and caching headers
	c.Header("Content-Type", contentType)
	c.Header("Cache-Control", "public, max-age=3600") // 1 hour browser cache
	c.Header("Last-Modified", fileInfo.ModTime().UTC().Format(http.TimeFormat))
	c.Header("ETag", fmt.Sprintf(`"%s-%d"`, file.ID, fileInfo.ModTime().Unix()))
	
	// Security headers
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("X-Frame-Options", "SAMEORIGIN")
	
	// For images, allow embedding
	if strings.HasPrefix(contentType, "image/") {
		c.Header("X-Frame-Options", "SAMEORIGIN")
	}

	// Check if client has cached version (ETag)
	if match := c.GetHeader("If-None-Match"); match != "" {
		if match == fmt.Sprintf(`"%s-%d"`, file.ID, fileInfo.ModTime().Unix()) {
			c.Status(http.StatusNotModified)
			return
		}
	}

	// Check if client has cached version (Last-Modified)
	if modSince := c.GetHeader("If-Modified-Since"); modSince != "" {
		if t, err := time.Parse(http.TimeFormat, modSince); err == nil {
			if !fileInfo.ModTime().After(t) {
				c.Status(http.StatusNotModified)
				return
			}
		}
	}

	// Set filename for downloads (not for images to allow embedding)
	if !strings.HasPrefix(contentType, "image/") {
		c.Header("Content-Disposition", fmt.Sprintf(`inline; filename="%s"`, file.OriginalName))
	}

	// Serve the file content
	http.ServeContent(c.Writer, c.Request, file.OriginalName, fileInfo.ModTime(), fileHandle)
}

// Public file access statistics (for future monitoring)
func (h *PublicFileHandler) LogPublicAccess(projectID uint, bucketName, filePath, clientIP string) {
	// Future: Log public file access for monitoring/analytics
	log.Printf("Public file access: Project=%d, Bucket=%s, File=%s, IP=%s", 
		projectID, bucketName, filePath, clientIP)
}