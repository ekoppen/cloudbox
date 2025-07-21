package handlers

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// StorageHandler handles file storage requests
type StorageHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewStorageHandler creates a new storage handler
func NewStorageHandler(db *gorm.DB, cfg *config.Config) *StorageHandler {
	return &StorageHandler{db: db, cfg: cfg}
}

// Bucket Management

// ListBuckets returns all buckets for a project
func (h *StorageHandler) ListBuckets(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var buckets []models.Bucket
	if err := h.db.Where("project_id = ?", project.ID).Find(&buckets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buckets"})
		return
	}
	
	c.JSON(http.StatusOK, buckets)
}

// CreateBucket creates a new storage bucket
func (h *StorageHandler) CreateBucket(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var req struct {
		Name         string   `json:"name" binding:"required"`
		Description  string   `json:"description"`
		MaxFileSize  *int64   `json:"max_file_size"`
		AllowedTypes []string `json:"allowed_types"`
		IsPublic     *bool    `json:"is_public"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate bucket name
	if !isValidBucketName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid bucket name. Use only letters, numbers, hyphens, and underscores"})
		return
	}
	
	// Check if bucket already exists
	var existingBucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, req.Name).First(&existingBucket).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Bucket already exists"})
		return
	}
	
	// Set defaults
	maxFileSize := int64(52428800) // 50MB default
	if req.MaxFileSize != nil {
		maxFileSize = *req.MaxFileSize
	}
	
	isPublic := false
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}
	
	// Create bucket
	bucket := models.Bucket{
		Name:         req.Name,
		Description:  req.Description,
		MaxFileSize:  maxFileSize,
		AllowedTypes: req.AllowedTypes,
		IsPublic:     isPublic,
		ProjectID:    project.ID,
		FileCount:    0,
		TotalSize:    0,
		LastModified: time.Now(),
	}
	
	if err := h.db.Create(&bucket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket"})
		return
	}
	
	// Create bucket directory on filesystem
	bucketPath := filepath.Join("./uploads", project.Slug, req.Name)
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create bucket directory"})
		return
	}
	
	c.JSON(http.StatusCreated, bucket)
}

// GetBucket returns a specific bucket
func (h *StorageHandler) GetBucket(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	c.JSON(http.StatusOK, bucket)
}

// DeleteBucket deletes a bucket and all its files
func (h *StorageHandler) DeleteBucket(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Delete all files in bucket
	if err := tx.Where("project_id = ? AND bucket_name = ?", project.ID, bucketName).Delete(&models.File{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete files"})
		return
	}
	
	// Delete bucket
	result := tx.Where("project_id = ? AND name = ?", project.ID, bucketName).Delete(&models.Bucket{})
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete bucket"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	tx.Commit()
	
	// Remove bucket directory from filesystem
	bucketPath := filepath.Join("./uploads", project.Slug, bucketName)
	os.RemoveAll(bucketPath)
	
	c.JSON(http.StatusOK, gin.H{"message": "Bucket deleted successfully"})
}

// File Management

// ListFiles returns all files in a bucket
func (h *StorageHandler) ListFiles(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	// Verify bucket exists
	if !h.bucketExists(project.ID, bucketName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Parse query parameters
	limit := 25 // Default limit
	offset := 0
	orderBy := "created_at DESC"
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	if order := c.Query("orderBy"); order != "" {
		// Simple validation for order by
		if strings.Contains(order, "created_at") || strings.Contains(order, "original_name") || strings.Contains(order, "size") {
			orderBy = order
		}
	}
	
	var files []models.File
	query := h.db.Where("project_id = ? AND bucket_name = ?", project.ID, bucketName).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}
	
	// Get total count
	var total int64
	h.db.Model(&models.File{}).Where("project_id = ? AND bucket_name = ?", project.ID, bucketName).Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"files":  files,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UploadFile handles file upload to a bucket
func (h *StorageHandler) UploadFile(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	// Verify bucket exists and get settings
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Parse multipart form
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No file uploaded"})
		return
	}
	defer file.Close()
	
	// Validate file size
	if header.Size > bucket.MaxFileSize {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("File too large. Maximum size: %d bytes", bucket.MaxFileSize),
		})
		return
	}
	
	// Validate MIME type if restricted
	if len(bucket.AllowedTypes) > 0 {
		mimeType := header.Header.Get("Content-Type")
		if !contains(bucket.AllowedTypes, mimeType) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File type not allowed. Allowed types: %v", bucket.AllowedTypes),
			})
			return
		}
	}
	
	// Generate file info
	fileID := uuid.New().String()
	fileName := fmt.Sprintf("%s_%s", fileID, sanitizeFileName(header.Filename))
	filePath := filepath.Join("./uploads", project.Slug, bucketName, fileName)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create directory"})
		return
	}
	
	// Save file to disk
	dst, err := os.Create(filePath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create file"})
		return
	}
	defer dst.Close()
	
	// Copy file and calculate checksum
	hasher := md5.New()
	multiWriter := io.MultiWriter(dst, hasher)
	
	if _, err := io.Copy(multiWriter, file); err != nil {
		os.Remove(filePath) // Cleanup on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	
	// Get API key info for author
	apiKey := c.MustGet("api_key").(models.APIKey)
	author := fmt.Sprintf("api_key:%s", apiKey.Name)
	
	// Generate URLs
	publicURL := ""
	privateURL := fmt.Sprintf("/p/%s/api/storage/%s/files/%s", project.Slug, bucketName, fileID)
	
	if bucket.IsPublic {
		publicURL = fmt.Sprintf("/storage/%s/%s/%s", project.Slug, bucketName, fileName)
	}
	
	// Create file record
	fileRecord := models.File{
		ID:           fileID,
		OriginalName: header.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		MimeType:     header.Header.Get("Content-Type"),
		Size:         header.Size,
		Checksum:     checksum,
		BucketName:   bucketName,
		ProjectID:    project.ID,
		IsPublic:     bucket.IsPublic,
		Author:       author,
		PublicURL:    publicURL,
		PrivateURL:   privateURL,
	}
	
	if err := h.db.Create(&fileRecord).Error; err != nil {
		os.Remove(filePath) // Cleanup on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file record"})
		return
	}
	
	// Update bucket statistics
	h.updateBucketStats(project.ID, bucketName)
	
	c.JSON(http.StatusCreated, fileRecord)
}

// GetFile downloads a file
func (h *StorageHandler) GetFile(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	fileID := c.Param("file_id")
	
	var file models.File
	if err := h.db.Where("project_id = ? AND bucket_name = ? AND id = ?", 
		project.ID, bucketName, fileID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Check if file exists on disk
	if _, err := os.Stat(file.FilePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found on disk"})
		return
	}
	
	// Set headers for file download
	c.Header("Content-Type", file.MimeType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.OriginalName))
	c.Header("Content-Length", fmt.Sprintf("%d", file.Size))
	
	c.File(file.FilePath)
}

// DeleteFile deletes a file
func (h *StorageHandler) DeleteFile(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	fileID := c.Param("file_id")
	
	var file models.File
	if err := h.db.Where("project_id = ? AND bucket_name = ? AND id = ?", 
		project.ID, bucketName, fileID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Delete file from database
	if err := h.db.Delete(&file).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete file record"})
		return
	}
	
	// Delete file from disk
	os.Remove(file.FilePath)
	
	// Update bucket statistics
	h.updateBucketStats(project.ID, bucketName)
	
	c.JSON(http.StatusOK, gin.H{"message": "File deleted successfully"})
}

// Helper functions

// bucketExists checks if a bucket exists
func (h *StorageHandler) bucketExists(projectID uint, bucketName string) bool {
	var bucket models.Bucket
	err := h.db.Where("project_id = ? AND name = ?", projectID, bucketName).First(&bucket).Error
	return err == nil
}

// updateBucketStats updates bucket statistics
func (h *StorageHandler) updateBucketStats(projectID uint, bucketName string) {
	var count int64
	var totalSize int64
	
	// Only count non-deleted files
	h.db.Model(&models.File{}).Where("project_id = ? AND bucket_name = ? AND deleted_at IS NULL", projectID, bucketName).Count(&count)
	h.db.Model(&models.File{}).Where("project_id = ? AND bucket_name = ? AND deleted_at IS NULL", projectID, bucketName).
		Select("COALESCE(SUM(size), 0)").Scan(&totalSize)
	
	h.db.Model(&models.Bucket{}).
		Where("project_id = ? AND name = ?", projectID, bucketName).
		Updates(models.Bucket{
			FileCount:    count,
			TotalSize:    totalSize,
			LastModified: time.Now(),
		})
}

// isValidBucketName validates bucket name
func isValidBucketName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	
	// Only allow letters, numbers, hyphens, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '-' || char == '_') {
			return false
		}
	}
	
	return true
}

// sanitizeFileName removes problematic characters from filename
func sanitizeFileName(filename string) string {
	// Replace spaces with underscores and remove problematic characters
	filename = strings.ReplaceAll(filename, " ", "_")
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	return filename
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}