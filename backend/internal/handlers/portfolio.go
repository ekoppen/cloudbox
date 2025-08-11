package handlers

import (
	"fmt"
	"log"
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

type PortfolioHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewPortfolioHandler(db *gorm.DB, cfg *config.Config) *PortfolioHandler {
	return &PortfolioHandler{
		db:  db,
		cfg: cfg,
	}
}

// Translation endpoints
func (h *PortfolioHandler) GetLanguages(c *gin.Context) {
	// Return default languages for now
	languages := []string{"en", "nl", "de", "fr"}
	c.JSON(http.StatusOK, languages)
}

func (h *PortfolioHandler) SetLanguages(c *gin.Context) {
	var req struct {
		Languages []string `json:"languages"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// For now, just acknowledge the request
	c.JSON(http.StatusOK, gin.H{"message": "Languages updated", "languages": req.Languages})
}

func (h *PortfolioHandler) TranslatePage(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	pageID := c.Param("pageId")
	
	var req struct {
		TargetLanguage string `json:"targetLanguage"`
		OpenAIAPIKey   string `json:"openaiApiKey"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in TranslatePage: %v", r)
		}
	}()
	
	// Verify the page exists
	var page models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "pages", pageID).First(&page).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify page"})
		}
		return
	}
	
	// Check if translation already exists
	translationID := pageID + "_" + req.TargetLanguage
	var existingTranslation models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "translations", translationID).First(&existingTranslation).Error; err == nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Translation already exists for this language"})
		return
	}
	
	// Create translation document
	translationData := map[string]interface{}{
		"pageId":     pageID,
		"language":   req.TargetLanguage,
		"status":     "completed",
		"content":    page.Data, // Copy original content for now (would be translated by AI service)
		"sourceLanguage": "en", // Assume original is English
	}
	
	translation := models.Document{
		ID:             translationID,
		CollectionName: "translations",
		ProjectID:      projectID,
		Data:           translationData,
		Version:        1,
		Author:         "translation_service",
	}
	
	if err := tx.Create(&translation).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create translation"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit translation"})
		return
	}
	
	// Return translation response
	response := gin.H{
		"id":        translation.ID,
		"pageId":    pageID,
		"language":  req.TargetLanguage,
		"status":    "completed",
		"createdAt": translation.CreatedAt,
	}
	
	// Merge translation data
	for key, value := range translation.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusOK, response)
}

func (h *PortfolioHandler) GetPageTranslations(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	pageID := c.Param("pageId")
	
	// Verify the page exists
	var page models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "pages", pageID).First(&page).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify page"})
		}
		return
	}
	
	// Query translation documents for this page
	var documents []models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND data->>'pageId' = ?", 
		projectID, "translations", pageID).Order("created_at DESC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch translations"})
		return
	}
	
	// Convert documents to translation format
	translations := make([]gin.H, 0, len(documents))
	for _, doc := range documents {
		translation := gin.H{
			"id":        doc.ID,
			"pageId":    pageID,
			"createdAt": doc.CreatedAt,
			"updatedAt": doc.UpdatedAt,
		}
		
		// Merge document data
		for key, value := range doc.Data {
			translation[key] = value
		}
		
		translations = append(translations, translation)
	}
	
	c.JSON(http.StatusOK, translations)
}

func (h *PortfolioHandler) DeleteTranslation(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	translationID := c.Param("translationId")
	if translationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Translation ID is required"})
		return
	}
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in DeleteTranslation: %v", r)
		}
	}()
	
	// Verify translation exists and get its data before deletion
	var translation models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "translations", translationID).First(&translation).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify translation"})
		}
		return
	}
	
	// Delete the translation
	result := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "translations", translationID).Delete(&models.Document{})
	
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete translation"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Translation not found"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deletion"})
		return
	}
	
	log.Printf("Deleted translation %s for project %d", translationID, projectID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Translation deleted successfully",
		"id": translationID,
	})
}

// Analytics endpoints
func (h *PortfolioHandler) GetAnalytics(c *gin.Context) {
	daysStr := c.DefaultQuery("days", "14")
	days, err := strconv.Atoi(daysStr)
	if err != nil {
		days = 14
	}
	
	// Generate mock analytics data
	var analyticsData []gin.H
	var topPages []gin.H
	
	now := time.Now()
	for i := 0; i < days; i++ {
		date := now.AddDate(0, 0, -i).Format("2006-01-02")
		visitors := 50 + (i * 3) // Mock increasing visitors
		uniqueVisitors := int(float64(visitors) * 0.8)
		
		analyticsData = append([]gin.H{{
			"date": date,
			"visitors": visitors,
			"uniqueVisitors": uniqueVisitors,
		}}, analyticsData...)
	}
	
	// Mock top pages
	topPages = []gin.H{
		{"name": "Home", "path": "/", "views": 245},
		{"name": "Portfolio", "path": "/portfolio", "views": 189},
		{"name": "About", "path": "/about", "views": 156},
		{"name": "Contact", "path": "/contact", "views": 98},
	}
	
	c.JSON(http.StatusOK, gin.H{
		"analyticsData": analyticsData,
		"topPages": topPages,
	})
}

// Images endpoints
func (h *PortfolioHandler) GetImages(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Query documents from images collection
	var documents []models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ?", projectID, "images").Order("created_at DESC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch images"})
		return
	}

	// Convert documents to images format
	images := make([]gin.H, 0, len(documents))
	for _, doc := range documents {
		image := gin.H{
			"_id": doc.ID,
			"id": doc.ID,
			"createdAt": doc.CreatedAt,
			"updatedAt": doc.UpdatedAt,
		}
		
		// Merge document data
		for key, value := range doc.Data {
			image[key] = value
		}
		
		images = append(images, image)
	}
	
	c.JSON(http.StatusOK, images)
}

// DeleteImage deletes an image document
func (h *PortfolioHandler) DeleteImage(c *gin.Context) {
	projectID, valid := h.validateProjectAccess(c)
	if !valid {
		return
	}

	imageID := c.Param("id")
	if imageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Image ID is required"})
		return
	}
	
	// Use utility function for deletion
	err := h.deleteDocumentWithCascade(projectID, "images", imageID, nil)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete image"})
		}
		return
	}
	
	log.Printf("Deleted image %s for project %d", imageID, projectID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Image deleted successfully",
		"id": imageID,
	})
}

func (h *PortfolioHandler) UpdateImage(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	imageID := c.Param("id")
	
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find the existing document
	var document models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "images", imageID).First(&document).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Image not found"})
		return
	}

	// Update document data
	for key, value := range req {
		document.Data[key] = value
	}
	document.Version++

	if err := h.db.Save(&document).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update image"})
		return
	}

	// Return updated document
	response := gin.H{
		"_id": document.ID,
		"id": document.ID,
		"createdAt": document.CreatedAt,
		"updatedAt": document.UpdatedAt,
	}
	
	// Merge document data
	for key, value := range document.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusOK, response)
}

// Albums endpoints
func (h *PortfolioHandler) GetAlbums(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Query documents from albums collection
	var documents []models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ?", projectID, "albums").Order("data->>'order' ASC, created_at DESC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch albums"})
		return
	}

	// Convert documents to albums format
	albums := make([]gin.H, 0, len(documents))
	for _, doc := range documents {
		album := gin.H{
			"_id": doc.ID,
			"id": doc.ID,
			"createdAt": doc.CreatedAt,
			"updatedAt": doc.UpdatedAt,
		}
		
		// Merge document data
		for key, value := range doc.Data {
			album[key] = value
		}
		
		albums = append(albums, album)
	}
	
	c.JSON(http.StatusOK, albums)
}

// CreateAlbum creates a new album document
func (h *PortfolioHandler) CreateAlbum(c *gin.Context) {
	projectID, valid := h.validateProjectAccess(c)
	if !valid {
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Generate unique album ID if not provided
	albumID, ok := req["id"].(string)
	if !ok || albumID == "" {
		albumID = uuid.New().String()
	}
	
	// Set default values
	if _, exists := req["title"]; !exists {
		req["title"] = "Untitled Album"
	}
	if _, exists := req["published"]; !exists {
		req["published"] = false
	}
	if _, exists := req["order"]; !exists {
		req["order"] = 0
	}
	
	album, err := h.createDocumentWithValidation(projectID, "albums", albumID, req, "portfolio_admin")
	if err != nil {
		if strings.Contains(err.Error(), "already exists") {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create album"})
		}
		return
	}
	
	log.Printf("Created album %s for project %d", albumID, projectID)
	
	// Return created album
	response := gin.H{
		"_id": album.ID,
		"id": album.ID,
		"createdAt": album.CreatedAt,
		"updatedAt": album.UpdatedAt,
		"message": "Album created successfully",
	}
	
	// Merge album data
	for key, value := range album.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusCreated, response)
}

// UpdateAlbum updates an existing album document
func (h *PortfolioHandler) UpdateAlbum(c *gin.Context) {
	projectID, valid := h.validateProjectAccess(c)
	if !valid {
		return
	}

	albumID := c.Param("id")
	if albumID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Album ID is required"})
		return
	}
	
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	album, err := h.updateDocumentWithValidation(projectID, "albums", albumID, req)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update album"})
		}
		return
	}
	
	log.Printf("Updated album %s for project %d", albumID, projectID)
	
	// Return updated album
	response := gin.H{
		"_id": album.ID,
		"id": album.ID,
		"createdAt": album.CreatedAt,
		"updatedAt": album.UpdatedAt,
		"message": "Album updated successfully",
	}
	
	// Merge album data
	for key, value := range album.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusOK, response)
}

// DeleteAlbum deletes an album document
func (h *PortfolioHandler) DeleteAlbum(c *gin.Context) {
	projectID, valid := h.validateProjectAccess(c)
	if !valid {
		return
	}

	albumID := c.Param("id")
	if albumID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Album ID is required"})
		return
	}
	
	// Use utility function for deletion
	err := h.deleteDocumentWithCascade(projectID, "albums", albumID, nil)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Album not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete album"})
		}
		return
	}
	
	log.Printf("Deleted album %s for project %d", albumID, projectID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Album deleted successfully",
		"id": albumID,
	})
}

// DeletePage deletes a page and all its related translations (cascade deletion)
func (h *PortfolioHandler) DeletePage(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page ID is required"})
		return
	}
	
	// Start transaction for cascade deletion
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in DeletePage: %v", r)
		}
	}()
	
	// Verify page exists
	var page models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "pages", pageID).First(&page).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify page"})
		}
		return
	}
	
	// First, delete all related translations (cascade deletion)
	var translations []models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND data->>'pageId' = ?", 
		projectID, "translations", pageID).Find(&translations).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find related translations"})
		return
	}
	
	// Delete all related translations
	if len(translations) > 0 {
		result := tx.Where("project_id = ? AND collection_name = ? AND data->>'pageId' = ?", 
			projectID, "translations", pageID).Delete(&models.Document{})
		
		if result.Error != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete related translations"})
			return
		}
		
		log.Printf("Deleted %d translations for page %s", result.RowsAffected, pageID)
	}
	
	// Delete the page itself
	result := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "pages", pageID).Delete(&models.Document{})
	
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete page"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit deletion"})
		return
	}
	
	log.Printf("Deleted page %s and %d related translations for project %d", pageID, len(translations), projectID)
	
	c.JSON(http.StatusOK, gin.H{
		"message": "Page and related translations deleted successfully",
		"id": pageID,
		"translationsDeleted": len(translations),
	})
}

// Pages endpoints
func (h *PortfolioHandler) GetPages(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get query parameters
	path := c.Query("path")
	language := c.DefaultQuery("language", "en")
	published := c.DefaultQuery("published", "true")

	// Query documents from pages collection
	var documents []models.Document
	query := h.db.Where("project_id = ? AND collection_name = ?", projectID, "pages")
	
	// Filter by language if specified
	if language != "" {
		query = query.Where("data->>'language' = ?", language)
	}
	
	// Filter by published status
	if published == "true" {
		query = query.Where("data->>'published' = 'true'")
	}

	// Filter by path/slug if specified
	if path != "" && path != "/" {
		// Remove leading slash for slug comparison
		slug := strings.TrimPrefix(path, "/")
		query = query.Where("data->>'slug' = ?", slug)
	}

	if err := query.Order("data->>'order' ASC, created_at ASC").Find(&documents).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch pages"})
		return
	}

	// Convert documents to pages format
	pages := make([]gin.H, 0, len(documents))
	for _, doc := range documents {
		page := gin.H{
			"_id": doc.ID,
			"id": doc.ID,
			"createdAt": doc.CreatedAt,
			"updatedAt": doc.UpdatedAt,
		}
		
		// Merge document data
		for key, value := range doc.Data {
			page[key] = value
		}
		
		pages = append(pages, page)
	}
	
	c.JSON(http.StatusOK, pages)
}

// Settings endpoints
func (h *PortfolioHandler) GetSettings(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get settings document
	var document models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "settings", "site").First(&document).Error; err != nil {
		// Return default settings if not found
		settings := gin.H{
			"siteName": "My Portfolio",
			"siteDescription": "Professional photography portfolio",
			"theme": "modern",
			"language": "en",
			"timezone": "UTC",
			"socialMedia": gin.H{
				"facebook": "",
				"instagram": "",
				"twitter": "",
				"linkedin": "",
			},
			"seo": gin.H{
				"metaTitle": "Photography Portfolio",
				"metaDescription": "Professional photography portfolio",
				"keywords": []string{"photography", "portfolio"},
			},
			"analytics": gin.H{
				"googleAnalytics": "",
				"enabled": false,
			},
			"contact": gin.H{
				"email": "contact@example.com",
				"phone": "",
				"address": "",
			},
		}
		c.JSON(http.StatusOK, settings)
		return
	}

	// Return document data with metadata
	settings := gin.H{
		"_id": document.ID,
		"id": document.ID,
		"createdAt": document.CreatedAt,
		"updatedAt": document.UpdatedAt,
	}
	
	// Merge document data
	for key, value := range document.Data {
		settings[key] = value
	}
	
	c.JSON(http.StatusOK, settings)
}

func (h *PortfolioHandler) UpdateSettings(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find the existing settings document
	var document models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "settings", "site").First(&document).Error; err != nil {
		// Create new settings document if not found
		document = models.Document{
			ID:             "site",
			CollectionName: "settings",
			ProjectID:      projectID,
			Data:           req,
			Version:        1,
			Author:         "portfolio_update",
		}
		
		if err := h.db.Create(&document).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create settings"})
			return
		}
	} else {
		// Update existing document
		for key, value := range req {
			document.Data[key] = value
		}
		document.Version++
		
		if err := h.db.Save(&document).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update settings"})
			return
		}
	}

	// Return updated settings
	response := gin.H{
		"_id": document.ID,
		"id": document.ID,
		"createdAt": document.CreatedAt,
		"updatedAt": document.UpdatedAt,
		"message": "Settings updated successfully",
	}
	
	// Merge document data
	for key, value := range document.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusOK, response)
}

// Branding endpoints
func (h *PortfolioHandler) GetBranding(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get branding document
	var document models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "branding", "visual").First(&document).Error; err != nil {
		// Return default branding if not found
		branding := gin.H{
			"logo": "/api/branding/logo.svg",
			"favicon": "/api/branding/favicon.ico",
			"primaryColor": "#1a1a1a",
			"secondaryColor": "#ffffff",
			"accentColor": "#007acc",
			"fonts": gin.H{
				"heading": "Inter",
				"body": "Inter",
			},
			"logoSettings": gin.H{
				"showText": true,
				"position": "left",
				"size": "medium",
			},
			"colorScheme": "light",
		}
		c.JSON(http.StatusOK, branding)
		return
	}

	// Return document data with metadata
	branding := gin.H{
		"_id": document.ID,
		"id": document.ID,
		"createdAt": document.CreatedAt,
		"updatedAt": document.UpdatedAt,
	}
	
	// Merge document data
	for key, value := range document.Data {
		branding[key] = value
	}
	
	c.JSON(http.StatusOK, branding)
}

// CreatePage creates a new page document
func (h *PortfolioHandler) CreatePage(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Generate unique page ID if not provided
	pageID, ok := req["id"].(string)
	if !ok || pageID == "" {
		pageID = uuid.New().String()
	}
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in CreatePage: %v", r)
		}
	}()
	
	// Check if page ID already exists
	var existingPage models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "pages", pageID).First(&existingPage).Error; err == nil {
		tx.Rollback()
		c.JSON(http.StatusConflict, gin.H{"error": "Page with this ID already exists"})
		return
	}
	
	// Set default values
	if _, exists := req["language"]; !exists {
		req["language"] = "en"
	}
	if _, exists := req["published"]; !exists {
		req["published"] = false
	}
	
	// Create page document
	page := models.Document{
		ID:             pageID,
		CollectionName: "pages",
		ProjectID:      projectID,
		Data:           req,
		Version:        1,
		Author:         "portfolio_admin",
	}
	
	if err := tx.Create(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create page"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit page creation"})
		return
	}
	
	log.Printf("Created page %s for project %d", pageID, projectID)
	
	// Return created page
	response := gin.H{
		"_id": page.ID,
		"id": page.ID,
		"createdAt": page.CreatedAt,
		"updatedAt": page.UpdatedAt,
		"message": "Page created successfully",
	}
	
	// Merge page data
	for key, value := range page.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusCreated, response)
}

// UpdatePage updates an existing page document
func (h *PortfolioHandler) UpdatePage(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	pageID := c.Param("pageId")
	if pageID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Page ID is required"})
		return
	}
	
	var req map[string]interface{}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in UpdatePage: %v", r)
		}
	}()
	
	// Find existing page
	var page models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, "pages", pageID).First(&page).Error; err != nil {
		tx.Rollback()
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find page"})
		}
		return
	}
	
	// Update page data
	for key, value := range req {
		page.Data[key] = value
	}
	page.Version++
	
	if err := tx.Save(&page).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update page"})
		return
	}
	
	if err := tx.Commit().Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to commit page update"})
		return
	}
	
	log.Printf("Updated page %s for project %d", pageID, projectID)
	
	// Return updated page
	response := gin.H{
		"_id": page.ID,
		"id": page.ID,
		"createdAt": page.CreatedAt,
		"updatedAt": page.UpdatedAt,
		"message": "Page updated successfully",
	}
	
	// Merge page data
	for key, value := range page.Data {
		response[key] = value
	}
	
	c.JSON(http.StatusOK, response)
}

// Users endpoints for portfolio
func (h *PortfolioHandler) GetPortfolioUsers(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Query AppUsers for this project
	var users []models.AppUser
	if err := h.db.Where("project_id = ?", projectID).Order("created_at DESC").Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	// Convert to response format
	response := make([]gin.H, 0, len(users))
	for _, user := range users {
		userResponse := gin.H{
			"_id": user.ID,
			"id": user.ID,
			"email": user.Email,
			"name": user.Name,
			"username": user.Username,
			"active": user.IsActive,
			"status": user.Status,
			"is_email_verified": user.IsEmailVerified,
			"last_login_at": user.LastLoginAt,
			"last_seen_at": user.LastSeenAt,
			"createdAt": user.CreatedAt,
			"updatedAt": user.UpdatedAt,
		}
		
		// Include profile data if available
		if user.ProfileData != nil {
			for key, value := range user.ProfileData {
				userResponse[key] = value
			}
		}
		
		response = append(response, userResponse)
	}
	
	c.JSON(http.StatusOK, response)
}

// Utility functions for portfolio management

// validateProjectAccess checks if the request has valid project access
func (h *PortfolioHandler) validateProjectAccess(c *gin.Context) (uint, bool) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return 0, false
	}
	return projectID, true
}

// createDocumentWithValidation creates a document with proper validation and transaction handling
func (h *PortfolioHandler) createDocumentWithValidation(projectID uint, collectionName, documentID string, data map[string]interface{}, author string) (*models.Document, error) {
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in createDocumentWithValidation: %v", r)
		}
	}()
	
	// Check if document already exists
	var existingDoc models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, collectionName, documentID).First(&existingDoc).Error; err == nil {
		tx.Rollback()
		return nil, fmt.Errorf("document with ID %s already exists", documentID)
	}
	
	// Create new document
	document := models.Document{
		ID:             documentID,
		CollectionName: collectionName,
		ProjectID:      projectID,
		Data:           data,
		Version:        1,
		Author:         author,
	}
	
	if err := tx.Create(&document).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	
	return &document, nil
}

// updateDocumentWithValidation updates a document with proper validation and transaction handling
func (h *PortfolioHandler) updateDocumentWithValidation(projectID uint, collectionName, documentID string, updates map[string]interface{}) (*models.Document, error) {
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in updateDocumentWithValidation: %v", r)
		}
	}()
	
	// Find existing document
	var document models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, collectionName, documentID).First(&document).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	// Update document data
	for key, value := range updates {
		document.Data[key] = value
	}
	document.Version++
	
	if err := tx.Save(&document).Error; err != nil {
		tx.Rollback()
		return nil, err
	}
	
	if err := tx.Commit().Error; err != nil {
		return nil, err
	}
	
	return &document, nil
}

// deleteDocumentWithCascade deletes a document and handles cascade deletion for related entities
func (h *PortfolioHandler) deleteDocumentWithCascade(projectID uint, collectionName, documentID string, cascadeFunc func(*gorm.DB, uint, string) error) error {
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("Panic in deleteDocumentWithCascade: %v", r)
		}
	}()
	
	// Verify document exists
	var document models.Document
	if err := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, collectionName, documentID).First(&document).Error; err != nil {
		tx.Rollback()
		return err
	}
	
	// Execute cascade deletion if function provided
	if cascadeFunc != nil {
		if err := cascadeFunc(tx, projectID, documentID); err != nil {
			tx.Rollback()
			return err
		}
	}
	
	// Delete the main document
	result := tx.Where("project_id = ? AND collection_name = ? AND id = ?", 
		projectID, collectionName, documentID).Delete(&models.Document{})
	
	if result.Error != nil {
		tx.Rollback()
		return result.Error
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		return gorm.ErrRecordNotFound
	}
	
	return tx.Commit().Error
}