package handlers

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	
	// Set defaults with validation
	maxFileSize := int64(52428800) // 50MB default
	if req.MaxFileSize != nil {
		if *req.MaxFileSize <= 0 || *req.MaxFileSize > 1073741824 { // Max 1GB
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file size limit (must be between 1 byte and 1GB)"})
			return
		}
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
	
	// Validate MIME type with content inspection (not just header)
	if len(bucket.AllowedTypes) > 0 {
		// Read first 512 bytes to detect actual content type
		buffer := make([]byte, 512)
		n, _ := file.Read(buffer)
		file.Seek(0, 0) // Reset file position
		
		// Detect MIME type from content
		actualMimeType := http.DetectContentType(buffer[:n])
		
		// Also check file extension
		extension := filepath.Ext(header.Filename)
		extMimeType := mime.TypeByExtension(extension)
		
		// Validate against both content-based and extension-based MIME types
		if !contains(bucket.AllowedTypes, actualMimeType) && !contains(bucket.AllowedTypes, extMimeType) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": fmt.Sprintf("File type not allowed. Detected: %s, Allowed: %v", actualMimeType, bucket.AllowedTypes),
			})
			return
		}
	}
	
	// Generate secure file info
	fileID := uuid.New().String()
	safeOriginalName := sanitizeFileName(header.Filename)
	
	// Use only the file ID and extension for storage (prevents any filename attacks)
	extension := filepath.Ext(safeOriginalName)
	if extension == "" {
		// Try to determine extension from MIME type if not present
		if exts, err := mime.ExtensionsByType(http.DetectContentType(make([]byte, 512))); err == nil && len(exts) > 0 {
			extension = exts[0]
		}
	}
	
	fileName := fmt.Sprintf("%s%s", fileID, extension)
	
	// Secure file path construction with validation
	baseDir := filepath.Clean("./uploads")
	projectDir := filepath.Clean(project.Slug)
	bucketDir := filepath.Clean(bucketName)
	
	// Prevent directory traversal
	if strings.Contains(projectDir, "..") || strings.Contains(bucketDir, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path components"})
		return
	}
	
	filePath := filepath.Join(baseDir, projectDir, bucketDir, fileName)
	
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
	
	// Copy file with size limit and calculate secure checksum
	hasher := sha256.New() // Use SHA256 instead of MD5 for better security
	multiWriter := io.MultiWriter(dst, hasher)
	
	// Limit copy size to prevent resource exhaustion
	limitedReader := io.LimitReader(file, bucket.MaxFileSize+1) // +1 to detect size exceeded
	
	bytesWritten, err := io.Copy(multiWriter, limitedReader)
	if err != nil {
		os.Remove(filePath) // Cleanup on error
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save file"})
		return
	}
	
	// Double-check file size after writing
	if bytesWritten > bucket.MaxFileSize {
		os.Remove(filePath) // Cleanup oversized file
		c.JSON(http.StatusBadRequest, gin.H{"error": "File size exceeded limit during upload"})
		return
	}
	
	checksum := fmt.Sprintf("%x", hasher.Sum(nil))
	
	// Verify file integrity
	if bytesWritten != header.Size {
		log.Printf("File size mismatch: expected %d, got %d", header.Size, bytesWritten)
		os.Remove(filePath)
		c.JSON(http.StatusBadRequest, gin.H{"error": "File upload integrity check failed"})
		return
	}
	
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

// isValidBucketName validates bucket name with security checks
func isValidBucketName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	
	// Check for reserved names
	reservedNames := []string{"admin", "api", "www", "mail", "ftp", "root", "public", "private", "system", "config", "uploads"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return false
		}
	}
	
	// Only allow letters, numbers, hyphens, and underscores (no dots to prevent subdomain issues)
	re := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	return re.MatchString(name)
}


// sanitizeFileName removes problematic characters from filename with comprehensive security
func sanitizeFileName(filename string) string {
	// Remove path separators and dangerous characters
	filename = strings.ReplaceAll(filename, "/", "_")
	filename = strings.ReplaceAll(filename, "\\", "_")
	filename = strings.ReplaceAll(filename, "..", "_")
	filename = strings.ReplaceAll(filename, ":", "_")
	filename = strings.ReplaceAll(filename, "*", "_")
	filename = strings.ReplaceAll(filename, "?", "_")
	filename = strings.ReplaceAll(filename, "\"", "_")
	filename = strings.ReplaceAll(filename, "<", "_")
	filename = strings.ReplaceAll(filename, ">", "_")
	filename = strings.ReplaceAll(filename, "|", "_")
	filename = strings.ReplaceAll(filename, " ", "_")
	
	// Remove control characters and non-printable characters
	re := regexp.MustCompile(`[\x00-\x1f\x7f-\x9f]`)
	filename = re.ReplaceAllString(filename, "_")
	
	// Limit filename length
	if len(filename) > 100 {
		ext := filepath.Ext(filename)
		name := filename[:100-len(ext)]
		filename = name + ext
	}
	
	// Ensure filename is not empty or dangerous
	if filename == "" || filename == "." || filename == ".." {
		filename = "unnamed_file"
	}
	
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