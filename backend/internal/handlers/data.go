package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// DataHandler handles data API requests (collections and documents)
type DataHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewDataHandler creates a new data handler
func NewDataHandler(db *gorm.DB, cfg *config.Config) *DataHandler {
	return &DataHandler{db: db, cfg: cfg}
}

// Collection Management

// ListCollections returns all collections for a project
func (h *DataHandler) ListCollections(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var collections []models.Collection
	if err := h.db.Where("project_id = ?", project.ID).Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
		return
	}
	
	c.JSON(http.StatusOK, collections)
}

// CreateCollection creates a new collection
func (h *DataHandler) CreateCollection(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var req struct {
		Name        string   `json:"name" binding:"required"`
		Description string   `json:"description"`
		Schema      []string `json:"schema"`
		Indexes     []string `json:"indexes"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate collection name (no spaces, special chars)
	if !isValidCollectionName(req.Name) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid collection name. Use only letters, numbers, and underscores"})
		return
	}
	
	// Check if collection already exists
	var existingCollection models.Collection
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, req.Name).First(&existingCollection).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Collection already exists"})
		return
	}
	
	// Create collection
	collection := models.Collection{
		Name:          req.Name,
		Description:   req.Description,
		Schema:        req.Schema,
		Indexes:       req.Indexes,
		ProjectID:     project.ID,
		DocumentCount: 0,
		LastModified:  time.Now(),
	}
	
	if err := h.db.Create(&collection).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create collection"})
		return
	}
	
	c.JSON(http.StatusCreated, collection)
}

// GetCollection returns a specific collection
func (h *DataHandler) GetCollection(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	
	var collection models.Collection
	if err := h.db.Where("project_id = ? AND name = ?", project.ID, collectionName).First(&collection).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}
	
	c.JSON(http.StatusOK, collection)
}

// DeleteCollection deletes a collection and all its documents
func (h *DataHandler) DeleteCollection(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Delete all documents in collection
	if err := tx.Where("project_id = ? AND collection_name = ?", project.ID, collectionName).Delete(&models.Document{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete documents"})
		return
	}
	
	// Delete collection
	result := tx.Where("project_id = ? AND name = ?", project.ID, collectionName).Delete(&models.Collection{})
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete collection"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}
	
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Collection deleted successfully"})
}

// Document Management

// ListDocuments returns all documents in a collection
func (h *DataHandler) ListDocuments(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	
	// Verify collection exists
	if !h.collectionExists(project.ID, collectionName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
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
		if strings.Contains(order, "created_at") || strings.Contains(order, "updated_at") {
			orderBy = order
		}
	}
	
	var documents []models.Document
	query := h.db.Where("project_id = ? AND collection_name = ?", project.ID, collectionName).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	// Add simple filtering if provided
	if filter := c.Query("filter"); filter != "" {
		// Simple JSON filtering - in production, implement proper query language
		query = query.Where("data::text ILIKE ?", "%"+filter+"%")
	}
	
	if err := query.Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}
	
	// Get total count
	var total int64
	h.db.Model(&models.Document{}).Where("project_id = ? AND collection_name = ?", project.ID, collectionName).Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"documents": documents,
		"total":     total,
		"limit":     limit,
		"offset":    offset,
	})
}

// CreateDocument creates a new document in a collection
func (h *DataHandler) CreateDocument(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	
	// Verify collection exists
	if !h.collectionExists(project.ID, collectionName) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Collection not found"})
		return
	}
	
	// Parse request body as JSON
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}
	
	// Generate ID if not provided
	docID := uuid.New().String()
	if id, exists := data["id"]; exists {
		if idStr, ok := id.(string); ok && idStr != "" {
			docID = idStr
		}
	}
	
	// Remove id from data if it exists (stored separately)
	delete(data, "id")
	
	// Get API key info for author
	apiKey := c.MustGet("api_key").(models.APIKey)
	author := fmt.Sprintf("api_key:%s", apiKey.Name)
	
	// Create document
	document := models.Document{
		ID:             docID,
		CollectionName: collectionName,
		ProjectID:      project.ID,
		Data:           data,
		Version:        1,
		Author:         author,
	}
	
	if err := h.db.Create(&document).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Document with this ID already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document"})
		return
	}
	
	// Update collection stats
	h.updateCollectionStats(project.ID, collectionName)
	
	c.JSON(http.StatusCreated, document)
}

// GetDocument returns a specific document
func (h *DataHandler) GetDocument(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	documentID := c.Param("id")
	
	var document models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", 
		project.ID, collectionName, documentID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	
	c.JSON(http.StatusOK, document)
}

// UpdateDocument updates a document
func (h *DataHandler) UpdateDocument(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	documentID := c.Param("id")
	
	// Parse request body
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}
	
	// Remove id from data if it exists
	delete(data, "id")
	
	// Get API key info for author
	apiKey := c.MustGet("api_key").(models.APIKey)
	author := fmt.Sprintf("api_key:%s", apiKey.Name)
	
	// Update document
	result := h.db.Model(&models.Document{}).
		Where("project_id = ? AND collection_name = ? AND id = ?", project.ID, collectionName, documentID).
		Updates(models.Document{
			Data:   data,
			Author: author,
		}).
		Update("version", gorm.Expr("version + 1"))
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	
	// Get updated document
	var document models.Document
	h.db.Where("project_id = ? AND collection_name = ? AND id = ?", 
		project.ID, collectionName, documentID).First(&document)
	
	// Update collection stats
	h.updateCollectionStats(project.ID, collectionName)
	
	c.JSON(http.StatusOK, document)
}

// DeleteDocument deletes a document
func (h *DataHandler) DeleteDocument(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	collectionName := c.Param("collection")
	documentID := c.Param("id")
	
	result := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", 
		project.ID, collectionName, documentID).Delete(&models.Document{})
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete document"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	
	// Update collection stats
	h.updateCollectionStats(project.ID, collectionName)
	
	c.JSON(http.StatusOK, gin.H{"message": "Document deleted successfully"})
}

// Helper functions

// collectionExists checks if a collection exists
func (h *DataHandler) collectionExists(projectID uint, collectionName string) bool {
	var collection models.Collection
	err := h.db.Where("project_id = ? AND name = ?", projectID, collectionName).First(&collection).Error
	return err == nil
}

// updateCollectionStats updates collection statistics
func (h *DataHandler) updateCollectionStats(projectID uint, collectionName string) {
	var count int64
	h.db.Model(&models.Document{}).Where("project_id = ? AND collection_name = ?", projectID, collectionName).Count(&count)
	
	h.db.Model(&models.Collection{}).
		Where("project_id = ? AND name = ?", projectID, collectionName).
		Updates(models.Collection{
			DocumentCount: count,
			LastModified:  time.Now(),
		})
}

// isValidCollectionName validates collection name
func isValidCollectionName(name string) bool {
	if name == "" || len(name) > 50 {
		return false
	}
	
	// Only allow letters, numbers, and underscores
	for _, char := range name {
		if !((char >= 'a' && char <= 'z') || (char >= 'A' && char <= 'Z') || 
			 (char >= '0' && char <= '9') || char == '_') {
			return false
		}
	}
	
	return true
}