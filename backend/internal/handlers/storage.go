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
	bucketPath := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), req.Name)
	if err := os.MkdirAll(bucketPath, 0755); err != nil {
		log.Printf("Failed to create bucket directory '%s': %v", bucketPath, err)
		// Don't fail the request if directory creation fails - it might already exist
		// or the filesystem might be read-only in some deployments
		log.Printf("Warning: Bucket created in database but directory creation failed")
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

// UpdateBucket updates bucket settings
func (h *StorageHandler) UpdateBucket(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	var req struct {
		Description  *string  `json:"description"`
		MaxFileSize  *int64   `json:"max_file_size"`
		AllowedTypes []string `json:"allowed_types"`
		IsPublic     *bool    `json:"is_public"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find bucket
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Update fields individually to avoid JSONB issues
	if req.Description != nil {
		bucket.Description = *req.Description
	}
	
	if req.MaxFileSize != nil {
		if *req.MaxFileSize <= 0 || *req.MaxFileSize > 1073741824 { // Max 1GB
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid file size limit (must be between 1 byte and 1GB)"})
			return
		}
		bucket.MaxFileSize = *req.MaxFileSize
	}
	
	if req.AllowedTypes != nil {
		bucket.AllowedTypes = req.AllowedTypes
	}
	
	if req.IsPublic != nil {
		bucket.IsPublic = *req.IsPublic
	}
	
	// Save the updated bucket
	if err := h.db.Save(&bucket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bucket"})
		return
	}
	
	// Return updated bucket
	h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket)
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
	bucketPath := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), bucketName)
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
	path := c.Query("path") // Optional path parameter for folder filtering
	
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
	
	// Clean and validate path if provided
	if path != "" {
		path = filepath.Clean(path)
		if strings.Contains(path, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
			return
		}
	}
	
	var files []models.File
	query := h.db.Where("project_id = ? AND bucket_name = ? AND folder_path = ?", project.ID, bucketName, path).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	if err := query.Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}
	
	// Get total count
	var total int64
	h.db.Model(&models.File{}).Where("project_id = ? AND bucket_name = ? AND folder_path = ?", project.ID, bucketName, path).Count(&total)
	
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
	
	// Get optional path parameter for folder uploads
	uploadPath := c.PostForm("path")
	if uploadPath != "" {
		// Clean and validate path
		uploadPath = filepath.Clean(uploadPath)
		if strings.Contains(uploadPath, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid upload path"})
			return
		}
	}
	
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
		
		// Create list of potential MIME types to check
		var mimeTypesToCheck []string
		mimeTypesToCheck = append(mimeTypesToCheck, actualMimeType)
		if extMimeType != "" {
			mimeTypesToCheck = append(mimeTypesToCheck, extMimeType)
		}
		
		// Special handling for SVG files - they can be detected as text/xml or image/svg+xml
		if strings.Contains(strings.ToLower(header.Filename), ".svg") || 
		   strings.Contains(actualMimeType, "xml") || 
		   strings.Contains(extMimeType, "svg") {
			mimeTypesToCheck = append(mimeTypesToCheck, "image/svg+xml", "image/svg", "text/xml")
		}
		
		// Check if any of the MIME types are allowed
		isAllowed := false
		for _, mimeType := range mimeTypesToCheck {
			if contains(bucket.AllowedTypes, mimeType) {
				isAllowed = true
				break
			}
		}
		
		if !isAllowed {
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
	projectDir := filepath.Clean(strconv.Itoa(int(project.ID)))
	bucketDir := filepath.Clean(bucketName)
	
	// Prevent directory traversal
	if strings.Contains(projectDir, "..") || strings.Contains(bucketDir, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path components"})
		return
	}
	
	// Construct file path with optional upload path (folder)
	var filePath string
	if uploadPath != "" {
		filePath = filepath.Join(baseDir, projectDir, bucketDir, uploadPath, fileName)
	} else {
		filePath = filepath.Join(baseDir, projectDir, bucketDir, fileName)
	}
	
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
	
	// Get author info - could be from API key or JWT
	var author string
	if apiKeyInterface, exists := c.Get("api_key"); exists {
		if apiKey, ok := apiKeyInterface.(models.APIKey); ok {
			author = fmt.Sprintf("api_key:%s", apiKey.Name)
		} else {
			author = "api_key:unknown"
		}
	} else if userInterface, exists := c.Get("user"); exists {
		if user, ok := userInterface.(models.User); ok {
			author = fmt.Sprintf("user:%s", user.Email)
		} else {
			author = "user:unknown"
		}
	} else {
		author = "unknown"
	}
	
	// Generate URLs with proper host
	host := c.Request.Host
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	baseURL := fmt.Sprintf("%s://%s", scheme, host)
	
	publicURL := ""
	privateURL := fmt.Sprintf("%s/p/%s/api/storage/%s/files/%s", baseURL, strconv.Itoa(int(project.ID)), bucketName, fileID)
	
	if bucket.IsPublic {
		publicURL = fmt.Sprintf("%s/storage/%s/%s/%s", baseURL, strconv.Itoa(int(project.ID)), bucketName, fileName)
	}
	
	// Create file record
	fileRecord := models.File{
		ID:           fileID,
		OriginalName: header.Filename,
		FileName:     fileName,
		FilePath:     filePath,
		FolderPath:   uploadPath, // Store the folder path within bucket
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

// MoveFile moves a file to a different folder within the same bucket
func (h *StorageHandler) MoveFile(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	fileID := c.Param("file_id")

	var req struct {
		NewFolderPath string `json:"new_folder_path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Clean and validate the new path
	newPath := req.NewFolderPath
	if newPath != "" {
		newPath = filepath.Clean(newPath)
		if strings.Contains(newPath, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder path"})
			return
		}
	}

	// Find the file
	var file models.File
	if err := h.db.Where("project_id = ? AND bucket_name = ? AND id = ?", 
		project.ID, bucketName, fileID).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}

	// If the file is already in the target folder, no need to move
	if file.FolderPath == newPath {
		c.JSON(http.StatusOK, file)
		return
	}

	// Construct the new file path on disk
	baseDir := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), bucketName)
	oldFilePath := file.FilePath
	
	var newFilePath string
	if newPath != "" {
		newFilePath = filepath.Join(baseDir, newPath, file.FileName)
	} else {
		newFilePath = filepath.Join(baseDir, file.FileName)
	}

	// Ensure the new directory exists
	if newPath != "" {
		newDirPath := filepath.Join(baseDir, newPath)
		if err := os.MkdirAll(newDirPath, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create target directory"})
			return
		}
	}

	// Move the file on disk
	if err := os.Rename(oldFilePath, newFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to move file on disk"})
		return
	}

	// Update the file record in database
	file.FolderPath = newPath
	file.FilePath = newFilePath

	if err := h.db.Save(&file).Error; err != nil {
		// Try to move the file back if database update fails
		os.Rename(newFilePath, oldFilePath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update file record"})
		return
	}

	c.JSON(http.StatusOK, file)
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

// CreateFolder creates a new folder in a bucket
func (h *StorageHandler) CreateFolder(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	log.Printf("DEBUG: CreateFolder called for project %d, bucket '%s'", project.ID, bucketName)
	
	var req struct {
		Name string `json:"name" binding:"required"`
		Path string `json:"path"` // Parent path, empty for root
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("DEBUG: Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	log.Printf("DEBUG: Folder request - Name: '%s', Path: '%s'", req.Name, req.Path)
	
	// Verify bucket exists
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		log.Printf("DEBUG: Bucket not found - project_id: %d, bucket_name: '%s', error: %v", project.ID, bucketName, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	log.Printf("DEBUG: Bucket found - ID: %d, Name: '%s'", bucket.ID, bucket.Name)
	
	// Validate folder name
	if !isValidFolderName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder name. Use only letters, numbers, hyphens, underscores, and spaces"})
		return
	}
	
	// Construct folder path
	basePath := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), bucketName)
	var folderPath string
	
	if req.Path != "" {
		// Clean and validate parent path
		parentPath := filepath.Clean(req.Path)
		if strings.Contains(parentPath, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid parent path"})
			return
		}
		folderPath = filepath.Join(basePath, parentPath, req.Name)
	} else {
		folderPath = filepath.Join(basePath, req.Name)
	}
	
	// Check if folder already exists
	if _, err := os.Stat(folderPath); err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Folder already exists"})
		return
	}
	
	// Create the folder
	log.Printf("DEBUG: Creating folder at path: '%s'", folderPath)
	if err := os.MkdirAll(folderPath, 0755); err != nil {
		log.Printf("DEBUG: Failed to create folder: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create folder"})
		return
	}
	
	log.Printf("DEBUG: Folder created successfully at: '%s'", folderPath)
	
	// Return success response
	relativePath := req.Path
	if relativePath != "" {
		relativePath = filepath.Join(relativePath, req.Name)
	} else {
		relativePath = req.Name
	}
	
	c.JSON(http.StatusCreated, gin.H{
		"name": req.Name,
		"path": relativePath,
		"message": "Folder created successfully",
	})
}

// ListFolders returns all folders in a bucket path
func (h *StorageHandler) ListFolders(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	path := c.Query("path") // Optional path parameter
	
	// Verify bucket exists
	if !h.bucketExists(project.ID, bucketName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Construct directory path
	basePath := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), bucketName)
	var targetPath string
	
	if path != "" {
		// Clean and validate path
		cleanPath := filepath.Clean(path)
		if strings.Contains(cleanPath, "..") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid path"})
			return
		}
		targetPath = filepath.Join(basePath, cleanPath)
	} else {
		targetPath = basePath
	}
	
	// Read directory
	entries, err := os.ReadDir(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Path not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read directory"})
		}
		return
	}
	
	var folders []gin.H
	for _, entry := range entries {
		if entry.IsDir() {
			// Get folder info
			folderPath := filepath.Join(targetPath, entry.Name())
			info, err := entry.Info()
			if err != nil {
				continue
			}
			
			// Count files in folder (recursive)
			fileCount := countFilesInDirectory(folderPath)
			
			relativePath := entry.Name()
			if path != "" {
				relativePath = filepath.Join(path, entry.Name())
			}
			
			folders = append(folders, gin.H{
				"name":         entry.Name(),
				"path":         relativePath,
				"created_at":   info.ModTime(),
				"file_count":   fileCount,
			})
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"folders": folders,
		"path":    path,
	})
}

// DeleteFolder deletes a folder and all its contents
func (h *StorageHandler) DeleteFolder(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	folderPath := c.Query("path")
	
	if folderPath == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Folder path is required"})
		return
	}
	
	// Verify bucket exists
	if !h.bucketExists(project.ID, bucketName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Clean and validate path
	cleanPath := filepath.Clean(folderPath)
	if strings.Contains(cleanPath, "..") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder path"})
		return
	}
	
	// Construct full path
	basePath := filepath.Join("./uploads", strconv.Itoa(int(project.ID)), bucketName)
	fullPath := filepath.Join(basePath, cleanPath)
	
	// Check if folder exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	
	// Delete folder and all contents
	if err := os.RemoveAll(fullPath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete folder"})
		return
	}
	
	// Delete file records for files that were in this folder
	folderPrefix := cleanPath + "/"
	h.db.Where("project_id = ? AND bucket_name = ? AND file_path LIKE ?", 
		project.ID, bucketName, "%"+folderPrefix+"%").Delete(&models.File{})
	
	// Update bucket statistics
	h.updateBucketStats(project.ID, bucketName)
	
	c.JSON(http.StatusOK, gin.H{"message": "Folder deleted successfully"})
}

// Helper functions

// isValidFolderName validates folder name
func isValidFolderName(name string) bool {
	if name == "" || len(name) > 100 {
		return false
	}
	
	// Check for reserved names
	reservedNames := []string{".", "..", "admin", "api", "www", "config", "system"}
	for _, reserved := range reservedNames {
		if strings.EqualFold(name, reserved) {
			return false
		}
	}
	
	// Allow letters, numbers, hyphens, underscores, and spaces
	re := regexp.MustCompile(`^[a-zA-Z0-9_\- ]+$`)
	return re.MatchString(name)
}

// countFilesInDirectory counts files in a directory recursively
func countFilesInDirectory(dirPath string) int {
	count := 0
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}

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

// Admin methods for storage management via JWT admin routes

// AdminListBuckets lists buckets for admin management (JWT authenticated)
func (h *StorageHandler) AdminListBuckets(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var buckets []models.Bucket
	if err := h.db.Where("project_id = ?", project.ID).Find(&buckets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buckets"})
		return
	}
	
	c.JSON(http.StatusOK, buckets)
}

// AdminCreateBucket creates a bucket via admin interface (JWT authenticated)
func (h *StorageHandler) AdminCreateBucket(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular CreateBucket method
	h.CreateBucket(c)
}

// AdminGetBucket gets a bucket via admin interface (JWT authenticated)
func (h *StorageHandler) AdminGetBucket(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular GetBucket method
	h.GetBucket(c)
}

// AdminUpdateBucket updates a bucket via admin interface (JWT authenticated)
func (h *StorageHandler) AdminUpdateBucket(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular UpdateBucket method
	h.UpdateBucket(c)
}

// AdminDeleteBucket deletes a bucket via admin interface (JWT authenticated)
func (h *StorageHandler) AdminDeleteBucket(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular DeleteBucket method
	h.DeleteBucket(c)
}

// AdminListFiles lists files via admin interface (JWT authenticated)
func (h *StorageHandler) AdminListFiles(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular ListFiles method
	h.ListFiles(c)
}

// AdminUploadFile uploads a file via admin interface (JWT authenticated)
func (h *StorageHandler) AdminUploadFile(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular UploadFile method
	h.UploadFile(c)
}

// AdminGetFile gets a file via admin interface (JWT authenticated)
func (h *StorageHandler) AdminGetFile(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular GetFile method
	h.GetFile(c)
}

// AdminMoveFile moves a file via admin interface (JWT authenticated)
func (h *StorageHandler) AdminMoveFile(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular MoveFile method
	h.MoveFile(c)
}

// AdminDeleteFile deletes a file via admin interface (JWT authenticated)
func (h *StorageHandler) AdminDeleteFile(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular DeleteFile method
	h.DeleteFile(c)
}

// AdminListFolders lists folders via admin interface (JWT authenticated)
func (h *StorageHandler) AdminListFolders(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular ListFolders method
	h.ListFolders(c)
}

// AdminCreateFolder creates a folder via admin interface (JWT authenticated)
func (h *StorageHandler) AdminCreateFolder(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular CreateFolder method
	h.CreateFolder(c)
}

// AdminDeleteFolder deletes a folder via admin interface (JWT authenticated)
func (h *StorageHandler) AdminDeleteFolder(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular DeleteFolder method
	h.DeleteFolder(c)
}

// Admin methods for bucket visibility management

// AdminSetBucketVisibility toggles bucket public/private status (JWT authenticated)
func (h *StorageHandler) AdminSetBucketVisibility(c *gin.Context) {
	projectID := c.Param("id")
	bucketName := c.Param("bucket")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var req struct {
		IsPublic      bool   `json:"is_public"`
		ConfirmAction string `json:"confirm_action,omitempty"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Find the bucket
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Update bucket visibility
	bucket.IsPublic = req.IsPublic
	bucket.LastModified = time.Now()
	
	if err := h.db.Save(&bucket).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update bucket visibility"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"message":   fmt.Sprintf("Bucket '%s' is now %s", bucketName, map[bool]string{true: "public", false: "private"}[req.IsPublic]),
		"bucket":    bucket,
		"is_public": bucket.IsPublic,
	})
}

// AdminGetFilePublicURL generates public URL for a file (JWT authenticated)
func (h *StorageHandler) AdminGetFilePublicURL(c *gin.Context) {
	projectID := c.Param("id")
	bucketName := c.Param("bucket")
	fileID := c.Param("file_id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	var file models.File
	if err := h.db.Where("id = ? AND project_id = ? AND bucket_name = ?", fileID, project.ID, bucketName).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Generate public URL
	var publicURL string
	if bucket.IsPublic {
		publicURL = fmt.Sprintf("%s/public/%s/%s/%s", h.cfg.BaseURL, project.Slug, bucketName, file.FilePath)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"public_url": publicURL,
		"is_public":  bucket.IsPublic,
		"file":       file,
		"bucket":     bucket,
		"shareable_url": publicURL, // Same for now, could be different for signed URLs
	})
}

// AdminListPublicBuckets lists all public buckets for a project (JWT authenticated)
func (h *StorageHandler) AdminListPublicBuckets(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var buckets []models.Bucket
	if err := h.db.Where("project_id = ? AND is_public = ?", project.ID, true).Find(&buckets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch public buckets"})
		return
	}
	
	c.JSON(http.StatusOK, buckets)
}

// Project API methods for connected apps (API key authentication)

// GetFilePublicURL generates public URL for a file in project API (API key authenticated)
func (h *StorageHandler) GetFilePublicURL(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	fileID := c.Param("file_id")
	
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	var file models.File
	if err := h.db.Where("id = ? AND project_id = ? AND bucket_name = ?", fileID, project.ID, bucketName).First(&file).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "File not found"})
		return
	}
	
	// Generate public URL
	var publicURL string
	if bucket.IsPublic {
		publicURL = fmt.Sprintf("%s/public/%s/%s/%s", h.cfg.BaseURL, project.Slug, bucketName, file.FilePath)
	}
	
	c.JSON(http.StatusOK, gin.H{
		"public_url": publicURL,
		"is_public":  bucket.IsPublic,
		"file":       file,
		"bucket":     bucket,
		"shareable_url": publicURL,
	})
}

// GetBatchFilePublicURLs generates public URLs for multiple files (API key authenticated)
func (h *StorageHandler) GetBatchFilePublicURLs(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	bucketName := c.Param("bucket")
	
	var req struct {
		FileIDs []string `json:"file_ids"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}
	
	// Check if bucket exists and is public
	var bucket models.Bucket
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, bucketName).First(&bucket).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Bucket not found"})
		return
	}
	
	// Get files
	var files []models.File
	if err := h.db.Where("id IN ? AND project_id = ? AND bucket_name = ?", req.FileIDs, project.ID, bucketName).Find(&files).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch files"})
		return
	}
	
	// Generate public URLs
	publicURLs := make(map[string]string)
	for _, file := range files {
		if bucket.IsPublic {
			publicURLs[file.ID] = fmt.Sprintf("%s/public/%s/%s/%s", h.cfg.BaseURL, project.Slug, bucketName, file.FilePath)
		} else {
			publicURLs[file.ID] = "" // Empty string for private buckets
		}
	}
	
	c.JSON(http.StatusOK, gin.H{
		"public_urls": publicURLs,
		"is_public":   bucket.IsPublic,
		"bucket":      bucket,
		"file_count":  len(files),
	})
}