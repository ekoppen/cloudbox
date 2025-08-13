package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
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
		// Strict validation for order by to prevent SQL injection
		validOrderPattern := regexp.MustCompile(`^(created_at|updated_at)\s+(ASC|DESC)$|^(created_at|updated_at)$`)
		if validOrderPattern.MatchString(strings.TrimSpace(order)) {
			orderBy = strings.TrimSpace(order)
		}
	}
	
	var documents []models.Document
	query := h.db.Where("project_id = ? AND collection_name = ?", project.ID, collectionName).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	// Add secure JSON filtering if provided
	if filter := c.Query("filter"); filter != "" {
		// Validate filter is valid JSON to prevent injection
		var filterJSON map[string]interface{}
		if err := json.Unmarshal([]byte(filter), &filterJSON); err == nil {
			// Use parameterized query with JSON containment operator
			query = query.Where("data @> ?", filter)
		} else {
			// Fallback to safe text search with proper escaping
			safeFilter := strings.ReplaceAll(filter, "'", "''")
			query = query.Where("data::text ILIKE ?", "%"+safeFilter+"%")
		}
	}
	
	if err := query.Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch documents"})
		return
	}
	
	// Get total count efficiently in single query
	var total int64
	countQuery := h.db.Model(&models.Document{}).Where("project_id = ? AND collection_name = ?", project.ID, collectionName)
	
	// Apply same filter to count query
	if filter := c.Query("filter"); filter != "" {
		var filterJSON map[string]interface{}
		if err := json.Unmarshal([]byte(filter), &filterJSON); err == nil {
			countQuery = countQuery.Where("data @> ?", filter)
		} else {
			safeFilter := strings.ReplaceAll(filter, "'", "''")
			countQuery = countQuery.Where("data::text ILIKE ?", "%"+safeFilter+"%")
		}
	}
	
	countQuery.Count(&total)
	
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
	
	// Parse and validate request body as JSON
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}
	
	// Validate document size (prevent DOS attacks)
	if len(data) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document too large (max 100 fields)"})
		return
	}
	
	// Sanitize document data (remove potentially dangerous fields)
	delete(data, "_sql")
	delete(data, "__proto__")
	delete(data, "constructor")
	
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
	
	// Use transaction for document creation
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in CreateDocument: %v", r)
		}
	}()
	
	if err := tx.Create(&document).Error; err != nil {
		tx.Rollback()
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Document with this ID already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create document"})
		return
	}
	
	// Update collection stats in same transaction
	if err := h.updateCollectionStatsInTx(tx, project.ID, collectionName); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection stats"})
		return
	}
	
	tx.Commit()
	
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
	
	// Parse and validate request body
	var data map[string]interface{}
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}
	
	// Validate document size
	if len(data) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Document too large (max 100 fields)"})
		return
	}
	
	// Sanitize document data
	delete(data, "id")
	delete(data, "_sql")
	delete(data, "__proto__")
	delete(data, "constructor")
	
	// Get API key info for author
	apiKey := c.MustGet("api_key").(models.APIKey)
	author := fmt.Sprintf("api_key:%s", apiKey.Name)
	
	// Update document with transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in UpdateDocument: %v", r)
		}
	}()
	
	result := tx.Model(&models.Document{}).
		Where("project_id = ? AND collection_name = ? AND id = ?", project.ID, collectionName, documentID).
		Updates(models.Document{
			Data:   data,
			Author: author,
		}).
		Update("version", gorm.Expr("version + 1"))
	
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update document"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Document not found"})
		return
	}
	
	// Get updated document
	var document models.Document
	tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		project.ID, collectionName, documentID).First(&document)
	
	// Update collection stats in same transaction
	if err := h.updateCollectionStatsInTx(tx, project.ID, collectionName); err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update collection stats"})
		return
	}
	
	tx.Commit()
	
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

// updateCollectionStatsInTx updates collection statistics within a transaction
func (h *DataHandler) updateCollectionStatsInTx(tx *gorm.DB, projectID uint, collectionName string) error {
	var count int64
	if err := tx.Model(&models.Document{}).Where("project_id = ? AND collection_name = ?", projectID, collectionName).Count(&count).Error; err != nil {
		return err
	}
	
	return tx.Model(&models.Collection{}).
		Where("project_id = ? AND name = ?", projectID, collectionName).
		Updates(models.Collection{
			DocumentCount: count,
			LastModified:  time.Now(),
		}).Error
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

// Admin methods for collections management via JWT admin routes

// AdminListCollections lists collections for admin management (JWT authenticated)
func (h *DataHandler) AdminListCollections(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	var collections []models.Collection
	if err := h.db.Where("project_id = ?", project.ID).Find(&collections).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch collections"})
		return
	}
	
	c.JSON(http.StatusOK, collections)
}

// AdminCreateCollection creates a collection via admin interface (JWT authenticated)
func (h *DataHandler) AdminCreateCollection(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular CreateCollection method
	h.CreateCollection(c)
}

// AdminGetCollection gets a collection via admin interface (JWT authenticated)
func (h *DataHandler) AdminGetCollection(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular GetCollection method
	h.GetCollection(c)
}

// AdminDeleteCollection deletes a collection via admin interface (JWT authenticated)
func (h *DataHandler) AdminDeleteCollection(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular DeleteCollection method
	h.DeleteCollection(c)
}

// AdminListDocuments lists documents via admin interface (JWT authenticated)
func (h *DataHandler) AdminListDocuments(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular ListDocuments method
	h.ListDocuments(c)
}

// AdminCreateDocument creates a document via admin interface (JWT authenticated)
func (h *DataHandler) AdminCreateDocument(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular CreateDocument method
	h.CreateDocument(c)
}

// AdminGetDocument gets a document via admin interface (JWT authenticated)
func (h *DataHandler) AdminGetDocument(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular GetDocument method
	h.GetDocument(c)
}

// AdminUpdateDocument updates a document via admin interface (JWT authenticated)
func (h *DataHandler) AdminUpdateDocument(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular UpdateDocument method
	h.UpdateDocument(c)
}

// AdminDeleteDocument deletes a document via admin interface (JWT authenticated)
func (h *DataHandler) AdminDeleteDocument(c *gin.Context) {
	projectID := c.Param("id")
	
	var project models.Project
	if err := h.db.Where("id = ?", projectID).First(&project).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}
	
	// Set the project in context for the regular handler
	c.Set("project", project)
	
	// Call the regular DeleteDocument method
	h.DeleteDocument(c)
}