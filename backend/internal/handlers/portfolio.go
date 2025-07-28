package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
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
	
	// Verify the page exists
	var page models.Document
	if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, "pages", pageID).First(&page).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}
	
	// Mock translation response (to be implemented with actual translation service)
	c.JSON(http.StatusOK, gin.H{
		"id": pageID + "_" + req.TargetLanguage,
		"pageId": pageID,
		"language": req.TargetLanguage,
		"status": "completed",
		"createdAt": time.Now(),
	})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "Page not found"})
		return
	}
	
	// Mock translations (to be implemented with actual translation tracking)
	translations := []gin.H{
		{
			"id": pageID + "_nl",
			"pageId": pageID,
			"language": "nl",
			"status": "completed",
			"createdAt": time.Now().Add(-24 * time.Hour),
		},
		{
			"id": pageID + "_de",
			"pageId": pageID,
			"language": "de", 
			"status": "completed",
			"createdAt": time.Now().Add(-48 * time.Hour),
		},
	}
	
	c.JSON(http.StatusOK, translations)
}

func (h *PortfolioHandler) DeleteTranslation(c *gin.Context) {
	translationID := c.Param("translationId")
	
	// Mock deletion
	c.JSON(http.StatusOK, gin.H{
		"message": "Translation deleted",
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