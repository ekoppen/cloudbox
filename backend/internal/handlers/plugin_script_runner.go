package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type ScriptRunnerHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewScriptRunnerHandler(db *gorm.DB, cfg *config.Config) *ScriptRunnerHandler {
	return &ScriptRunnerHandler{
		db:  db,
		cfg: cfg,
	}
}

type Script struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	SQL         string    `json:"sql"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Template struct {
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Category    string            `json:"category"`
	Scripts     []TemplateScript  `json:"scripts"`
	Collections []string          `json:"collections"`
	Buckets     []string          `json:"buckets"`
	Settings    map[string]string `json:"settings"`
}

type TemplateScript struct {
	Name string `json:"name"`
	SQL  string `json:"sql"`
}

// GetProjectScripts returns all scripts for a project
func (h *ScriptRunnerHandler) GetProjectScripts(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	// In a real implementation, load scripts from database
	// For now, return mock data
	scripts := []Script{
		{
			ID:          "1",
			Name:        "Create Users Table",
			Description: "Creates a users table with authentication fields",
			Category:    "authentication",
			SQL: `CREATE TABLE IF NOT EXISTS users (
				id SERIAL PRIMARY KEY,
				email VARCHAR(255) UNIQUE NOT NULL,
				password_hash VARCHAR(255) NOT NULL,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          "2",
			Name:        "Create Posts Table",
			Description: "Creates a posts table for blog content",
			Category:    "content",
			SQL: `CREATE TABLE IF NOT EXISTS posts (
				id SERIAL PRIMARY KEY,
				user_id INTEGER REFERENCES users(id),
				title VARCHAR(255) NOT NULL,
				content TEXT,
				published BOOLEAN DEFAULT false,
				created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
				updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			);`,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"scripts": scripts,
	})
}

// ExecuteScript executes a SQL script for a project
func (h *ScriptRunnerHandler) ExecuteScript(c *gin.Context) {
	projectID := c.Param("projectId")
	scriptID := c.Param("scriptId")

	if projectID == "" || scriptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and Script ID are required",
		})
		return
	}

	// In a real implementation:
	// 1. Get project database connection
	// 2. Execute the script
	// 3. Return results

	// For now, return mock success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Script %s executed successfully for project %s", scriptID, projectID),
		"result": map[string]interface{}{
			"rows_affected": 0,
			"execution_time": "23ms",
			"status": "success",
		},
	})
}

// GetTemplates returns available setup templates
func (h *ScriptRunnerHandler) GetTemplates(c *gin.Context) {
	templates := []Template{
		{
			Name:        "blog-starter",
			Description: "Complete blog setup with users, posts, and comments",
			Category:    "starter",
			Scripts: []TemplateScript{
				{
					Name: "Create Users Table",
					SQL:  "CREATE TABLE users (...);",
				},
				{
					Name: "Create Posts Table",
					SQL:  "CREATE TABLE posts (...);",
				},
			},
			Collections: []string{"users", "posts", "comments"},
			Buckets:     []string{"avatars", "post-images"},
			Settings: map[string]string{
				"auth_enabled": "true",
				"storage_enabled": "true",
			},
		},
		{
			Name:        "ecommerce-starter",
			Description: "E-commerce setup with products, orders, and customers",
			Category:    "starter",
			Scripts: []TemplateScript{
				{
					Name: "Create Products Table",
					SQL:  "CREATE TABLE products (...);",
				},
				{
					Name: "Create Orders Table",
					SQL:  "CREATE TABLE orders (...);",
				},
			},
			Collections: []string{"products", "orders", "customers", "inventory"},
			Buckets:     []string{"product-images", "invoices"},
			Settings: map[string]string{
				"payment_enabled": "true",
				"inventory_tracking": "true",
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"templates": templates,
	})
}

// SetupProjectTemplate applies a template to a project
func (h *ScriptRunnerHandler) SetupProjectTemplate(c *gin.Context) {
	projectID := c.Param("projectId")
	templateName := c.Param("templateName")

	if projectID == "" || templateName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and Template Name are required",
		})
		return
	}

	// In a real implementation:
	// 1. Load template configuration
	// 2. Execute all template scripts
	// 3. Create collections
	// 4. Create storage buckets
	// 5. Apply settings

	// For now, return mock success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Template %s applied successfully to project %s", templateName, projectID),
		"result": map[string]interface{}{
			"scripts_executed": 5,
			"collections_created": 4,
			"buckets_created": 2,
			"total_time": "1.2s",
		},
	})
}

// CreateScript creates a new script for a project
func (h *ScriptRunnerHandler) CreateScript(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	var script Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid script data",
		})
		return
	}

	// Generate ID for new script
	script.ID = fmt.Sprintf("script_%d", time.Now().Unix())
	script.CreatedAt = time.Now()
	script.UpdatedAt = time.Now()

	// In a real implementation, save to database

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"script":  script,
	})
}

// UpdateScript updates an existing script
func (h *ScriptRunnerHandler) UpdateScript(c *gin.Context) {
	projectID := c.Param("projectId")
	scriptID := c.Param("scriptId")

	if projectID == "" || scriptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and Script ID are required",
		})
		return
	}

	var script Script
	if err := c.ShouldBindJSON(&script); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid script data",
		})
		return
	}

	script.ID = scriptID
	script.UpdatedAt = time.Now()

	// In a real implementation, update in database

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"script":  script,
	})
}

// DeleteScript deletes a script
func (h *ScriptRunnerHandler) DeleteScript(c *gin.Context) {
	projectID := c.Param("projectId")
	scriptID := c.Param("scriptId")

	if projectID == "" || scriptID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID and Script ID are required",
		})
		return
	}

	// In a real implementation, delete from database

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Script %s deleted successfully", scriptID),
	})
}

// ExecuteRawSQL executes raw SQL query (admin only)
func (h *ScriptRunnerHandler) ExecuteRawSQL(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Project ID is required",
		})
		return
	}

	var request struct {
		SQL string `json:"sql"`
	}

	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data",
		})
		return
	}

	// SECURITY: In production, this should have strict validation
	// and only allow certain SQL commands based on permissions

	// Mock execution result
	result := map[string]interface{}{
		"rows": []map[string]interface{}{
			{"id": 1, "name": "Test User", "email": "test@example.com"},
			{"id": 2, "name": "Admin User", "email": "admin@example.com"},
		},
		"rows_affected": 2,
		"execution_time": "15ms",
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"result":  result,
	})
}