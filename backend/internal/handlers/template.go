package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type TemplateHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewTemplateHandler(db *gorm.DB, cfg *config.Config) *TemplateHandler {
	return &TemplateHandler{
		db:  db,
		cfg: cfg,
	}
}

// TemplateDefinition represents an app template for setting up collections
type TemplateDefinition struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Collections []CollectionTemplate   `json:"collections"`
	SeedData    map[string]interface{} `json:"seedData,omitempty"`
}

// CollectionTemplate defines a collection structure
type CollectionTemplate struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Schema      map[string]interface{} `json:"schema"`
	Indexes     []string               `json:"indexes,omitempty"`
	SeedData    []map[string]interface{} `json:"seedData,omitempty"`
}

// ListTemplates returns available templates
func (h *TemplateHandler) ListTemplates(c *gin.Context) {
	// For now, return available templates
	templates := []gin.H{
		{
			"name":        "photoportfolio",
			"version":     "1.0.0",
			"description": "Photography Portfolio Template",
			"collections": []string{"pages", "albums", "images", "settings", "branding"},
		},
		{
			"name":        "blog",
			"version":     "1.0.0", 
			"description": "Blog Template",
			"collections": []string{"posts", "categories", "tags", "comments"},
		},
		{
			"name":        "ecommerce",
			"version":     "1.0.0",
			"description": "E-commerce Template", 
			"collections": []string{"products", "categories", "orders", "customers"},
		},
	}
	
	c.JSON(http.StatusOK, templates)
}

// GetTemplate returns a specific template definition
func (h *TemplateHandler) GetTemplate(c *gin.Context) {
	templateName := c.Param("template")
	
	switch templateName {
	case "photoportfolio":
		template := h.getPhotoPortfolioTemplate()
		c.JSON(http.StatusOK, template)
	default:
		c.JSON(http.StatusNotFound, gin.H{"error": "Template not found"})
	}
}

// SetupPhotoPortfolio sets up the PhotoPortfolio template
func (h *TemplateHandler) SetupPhotoPortfolio(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Parse template from request body
	var templateDef TemplateDefinition
	if err := c.ShouldBindJSON(&templateDef); err != nil {
		// If no body provided, use default template
		templateDef = h.getPhotoPortfolioTemplate()
	}

	// Validate template name
	if templateDef.Name != "photoportfolio" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid template name"})
		return
	}

	// Setup collections
	results := make(map[string]interface{})
	
	for _, collectionTemplate := range templateDef.Collections {
		result, err := h.setupCollection(projectID, collectionTemplate)
		if err != nil {
			results[collectionTemplate.Name] = gin.H{
				"status": "error",
				"error":  err.Error(),
			}
			continue
		}
		results[collectionTemplate.Name] = gin.H{
			"status":     "success",
			"collection": result,
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":     "PhotoPortfolio template setup completed",
		"template":    templateDef.Name,
		"version":     templateDef.Version,
		"results":     results,
		"setupAt":     time.Now(),
	})
}

// setupCollection creates a collection and seeds it with data
func (h *TemplateHandler) setupCollection(projectID uint, template CollectionTemplate) (*models.Collection, error) {
	// Check if collection already exists
	var existingCollection models.Collection
	if err := h.db.Where("project_id = ? AND name = ?", projectID, template.Name).First(&existingCollection).Error; err == nil {
		// Collection exists, update it
		schemaJSON, _ := json.Marshal(template.Schema)
		indexesJSON, _ := json.Marshal(template.Indexes)
		
		existingCollection.Description = template.Description
		existingCollection.Schema = []string{string(schemaJSON)}
		existingCollection.Indexes = []string{string(indexesJSON)}
		existingCollection.LastModified = time.Now()
		
		if err := h.db.Save(&existingCollection).Error; err != nil {
			return nil, fmt.Errorf("failed to update collection: %v", err)
		}
		
		// Add seed data if provided
		if len(template.SeedData) > 0 {
			h.seedCollectionData(projectID, template.Name, template.SeedData)
		}
		
		return &existingCollection, nil
	}

	// Create new collection
	schemaJSON, _ := json.Marshal(template.Schema)
	indexesJSON, _ := json.Marshal(template.Indexes)
	
	collection := models.Collection{
		Name:          template.Name,
		Description:   template.Description,
		Schema:        []string{string(schemaJSON)},
		Indexes:       []string{string(indexesJSON)},
		ProjectID:     projectID,
		DocumentCount: 0,
		LastModified:  time.Now(),
	}

	if err := h.db.Create(&collection).Error; err != nil {
		return nil, fmt.Errorf("failed to create collection: %v", err)
	}

	// Add seed data if provided
	if len(template.SeedData) > 0 {
		h.seedCollectionData(projectID, template.Name, template.SeedData)
		
		// Update document count
		var count int64
		h.db.Model(&models.Document{}).Where("project_id = ? AND collection_name = ?", projectID, template.Name).Count(&count)
		collection.DocumentCount = count
		h.db.Save(&collection)
	}

	return &collection, nil
}

// seedCollectionData adds seed data to a collection
func (h *TemplateHandler) seedCollectionData(projectID uint, collectionName string, seedData []map[string]interface{}) {
	for i, data := range seedData {
		// Generate ID if not provided
		documentID, ok := data["id"].(string)
		if !ok {
			documentID = fmt.Sprintf("%s_%d", collectionName, i+1)
		}

		// Check if document already exists
		var existingDoc models.Document
		if err := h.db.Where("project_id = ? AND collection_name = ? AND id = ?", projectID, collectionName, documentID).First(&existingDoc).Error; err == nil {
			// Document exists, skip
			continue
		}

		// Create new document
		document := models.Document{
			ID:             documentID,
			CollectionName: collectionName,
			ProjectID:      projectID,
			Data:           data,
			Version:        1,
			Author:         "template_setup",
		}

		h.db.Create(&document)
	}
}

// getPhotoPortfolioTemplate returns the default PhotoPortfolio template
func (h *TemplateHandler) getPhotoPortfolioTemplate() TemplateDefinition {
	return TemplateDefinition{
		Name:        "photoportfolio",
		Version:     "1.0.0",
		Description: "Complete Photography Portfolio Template with pages, albums, images and settings",
		Collections: []CollectionTemplate{
			{
				Name:        "pages",
				Description: "Website pages and content",
				Schema: map[string]interface{}{
					"title":       "string",
					"slug":        "string",
					"content":     "text",
					"metaTitle":   "string",
					"metaDescription": "string",
					"published":   "boolean",
					"language":    "string",
					"template":    "string",
					"order":       "number",
				},
				Indexes: []string{"slug", "published", "language", "order"},
				SeedData: []map[string]interface{}{
					{
						"id":          "home",
						"title":       "Home",
						"slug":        "home",
						"content":     "Welcome to my photography portfolio. Discover my work and passion for capturing beautiful moments.",
						"metaTitle":   "Photography Portfolio - Home",
						"metaDescription": "Professional photography portfolio showcasing artistic vision and technical excellence.",
						"published":   true,
						"language":    "en",
						"template":    "home",
						"order":       1,
					},
					{
						"id":          "about",
						"title":       "About",
						"slug":        "about",
						"content":     "I'm a passionate photographer with over 10 years of experience capturing life's precious moments.",
						"metaTitle":   "About - Photography Portfolio",
						"metaDescription": "Learn about my photography journey and artistic approach.",
						"published":   true,
						"language":    "en",
						"template":    "page",
						"order":       2,
					},
					{
						"id":          "contact",
						"title":       "Contact",
						"slug":        "contact",
						"content":     "Get in touch to discuss your photography needs.",
						"metaTitle":   "Contact - Photography Portfolio",
						"metaDescription": "Contact information for photography services and inquiries.",
						"published":   true,
						"language":    "en",
						"template":    "contact",
						"order":       3,
					},
				},
			},
			{
				Name:        "albums",
				Description: "Photo albums and galleries",
				Schema: map[string]interface{}{
					"name":         "string",
					"description":  "text",
					"coverPhoto":   "string",
					"photos":       "array",
					"published":    "boolean",
					"order":        "number",
					"tags":         "array",
					"location":     "string",
					"date":         "date",
				},
				Indexes: []string{"published", "order", "tags"},
				SeedData: []map[string]interface{}{
					{
						"id":          "nature",
						"name":        "Nature Photography",
						"description": "Beautiful landscapes and nature scenes",
						"coverPhoto":  "nature_1",
						"photos":      []string{"nature_1", "nature_2", "nature_3"},
						"published":   true,
						"order":       1,
						"tags":        []string{"nature", "landscape", "outdoor"},
						"location":    "Various National Parks",
						"date":        "2024-01-01",
					},
					{
						"id":          "portraits",
						"name":        "Portrait Collection",
						"description": "Professional portrait photography",
						"coverPhoto":  "portrait_1",
						"photos":      []string{"portrait_1", "portrait_2", "portrait_3"},
						"published":   true,
						"order":       2,
						"tags":        []string{"portrait", "people", "studio"},
						"location":    "Studio",
						"date":        "2024-02-01",
					},
				},
			},
			{
				Name:        "images",
				Description: "Image metadata and information",
				Schema: map[string]interface{}{
					"title":        "string",
					"description":  "text",
					"filename":     "string",
					"url":          "string",
					"thumbnailUrl": "string",
					"altText":      "string",
					"tags":         "array",
					"camera":       "string",
					"lens":         "string",
					"settings":     "object",
					"location":     "string",
					"dateToken":    "date",
					"fileSize":     "number",
					"dimensions":   "object",
				},
				Indexes: []string{"tags", "dateToken", "camera"},
				SeedData: []map[string]interface{}{
					{
						"id":           "nature_1",
						"title":        "Mountain Sunrise",
						"description":  "Beautiful sunrise over mountain peaks",
						"filename":     "mountain_sunrise.jpg",
						"url":          "/api/images/mountain_sunrise.jpg",
						"thumbnailUrl": "/api/images/thumbs/mountain_sunrise.jpg",
						"altText":      "Sunrise over mountain peaks with golden light",
						"tags":         []string{"mountain", "sunrise", "landscape"},
						"camera":       "Canon EOS R5",
						"lens":         "Canon RF 24-70mm f/2.8L",
						"settings": map[string]interface{}{
							"aperture":     "f/8",
							"shutter":      "1/125",
							"iso":          200,
							"focalLength":  "35mm",
						},
						"location":    "Rocky Mountains",
						"dateToken":   "2024-01-15",
						"fileSize":    2485760,
						"dimensions": map[string]interface{}{
							"width":  3840,
							"height": 2160,
						},
					},
				},
			},
			{
				Name:        "settings",
				Description: "Site configuration and settings",
				Schema: map[string]interface{}{
					"siteName":        "string",
					"siteDescription": "text",
					"theme":           "string",
					"language":        "string",
					"timezone":        "string",
					"socialMedia":     "object",
					"seo":             "object",
					"analytics":       "object",
					"contact":         "object",
				},
				Indexes: []string{"language"},
				SeedData: []map[string]interface{}{
					{
						"id":              "site",
						"siteName":        "My Photography Portfolio",
						"siteDescription": "Professional photography portfolio showcasing artistic vision and technical excellence",
						"theme":           "modern",
						"language":        "en",
						"timezone":        "UTC",
						"socialMedia": map[string]interface{}{
							"instagram": "",
							"facebook":  "",
							"twitter":   "",
							"linkedin":  "",
						},
						"seo": map[string]interface{}{
							"metaTitle":       "Photography Portfolio - Professional Photographer",
							"metaDescription": "Discover stunning professional photography showcasing artistic vision and technical excellence",
							"keywords":        []string{"photography", "portfolio", "professional", "artistic"},
						},
						"analytics": map[string]interface{}{
							"googleAnalytics": "",
							"enabled":         false,
						},
						"contact": map[string]interface{}{
							"email":   "contact@example.com",
							"phone":   "",
							"address": "",
						},
					},
				},
			},
			{
				Name:        "branding",
				Description: "Visual branding and design settings",
				Schema: map[string]interface{}{
					"logo":            "string",
					"favicon":         "string",
					"primaryColor":    "string",
					"secondaryColor":  "string",
					"accentColor":     "string",
					"fonts":           "object",
					"logoSettings":    "object",
					"colorScheme":     "string",
				},
				Indexes: []string{"colorScheme"},
				SeedData: []map[string]interface{}{
					{
						"id":             "visual",
						"logo":           "/api/branding/logo.svg",
						"favicon":        "/api/branding/favicon.ico",
						"primaryColor":   "#1a1a1a",
						"secondaryColor": "#ffffff",
						"accentColor":    "#007acc",
						"fonts": map[string]interface{}{
							"heading": "Inter",
							"body":    "Inter",
						},
						"logoSettings": map[string]interface{}{
							"showText": true,
							"position": "left",
							"size":     "medium",
						},
						"colorScheme": "light",
					},
				},
			},
		},
	}
}