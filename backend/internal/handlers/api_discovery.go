package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"unicode"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type APIDiscoveryHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewAPIDiscoveryHandler(db *gorm.DB, cfg *config.Config) *APIDiscoveryHandler {
	return &APIDiscoveryHandler{db: db, cfg: cfg}
}

// capitalize capitalizes the first letter of a string
func capitalize(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

// APIRoute represents an API endpoint with its metadata
type APIRoute struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Description  string            `json:"description"`
	Category     string            `json:"category"`
	RequiresAuth bool              `json:"requiresAuth"`
	Parameters   []APIParameter    `json:"parameters,omitempty"`
	Example      *APIExample       `json:"example,omitempty"`
	Source       string            `json:"source"` // "database", "template", "core"
}

type APIParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Required    bool   `json:"required"`
	Description string `json:"description"`
}

type APIExample struct {
	Curl       string `json:"curl"`
	JavaScript string `json:"javascript"`
	Response   string `json:"response,omitempty"`
}

type APIDiscoveryResponse struct {
	BaseURL    string     `json:"baseURL"`
	Routes     []APIRoute `json:"routes"`
	Categories []string   `json:"categories"`
	Schema     *DBSchema  `json:"schema,omitempty"`
}

type DBSchema struct {
	Tables []TableSchema `json:"tables"`
}

type TableSchema struct {
	Name        string         `json:"name"`
	Columns     []ColumnSchema `json:"columns"`
	Description string         `json:"description,omitempty"`
}

type ColumnSchema struct {
	Name         string `json:"name"`
	Type         string `json:"type"`
	Nullable     bool   `json:"nullable"`
	PrimaryKey   bool   `json:"primaryKey"`
	ForeignKey   *ForeignKeyInfo `json:"foreignKey,omitempty"`
	Description  string `json:"description,omitempty"`
}

type ForeignKeyInfo struct {
	Table  string `json:"table"`
	Column string `json:"column"`
}

// GetAPIDiscovery returns all available API routes for a project
func (h *APIDiscoveryHandler) GetAPIDiscovery(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project info
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Generate base URL using config
	baseURL := fmt.Sprintf("%s/p/%d/api", h.cfg.BaseURL, project.ID)

	// Collect all routes
	var allRoutes []APIRoute
	var categories []string
	categorySet := make(map[string]bool)

	// Add core routes (always available)
	coreRoutes := h.getCoreRoutes()
	allRoutes = append(allRoutes, coreRoutes...)
	for _, route := range coreRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	// Add database-generated routes
	dbRoutes, dbSchema := h.getDatabaseRoutes(projectID)
	allRoutes = append(allRoutes, dbRoutes...)
	for _, route := range dbRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	// Add template-specific routes
	templateRoutes := h.getTemplateRoutes(projectID)
	allRoutes = append(allRoutes, templateRoutes...)
	for _, route := range templateRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	response := APIDiscoveryResponse{
		BaseURL:    baseURL,
		Routes:     allRoutes,
		Categories: categories,
		Schema:     dbSchema,
	}

	c.JSON(http.StatusOK, response)
}

// getCoreRoutes returns routes that are always available
func (h *APIDiscoveryHandler) getCoreRoutes() []APIRoute {
	return []APIRoute{
		{
			Method:      "GET",
			Path:        "/discovery/routes",
			Description: "API discovery - alle beschikbare routes ophalen",
			Category:    "System",
			RequiresAuth: false,
			Source:      "core",
			Example: &APIExample{
				Curl:       "curl \"{{baseURL}}/discovery/routes\"",
				JavaScript: "fetch('{{baseURL}}/discovery/routes')",
			},
		},
		{
			Method:      "GET",
			Path:        "/discovery/schema",
			Description: "Database schema informatie ophalen",
			Category:    "System",
			RequiresAuth: true,
			Source:      "core",
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/discovery/schema\"",
				JavaScript: "fetch('{{baseURL}}/discovery/schema', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "GET",
			Path:        "/health",
			Description: "API status controleren",
			Category:    "System",
			RequiresAuth: false,
			Source:      "core",
			Example: &APIExample{
				Curl:       "curl \"{{baseURL}}/health\"",
				JavaScript: "fetch('{{baseURL}}/health')",
			},
		},
		{
			Method:      "POST",
			Path:        "/users/register",
			Description: "Nieuwe gebruiker registreren",
			Category:    "Authentication",
			RequiresAuth: false,
			Source:      "core",
			Parameters: []APIParameter{
				{Name: "email", Type: "string", Required: true, Description: "Email address"},
				{Name: "password", Type: "string", Required: true, Description: "Password"},
				{Name: "name", Type: "string", Required: false, Description: "Display name"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"Content-Type: application/json\" -d '{\"email\":\"user@example.com\",\"password\":\"password\"}' \"{{baseURL}}/users/register\"",
				JavaScript: "fetch('{{baseURL}}/users/register', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: 'user@example.com', password: 'password' }) })",
			},
		},
		{
			Method:      "POST",
			Path:        "/users/login",
			Description: "Gebruiker inloggen",
			Category:    "Authentication",
			RequiresAuth: false,
			Source:      "core",
			Parameters: []APIParameter{
				{Name: "email", Type: "string", Required: true, Description: "Email address"},
				{Name: "password", Type: "string", Required: true, Description: "Password"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"Content-Type: application/json\" -d '{\"email\":\"user@example.com\",\"password\":\"password\"}' \"{{baseURL}}/users/login\"",
				JavaScript: "fetch('{{baseURL}}/users/login', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ email: 'user@example.com', password: 'password' }) })",
			},
		},
		{
			Method:      "GET",
			Path:        "/storage/buckets",
			Description: "Alle buckets ophalen",
			Category:    "Storage",
			RequiresAuth: true,
			Source:      "core",
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/storage/buckets\"",
				JavaScript: "fetch('{{baseURL}}/storage/buckets', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "POST",
			Path:        "/storage/buckets/{bucket}/files",
			Description: "Bestand uploaden naar bucket",
			Category:    "Storage",
			RequiresAuth: true,
			Source:      "core",
			Parameters: []APIParameter{
				{Name: "bucket", Type: "string", Required: true, Description: "Bucket name"},
				{Name: "file", Type: "file", Required: true, Description: "File to upload"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"X-API-Key: your-api-key\" -F \"file=@image.jpg\" \"{{baseURL}}/storage/buckets/{bucket}/files\"",
				JavaScript: "const formData = new FormData(); formData.append('file', file); fetch('{{baseURL}}/storage/buckets/{bucket}/files', { method: 'POST', headers: { 'X-API-Key': 'your-api-key' }, body: formData })",
			},
		},
	}
}

// getDatabaseRoutes scans the database and intelligently generates routes for tables
func (h *APIDiscoveryHandler) getDatabaseRoutes(projectID int) ([]APIRoute, *DBSchema) {
	var routes []APIRoute
	var schema DBSchema

	// Get all tables for this project
	tables := h.getProjectTables(projectID)
	
	for _, table := range tables {
		// Generate intelligent CRUD routes based on table structure and relationships
		routes = append(routes, h.generateIntelligentRoutes(table)...)
		
		// Add table to schema
		schema.Tables = append(schema.Tables, table)
	}

	return routes, &schema
}

// getProjectTables retrieves database tables that belong to this project
func (h *APIDiscoveryHandler) getProjectTables(projectID int) []TableSchema {
	var tables []TableSchema

	// Get project-specific tables - only show tables that are actually used by templates
	// or are part of the project's data model, not all CloudBox system tables
	
	// First, get template-specific tables for this project
	templateTables := h.getTemplateSpecificTables(projectID)
	
	// Query only template-relevant tables from information_schema
	if len(templateTables) == 0 {
		// No templates detected, return empty (only core routes will be shown)
		return tables
	}
	
	// Build placeholders for IN clause
	placeholders := make([]string, len(templateTables))
	args := make([]interface{}, len(templateTables))
	for i, table := range templateTables {
		placeholders[i] = "?"
		args[i] = table
	}
	
	query := fmt.Sprintf(`
		SELECT table_name, column_name, data_type, is_nullable, 
		       CASE WHEN column_name = 'id' THEN 'PRI' ELSE '' END as column_key
		FROM information_schema.columns 
		WHERE table_schema = 'public'
		AND table_name IN (%s)
		ORDER BY table_name, ordinal_position
	`, strings.Join(placeholders, ","))
	
	rows, err := h.db.Raw(query, args...).Rows()
	
	if err != nil {
		return tables
	}
	defer rows.Close()

	tableMap := make(map[string]*TableSchema)
	
	for rows.Next() {
		var tableName, columnName, dataType, isNullable, columnKey string
		if err := rows.Scan(&tableName, &columnName, &dataType, &isNullable, &columnKey); err != nil {
			continue
		}

		// Initialize table if not exists
		if _, exists := tableMap[tableName]; !exists {
			tableMap[tableName] = &TableSchema{
				Name:    tableName,
				Columns: []ColumnSchema{},
			}
		}

		// Add column to table
		column := ColumnSchema{
			Name:       columnName,
			Type:       dataType,
			Nullable:   isNullable == "YES",
			PrimaryKey: columnKey == "PRI",
		}

		tableMap[tableName].Columns = append(tableMap[tableName].Columns, column)
	}

	// Convert map to slice
	for _, table := range tableMap {
		tables = append(tables, *table)
	}

	return tables
}

// getTemplateSpecificTables returns only tables that are relevant for the project's templates
func (h *APIDiscoveryHandler) getTemplateSpecificTables(projectID int) []string {
	var templateTables []string
	
	// Get installed templates for this project
	installedTemplates := h.getInstalledTemplates(projectID)
	
	for _, templateName := range installedTemplates {
		switch templateName {
		case "photoportfolio":
			templateTables = append(templateTables, []string{
				"pages", "albums", "images", "portfolio_settings", "branding",
				"translations", "portfolio_users",
			}...)
		case "blog":
			templateTables = append(templateTables, []string{
				"posts", "categories", "tags", "comments", "blog_settings",
			}...)
		case "ecommerce":
			templateTables = append(templateTables, []string{
				"products", "categories", "orders", "payments", "customers",
			}...)
		// Add more template mappings as needed
		}
	}
	
	// Remove duplicates
	seen := make(map[string]bool)
	var uniqueTables []string
	for _, table := range templateTables {
		if !seen[table] {
			seen[table] = true
			uniqueTables = append(uniqueTables, table)
		}
	}
	
	return uniqueTables
}

// generateIntelligentRoutes creates intelligent CRUD routes based on table structure and content
func (h *APIDiscoveryHandler) generateIntelligentRoutes(table TableSchema) []APIRoute {
	routes := []APIRoute{}
	
	// Determine the table's purpose and generate appropriate routes
	tableType := h.analyzeTableType(table)
	category := h.getCategoryForTable(table)
	
	// Basic CRUD routes (always generated)
	routes = append(routes, h.generateBasicCRUD(table, tableType, category)...)
	
	// Add intelligent extra routes based on table analysis
	routes = append(routes, h.generateSmartRoutes(table, tableType, category)...)
	
	return routes
}

// analyzeTableType determines what kind of data this table holds
func (h *APIDiscoveryHandler) analyzeTableType(table TableSchema) string {
	tableName := strings.ToLower(table.Name)
	columns := make(map[string]bool)
	
	// Build column map for analysis
	for _, col := range table.Columns {
		columns[strings.ToLower(col.Name)] = true
	}
	
	// Analyze patterns to determine table type
	switch {
	case strings.Contains(tableName, "user") && (columns["email"] || columns["password"]):
		return "users"
	case strings.Contains(tableName, "page") && (columns["title"] || columns["content"]):
		return "content"
	case strings.Contains(tableName, "post") && (columns["title"] || columns["content"]):
		return "content"
	case strings.Contains(tableName, "article") && (columns["title"] || columns["content"]):
		return "content"
	case (strings.Contains(tableName, "image") || strings.Contains(tableName, "photo")) && (columns["url"] || columns["filename"] || columns["file_id"]):
		return "media"
	case strings.Contains(tableName, "album") && (columns["name"] || columns["title"]):
		return "collections"
	case strings.Contains(tableName, "category") && (columns["name"] || columns["title"]):
		return "taxonomy"
	case strings.Contains(tableName, "tag") && (columns["name"] || columns["title"]):
		return "taxonomy"
	case strings.Contains(tableName, "setting") && (columns["key"] || columns["name"]):
		return "settings"
	case strings.Contains(tableName, "config") && (columns["key"] || columns["name"]):
		return "settings"
	case strings.Contains(tableName, "translation") && (columns["language"] || columns["locale"]):
		return "i18n"
	case columns["published"] || columns["status"]:
		return "publishable"
	case columns["created_at"] && columns["updated_at"]:
		return "timestamped"
	default:
		return "generic"
	}
}

// getCategoryForTable determines the API category based on table analysis
func (h *APIDiscoveryHandler) getCategoryForTable(table TableSchema) string {
	tableType := h.analyzeTableType(table)
	tableName := strings.ToLower(table.Name)
	
	switch tableType {
	case "users":
		return "Users & Authentication"
	case "content":
		return "Content Management"
	case "media":
		return "Media & Files"
	case "collections":
		return "Collections & Albums"
	case "taxonomy":
		return "Categories & Tags"
	case "settings":
		return "Configuration"
	case "i18n":
		return "Internationalization"
	default:
		// Try to categorize by table name patterns
		switch {
		case strings.Contains(tableName, "order") || strings.Contains(tableName, "payment") || strings.Contains(tableName, "product"):
			return "E-commerce"
		case strings.Contains(tableName, "message") || strings.Contains(tableName, "chat") || strings.Contains(tableName, "comment"):
			return "Communication"
		case strings.Contains(tableName, "analytics") || strings.Contains(tableName, "stats") || strings.Contains(tableName, "metric"):
			return "Analytics & Reporting"
		default:
			return "Data Collections"
		}
	}
}

// generateBasicCRUD creates standard CRUD operations for any table
func (h *APIDiscoveryHandler) generateBasicCRUD(table TableSchema, tableType string, category string) []APIRoute {
	tableName := table.Name
	singularName := h.getSingularName(tableName)
	
	// Determine appropriate query parameters based on table structure
	queryParams := h.getIntelligentQueryParams(table)
	
	routes := []APIRoute{
		{
			Method:      "GET",
			Path:        "/" + tableName,
			Description: fmt.Sprintf("Alle %s ophalen", tableName),
			Category:    category,
			RequiresAuth: true,
			Source:      "database",
			Parameters:  queryParams,
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s', { headers: { 'X-API-Key': 'your-api-key' } })", tableName),
			},
		},
		{
			Method:      "POST",
			Path:        "/" + tableName,
			Description: fmt.Sprintf("Nieuwe %s aanmaken", singularName),
			Category:    category,
			RequiresAuth: true,
			Source:      "database",
			Parameters:  h.getCreateParameters(table),
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '%s' \"{{baseURL}}/%s\"", h.generateExampleData(table, "create"), tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify(%s) })", tableName, h.generateExampleData(table, "create")),
			},
		},
		{
			Method:      "GET",
			Path:        "/" + tableName + "/{id}",
			Description: fmt.Sprintf("Specifieke %s ophalen", singularName),
			Category:    category,
			RequiresAuth: true,
			Source:      "database",
			Parameters: []APIParameter{
				{Name: "id", Type: h.getPrimaryKeyType(table), Required: true, Description: fmt.Sprintf("%s ID", title(singularName))},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s/{id}\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}', { headers: { 'X-API-Key': 'your-api-key' } })", tableName),
			},
		},
		{
			Method:      "PUT",
			Path:        "/" + tableName + "/{id}",
			Description: fmt.Sprintf("%s bijwerken", title(singularName)),
			Category:    category,
			RequiresAuth: true,
			Source:      "database",
			Parameters:  append([]APIParameter{{Name: "id", Type: h.getPrimaryKeyType(table), Required: true, Description: fmt.Sprintf("%s ID", title(singularName))}}, h.getUpdateParameters(table)...),
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '%s' \"{{baseURL}}/%s/{id}\"", h.generateExampleData(table, "update"), tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify(%s) })", tableName, h.generateExampleData(table, "update")),
			},
		},
		{
			Method:      "DELETE",
			Path:        "/" + tableName + "/{id}",
			Description: fmt.Sprintf("%s verwijderen", title(singularName)),
			Category:    category,
			RequiresAuth: true,
			Source:      "database",
			Parameters: []APIParameter{
				{Name: "id", Type: h.getPrimaryKeyType(table), Required: true, Description: fmt.Sprintf("%s ID", title(singularName))},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -X DELETE -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s/{id}\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}', { method: 'DELETE', headers: { 'X-API-Key': 'your-api-key' } })", tableName),
			},
		},
	}
	
	return routes
}

// generateSmartRoutes creates additional intelligent routes based on table analysis
func (h *APIDiscoveryHandler) generateSmartRoutes(table TableSchema, tableType string, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Build column map for quick lookup
	columns := make(map[string]ColumnSchema)
	for _, col := range table.Columns {
		columns[strings.ToLower(col.Name)] = col
	}
	
	// Generate smart routes based on table type and structure
	switch tableType {
	case "content", "publishable":
		if _, hasPublished := columns["published"]; hasPublished {
			routes = append(routes, APIRoute{
				Method:      "PUT",
				Path:        "/" + tableName + "/{id}/publish",
				Description: fmt.Sprintf("%s publiceren/depubliceren", title(h.getSingularName(tableName))),
				Category:    category,
				RequiresAuth: true,
				Source:      "database",
				Parameters: []APIParameter{
					{Name: "id", Type: h.getPrimaryKeyType(table), Required: true, Description: "Item ID"},
					{Name: "published", Type: "boolean", Required: true, Description: "Published status"},
				},
				Example: &APIExample{
					Curl:       fmt.Sprintf("curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"published\":true}' \"{{baseURL}}/%s/{id}/publish\"", tableName),
					JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}/publish', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({published:true}) })", tableName),
				},
			})
		}
	
	case "media":
		if _, hasUrl := columns["url"]; hasUrl {
			routes = append(routes, APIRoute{
				Method:      "GET",
				Path:        "/" + tableName + "/{id}/download",
				Description: fmt.Sprintf("%s downloaden", title(h.getSingularName(tableName))),
				Category:    category,
				RequiresAuth: true,
				Source:      "database",
				Parameters: []APIParameter{
					{Name: "id", Type: h.getPrimaryKeyType(table), Required: true, Description: "Media ID"},
				},
				Example: &APIExample{
					Curl:       fmt.Sprintf("curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s/{id}/download\"", tableName),
					JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}/download', { headers: { 'X-API-Key': 'your-api-key' } })", tableName),
				},
			})
		}
		
	case "users":
		if _, hasEmail := columns["email"]; hasEmail {
			routes = append(routes, APIRoute{
				Method:      "POST",
				Path:        "/" + tableName + "/search",
				Description: fmt.Sprintf("%s zoeken op email, naam, etc.", title(tableName)),
				Category:    category,
				RequiresAuth: true,
				Source:      "database",
				Parameters: []APIParameter{
					{Name: "query", Type: "string", Required: true, Description: "Search query"},
					{Name: "fields", Type: "array", Required: false, Description: "Fields to search in"},
				},
				Example: &APIExample{
					Curl:       fmt.Sprintf("curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"query\":\"john\"}' \"{{baseURL}}/%s/search\"", tableName),
					JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/search', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({query:'john'}) })", tableName),
				},
			})
		}
		
	case "settings":
		if _, hasKey := columns["key"]; hasKey {
			routes = append(routes, APIRoute{
				Method:      "GET",
				Path:        "/" + tableName + "/key/{key}",
				Description: fmt.Sprintf("Specifieke instelling ophalen via key"),
				Category:    category,
				RequiresAuth: true,
				Source:      "database",
				Parameters: []APIParameter{
					{Name: "key", Type: "string", Required: true, Description: "Setting key"},
				},
				Example: &APIExample{
					Curl:       fmt.Sprintf("curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s/key/{key}\"", tableName),
					JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/key/{key}', { headers: { 'X-API-Key': 'your-api-key' } })", tableName),
				},
			})
		}
	}
	
	// Add bulk operations for tables with many records
	routes = append(routes, APIRoute{
		Method:      "POST",
		Path:        "/" + tableName + "/bulk",
		Description: fmt.Sprintf("Meerdere %s tegelijk aanmaken", tableName),
		Category:    category,
		RequiresAuth: true,
		Source:      "database",
		Parameters: []APIParameter{
			{Name: "items", Type: "array", Required: true, Description: fmt.Sprintf("Array of %s objects", h.getSingularName(tableName))},
		},
		Example: &APIExample{
			Curl:       fmt.Sprintf("curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"items\":[%s]}' \"{{baseURL}}/%s/bulk\"", h.generateExampleData(table, "create"), tableName),
			JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/bulk', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({items:[%s]}) })", tableName, h.generateExampleData(table, "create")),
		},
	})
	
	return routes
}

// Helper functions for intelligent route generation

func (h *APIDiscoveryHandler) getSingularName(tableName string) string {
	// Handle common pluralization patterns
	lower := strings.ToLower(tableName)
	switch {
	case strings.HasSuffix(lower, "ies"):
		return tableName[:len(tableName)-3] + "y"
	case strings.HasSuffix(lower, "ses"):
		return tableName[:len(tableName)-2]
	case strings.HasSuffix(lower, "s") && !strings.HasSuffix(lower, "ss"):
		return tableName[:len(tableName)-1]
	default:
		return tableName
	}
}

func (h *APIDiscoveryHandler) getPrimaryKeyType(table TableSchema) string {
	for _, col := range table.Columns {
		if col.PrimaryKey {
			// Map database types to API types
			switch strings.ToLower(col.Type) {
			case "int", "integer", "bigint", "smallint":
				return "integer"
			case "varchar", "text", "char", "uuid":
				return "string"
			default:
				return "string"
			}
		}
	}
	return "integer" // default assumption
}

func (h *APIDiscoveryHandler) getIntelligentQueryParams(table TableSchema) []APIParameter {
	params := []APIParameter{
		{Name: "limit", Type: "integer", Required: false, Description: "Maximum number of results"},
		{Name: "offset", Type: "integer", Required: false, Description: "Number of results to skip"},
		{Name: "sort", Type: "string", Required: false, Description: "Sort field (e.g. 'created_at', 'name')"},
		{Name: "order", Type: "string", Required: false, Description: "Sort direction ('asc' or 'desc')"},
	}
	
	// Add intelligent filters based on column analysis
	for _, col := range table.Columns {
		colName := strings.ToLower(col.Name)
		switch {
		case colName == "published" || colName == "active" || colName == "enabled":
			params = append(params, APIParameter{
				Name: colName, Type: "boolean", Required: false, 
				Description: fmt.Sprintf("Filter by %s status", colName),
			})
		case colName == "status":
			params = append(params, APIParameter{
				Name: "status", Type: "string", Required: false,
				Description: "Filter by status (e.g. 'active', 'inactive', 'pending')",
			})
		case colName == "category_id" || colName == "album_id" || strings.HasSuffix(colName, "_id"):
			params = append(params, APIParameter{
				Name: colName, Type: "integer", Required: false,
				Description: fmt.Sprintf("Filter by %s", strings.ReplaceAll(colName, "_", " ")),
			})
		case colName == "language" || colName == "locale":
			params = append(params, APIParameter{
				Name: colName, Type: "string", Required: false,
				Description: fmt.Sprintf("Filter by %s (e.g. 'en', 'nl')", colName),
			})
		case colName == "created_at" || colName == "updated_at":
			params = append(params, APIParameter{
				Name: colName + "_from", Type: "string", Required: false,
				Description: fmt.Sprintf("Filter %s from date (ISO format)", strings.ReplaceAll(colName, "_", " ")),
			})
			params = append(params, APIParameter{
				Name: colName + "_to", Type: "string", Required: false,
				Description: fmt.Sprintf("Filter %s to date (ISO format)", strings.ReplaceAll(colName, "_", " ")),
			})
		}
	}
	
	return params
}

func (h *APIDiscoveryHandler) getCreateParameters(table TableSchema) []APIParameter {
	var params []APIParameter
	
	for _, col := range table.Columns {
		// Skip auto-generated columns
		if col.PrimaryKey || strings.ToLower(col.Name) == "created_at" || strings.ToLower(col.Name) == "updated_at" {
			continue
		}
		
		param := APIParameter{
			Name:        col.Name,
			Type:        h.mapDBTypeToAPIType(col.Type),
			Required:    !col.Nullable,
			Description: h.generateColumnDescription(col),
		}
		params = append(params, param)
	}
	
	return params
}

func (h *APIDiscoveryHandler) getUpdateParameters(table TableSchema) []APIParameter {
	var params []APIParameter
	
	for _, col := range table.Columns {
		// Skip primary key and auto-generated columns for updates
		if col.PrimaryKey || strings.ToLower(col.Name) == "created_at" || strings.ToLower(col.Name) == "updated_at" {
			continue
		}
		
		param := APIParameter{
			Name:        col.Name,
			Type:        h.mapDBTypeToAPIType(col.Type),
			Required:    false, // Updates are usually partial
			Description: h.generateColumnDescription(col),
		}
		params = append(params, param)
	}
	
	return params
}

func (h *APIDiscoveryHandler) mapDBTypeToAPIType(dbType string) string {
	lower := strings.ToLower(dbType)
	switch {
	case strings.Contains(lower, "int"):
		return "integer"
	case strings.Contains(lower, "float") || strings.Contains(lower, "double") || strings.Contains(lower, "decimal"):
		return "number"
	case strings.Contains(lower, "bool"):
		return "boolean"
	case strings.Contains(lower, "json"):
		return "object"
	case strings.Contains(lower, "text") || strings.Contains(lower, "varchar") || strings.Contains(lower, "char"):
		return "string"
	case strings.Contains(lower, "date") || strings.Contains(lower, "time"):
		return "string"
	default:
		return "string"
	}
}

func (h *APIDiscoveryHandler) generateColumnDescription(col ColumnSchema) string {
	name := strings.ReplaceAll(col.Name, "_", " ")
	baseDesc := title(name)
	
	// Add type-specific information
	switch col.Type {
	case "text":
		baseDesc += " (long text)"
	case "json", "jsonb":
		baseDesc += " (JSON object)"
	}
	
	// Add constraints
	if !col.Nullable {
		baseDesc += " (required)"
	}
	
	return baseDesc
}

func (h *APIDiscoveryHandler) generateExampleData(table TableSchema, operation string) string {
	example := make(map[string]interface{})
	
	for _, col := range table.Columns {
		// Skip auto-generated columns
		if col.PrimaryKey || strings.ToLower(col.Name) == "created_at" || strings.ToLower(col.Name) == "updated_at" {
			continue
		}
		
		colName := strings.ToLower(col.Name)
		switch {
		case strings.Contains(colName, "name") || strings.Contains(colName, "title"):
			example[col.Name] = "Example Name"
		case strings.Contains(colName, "email"):
			example[col.Name] = "user@example.com"
		case strings.Contains(colName, "description") || strings.Contains(colName, "content"):
			example[col.Name] = "Description text here"
		case strings.Contains(colName, "url"):
			example[col.Name] = "https://example.com"
		case strings.Contains(colName, "published") || strings.Contains(colName, "active"):
			example[col.Name] = true
		case strings.Contains(colName, "status"):
			example[col.Name] = "active"
		case h.mapDBTypeToAPIType(col.Type) == "integer":
			example[col.Name] = 1
		case h.mapDBTypeToAPIType(col.Type) == "boolean":
			example[col.Name] = true
		default:
			example[col.Name] = "value"
		}
		
		// For update operations, only include 1-2 example fields
		if operation == "update" && len(example) >= 2 {
			break
		}
	}
	
	// Convert to JSON string
	if jsonBytes, err := json.Marshal(example); err == nil {
		return string(jsonBytes)
	}
	
	return "{}"
}

// getTemplateRoutes returns routes specific to installed templates
func (h *APIDiscoveryHandler) getTemplateRoutes(projectID int) []APIRoute {
	var routes []APIRoute

	// Query installed templates from template_deployments table
	installedTemplates := h.getInstalledTemplates(projectID)
	
	for _, template := range installedTemplates {
		templateRoutes := h.getTemplateRoutesByName(template)
		routes = append(routes, templateRoutes...)
	}

	return routes
}

// getInstalledTemplates queries which templates are deployed for this project
func (h *APIDiscoveryHandler) getInstalledTemplates(projectID int) []string {
	var templates []string
	
	// First, query the template_deployments table for officially deployed templates
	rows, err := h.db.Raw(`
		SELECT DISTINCT template_name 
		FROM template_deployments 
		WHERE project_id = ? AND status = 'deployed'
	`, projectID).Rows()
	
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var templateName string
			if err := rows.Scan(&templateName); err == nil {
				templates = append(templates, templateName)
			}
		}
	}

	// Note: Removed backwards compatibility check as it was causing new projects
	// to incorrectly detect PhotoPortfolio as installed. New projects should only
	// show templates that are explicitly deployed via template_deployments table.

	return templates
}

// containsTemplate checks if a template is already in the list
func (h *APIDiscoveryHandler) containsTemplate(templates []string, templateName string) bool {
	for _, t := range templates {
		if t == templateName {
			return true
		}
	}
	return false
}

// hasPhotoPortfolioData checks for existing PhotoPortfolio data (backwards compatibility)
func (h *APIDiscoveryHandler) hasPhotoPortfolioData(projectID int) bool {
	// Check if PhotoPortfolio-specific tables exist and have actual data
	var count int64
	err := h.db.Raw(`
		SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_name IN ('pages', 'albums', 'images', 'portfolio_settings', 'branding')
	`).Scan(&count).Error
	
	if err != nil || count == 0 {
		return false
	}

	// Additional check: see if there's actual PhotoPortfolio data (must have pages > 0)
	var dataCount int64
	h.db.Raw("SELECT COUNT(*) FROM pages").Scan(&dataCount)
	
	// Only consider PhotoPortfolio installed if we have the tables AND actual data
	return count >= 2 && dataCount > 0
}

// getTemplateRoutesByName returns routes for a specific template
func (h *APIDiscoveryHandler) getTemplateRoutesByName(templateName string) []APIRoute {
	switch templateName {
	case "photoportfolio":
		return h.getPhotoPortfolioRoutes()
	// Add more templates here as they're built
	// case "ecommerce":
	//     return h.getECommerceRoutes()
	default:
		return []APIRoute{}
	}
}

// getPhotoPortfolioRoutes returns PhotoPortfolio-specific routes
func (h *APIDiscoveryHandler) getPhotoPortfolioRoutes() []APIRoute {
	return []APIRoute{
		// Analytics API
		{
			Method:      "GET",
			Path:        "/analytics",
			Description: "PhotoPortfolio analytics en statistieken ophalen",
			Category:    "PhotoPortfolio Analytics",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "days", Type: "integer", Required: false, Description: "Number of days for analytics (default: 30)"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/analytics?days=14\"",
				JavaScript: "fetch('{{baseURL}}/analytics?days=14', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		// Branding API
		{
			Method:      "GET",
			Path:        "/branding",
			Description: "Brand instellingen, logo's en kleuren ophalen",
			Category:    "PhotoPortfolio Branding",
			RequiresAuth: true,
			Source:      "template",
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/branding\"",
				JavaScript: "fetch('{{baseURL}}/branding', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/branding",
			Description: "Brand instellingen, logo's en kleuren bijwerken",
			Category:    "PhotoPortfolio Branding",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "logo_url", Type: "string", Required: false, Description: "Logo URL"},
				{Name: "primary_color", Type: "string", Required: false, Description: "Primary brand color (hex)"},
				{Name: "font_family", Type: "string", Required: false, Description: "Primary font family"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"primary_color\":\"#3B82F6\",\"font_family\":\"Inter\"}' \"{{baseURL}}/branding\"",
				JavaScript: "fetch('{{baseURL}}/branding', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({primary_color:'#3B82F6', font_family:'Inter'}) })",
			},
		},
		// Translation/i18n API  
		{
			Method:      "GET",
			Path:        "/translations/languages",
			Description: "Beschikbare talen en locale instellingen ophalen",
			Category:    "PhotoPortfolio i18n",
			RequiresAuth: true,
			Source:      "template",
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/translations/languages\"",
				JavaScript: "fetch('{{baseURL}}/translations/languages', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/translations/languages",
			Description: "Ondersteunde talen en locale instellingen instellen",
			Category:    "PhotoPortfolio i18n",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "languages", Type: "array", Required: true, Description: "Array of supported language codes (e.g. ['en', 'nl', 'fr'])"},
				{Name: "default_language", Type: "string", Required: true, Description: "Default language code"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"languages\":[\"en\",\"nl\"],\"default_language\":\"en\"}' \"{{baseURL}}/translations/languages\"",
				JavaScript: "fetch('{{baseURL}}/translations/languages', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({languages:['en','nl'], default_language:'en'}) })",
			},
		},
		{
			Method:      "POST",
			Path:        "/translations/translate/{pageId}",
			Description: "Automatisch een pagina vertalen naar andere talen",
			Category:    "PhotoPortfolio i18n",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "pageId", Type: "integer", Required: true, Description: "ID van de pagina om te vertalen"},
				{Name: "target_language", Type: "string", Required: true, Description: "Doeltaal code (e.g. 'nl', 'fr')"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"target_language\":\"nl\"}' \"{{baseURL}}/translations/translate/{pageId}\"",
				JavaScript: "fetch('{{baseURL}}/translations/translate/{pageId}', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({target_language:'nl'}) })",
			},
		},
		{
			Method:      "GET",
			Path:        "/translations/page/{pageId}",
			Description: "Alle vertalingen van een pagina ophalen",
			Category:    "PhotoPortfolio i18n",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "pageId", Type: "integer", Required: true, Description: "Page ID"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/translations/page/{pageId}\"",
				JavaScript: "fetch('{{baseURL}}/translations/page/{pageId}', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		// Images Management API
		{
			Method:      "GET",
			Path:        "/images",
			Description: "Alle portfolio afbeeldingen en metadata ophalen",
			Category:    "PhotoPortfolio Images",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "album_id", Type: "integer", Required: false, Description: "Filter by album ID"},
				{Name: "tags", Type: "string", Required: false, Description: "Filter by tags (comma separated)"},
				{Name: "limit", Type: "integer", Required: false, Description: "Maximum number of results"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/images?album_id=1&limit=20\"",
				JavaScript: "fetch('{{baseURL}}/images?album_id=1&limit=20', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "POST",
			Path:        "/images",
			Description: "Nieuwe afbeelding metadata toevoegen (na upload naar storage)",
			Category:    "PhotoPortfolio Images",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "file_id", Type: "string", Required: true, Description: "CloudBox storage file ID"},
				{Name: "title", Type: "string", Required: false, Description: "Image title"},
				{Name: "description", Type: "string", Required: false, Description: "Image description"},
				{Name: "album_id", Type: "integer", Required: false, Description: "Album to add image to"},
				{Name: "tags", Type: "array", Required: false, Description: "Image tags"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"file_id\":\"uuid-123\",\"title\":\"Sunset\",\"album_id\":1}' \"{{baseURL}}/images\"",
				JavaScript: "fetch('{{baseURL}}/images', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({file_id:'uuid-123', title:'Sunset', album_id:1}) })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/images/{id}",
			Description: "Afbeelding metadata bijwerken",
			Category:    "PhotoPortfolio Images",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "id", Type: "integer", Required: true, Description: "Image ID"},
				{Name: "title", Type: "string", Required: false, Description: "Image title"},
				{Name: "description", Type: "string", Required: false, Description: "Image description"},
				{Name: "tags", Type: "array", Required: false, Description: "Image tags"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"title\":\"Beautiful Sunset\"}' \"{{baseURL}}/images/{id}\"",
				JavaScript: "fetch('{{baseURL}}/images/{id}', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({title:'Beautiful Sunset'}) })",
			},
		},
		{
			Method:      "DELETE",
			Path:        "/images/{id}",
			Description: "Afbeelding verwijderen (inclusief storage file)",
			Category:    "PhotoPortfolio Images",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "id", Type: "integer", Required: true, Description: "Image ID"},
			},
			Example: &APIExample{
				Curl:       "curl -X DELETE -H \"X-API-Key: your-api-key\" \"{{baseURL}}/images/{id}\"",
				JavaScript: "fetch('{{baseURL}}/images/{id}', { method: 'DELETE', headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		// Albums Management API
		{
			Method:      "GET",
			Path:        "/albums",
			Description: "Alle foto albums ophalen",
			Category:    "PhotoPortfolio Albums",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "published", Type: "boolean", Required: false, Description: "Filter by published status"},
				{Name: "featured", Type: "boolean", Required: false, Description: "Filter by featured albums"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/albums?published=true\"",
				JavaScript: "fetch('{{baseURL}}/albums?published=true', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "POST",
			Path:        "/albums",
			Description: "Nieuw foto album aanmaken",
			Category:    "PhotoPortfolio Albums",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "name", Type: "string", Required: true, Description: "Album name"},
				{Name: "description", Type: "string", Required: false, Description: "Album description"},
				{Name: "cover_image_id", Type: "integer", Required: false, Description: "Cover image ID"},
				{Name: "published", Type: "boolean", Required: false, Description: "Published status"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"name\":\"Wedding Photos\",\"published\":true}' \"{{baseURL}}/albums\"",
				JavaScript: "fetch('{{baseURL}}/albums', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({name:'Wedding Photos', published:true}) })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/albums/{id}",
			Description: "Album bijwerken",
			Category:    "PhotoPortfolio Albums",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "id", Type: "integer", Required: true, Description: "Album ID"},
				{Name: "name", Type: "string", Required: false, Description: "Album name"},
				{Name: "description", Type: "string", Required: false, Description: "Album description"},
				{Name: "published", Type: "boolean", Required: false, Description: "Published status"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"published\":false}' \"{{baseURL}}/albums/{id}\"",
				JavaScript: "fetch('{{baseURL}}/albums/{id}', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({published:false}) })",
			},
		},
		{
			Method:      "DELETE",
			Path:        "/albums/{id}",
			Description: "Album verwijderen (images blijven bestaan)",
			Category:    "PhotoPortfolio Albums",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "id", Type: "integer", Required: true, Description: "Album ID"},
			},
			Example: &APIExample{
				Curl:       "curl -X DELETE -H \"X-API-Key: your-api-key\" \"{{baseURL}}/albums/{id}\"",
				JavaScript: "fetch('{{baseURL}}/albums/{id}', { method: 'DELETE', headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		// Pages/Content Management API
		{
			Method:      "GET",
			Path:        "/pages",
			Description: "Website pagina's ophalen (home, about, contact, etc.)",
			Category:    "PhotoPortfolio Content",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "language", Type: "string", Required: false, Description: "Language code (e.g. 'en', 'nl')"},
				{Name: "published", Type: "boolean", Required: false, Description: "Filter by published status"},
				{Name: "path", Type: "string", Required: false, Description: "Filter by page path/slug"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/pages?language=en&published=true\"",
				JavaScript: "fetch('{{baseURL}}/pages?language=en&published=true', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "POST",
			Path:        "/pages",
			Description: "Nieuwe website pagina aanmaken",
			Category:    "PhotoPortfolio Content",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "title", Type: "string", Required: true, Description: "Page title"},
				{Name: "content", Type: "string", Required: true, Description: "Page content (HTML/Markdown)"},
				{Name: "path", Type: "string", Required: true, Description: "Page path/slug"},
				{Name: "language", Type: "string", Required: false, Description: "Language code"},
				{Name: "published", Type: "boolean", Required: false, Description: "Published status"},
			},
			Example: &APIExample{
				Curl:       "curl -X POST -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"title\":\"About Me\",\"content\":\"<p>Hello!</p>\",\"path\":\"/about\"}' \"{{baseURL}}/pages\"",
				JavaScript: "fetch('{{baseURL}}/pages', { method: 'POST', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({title:'About Me', content:'<p>Hello!</p>', path:'/about'}) })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/pages/{pageId}",
			Description: "Website pagina bijwerken",
			Category:    "PhotoPortfolio Content",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "pageId", Type: "integer", Required: true, Description: "Page ID"},
				{Name: "title", Type: "string", Required: false, Description: "Page title"},
				{Name: "content", Type: "string", Required: false, Description: "Page content"},
				{Name: "published", Type: "boolean", Required: false, Description: "Published status"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"published\":true}' \"{{baseURL}}/pages/{pageId}\"",
				JavaScript: "fetch('{{baseURL}}/pages/{pageId}', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({published:true}) })",
			},
		},
		{
			Method:      "DELETE",
			Path:        "/pages/{pageId}",
			Description: "Website pagina verwijderen (inclusief vertalingen)",
			Category:    "PhotoPortfolio Content",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "pageId", Type: "integer", Required: true, Description: "Page ID"},
			},
			Example: &APIExample{
				Curl:       "curl -X DELETE -H \"X-API-Key: your-api-key\" \"{{baseURL}}/pages/{pageId}\"",
				JavaScript: "fetch('{{baseURL}}/pages/{pageId}', { method: 'DELETE', headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		// Settings API
		{
			Method:      "GET",
			Path:        "/settings",
			Description: "Portfolio instellingen ophalen (SEO, sociale media, etc.)",
			Category:    "PhotoPortfolio Settings",
			RequiresAuth: true,
			Source:      "template",
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/settings\"",
				JavaScript: "fetch('{{baseURL}}/settings', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
		{
			Method:      "PUT",
			Path:        "/settings",
			Description: "Portfolio instellingen bijwerken",
			Category:    "PhotoPortfolio Settings",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "site_title", Type: "string", Required: false, Description: "Website title"},
				{Name: "site_description", Type: "string", Required: false, Description: "SEO description"},
				{Name: "contact_email", Type: "string", Required: false, Description: "Contact email"},
				{Name: "social_media", Type: "object", Required: false, Description: "Social media links"},
			},
			Example: &APIExample{
				Curl:       "curl -X PUT -H \"X-API-Key: your-api-key\" -H \"Content-Type: application/json\" -d '{\"site_title\":\"My Portfolio\",\"contact_email\":\"hello@example.com\"}' \"{{baseURL}}/settings\"",
				JavaScript: "fetch('{{baseURL}}/settings', { method: 'PUT', headers: { 'X-API-Key': 'your-api-key', 'Content-Type': 'application/json' }, body: JSON.stringify({site_title:'My Portfolio', contact_email:'hello@example.com'}) })",
			},
		},
		// Portfolio Users API
		{
			Method:      "GET",
			Path:        "/portfolio/users",
			Description: "Portfolio app gebruikers ophalen (website bezoekers, niet CloudBox admins)",
			Category:    "PhotoPortfolio Users",
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "active", Type: "boolean", Required: false, Description: "Filter by active status"},
				{Name: "limit", Type: "integer", Required: false, Description: "Maximum results"},
			},
			Example: &APIExample{
				Curl:       "curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/portfolio/users?active=true&limit=50\"",
				JavaScript: "fetch('{{baseURL}}/portfolio/users?active=true&limit=50', { headers: { 'X-API-Key': 'your-api-key' } })",
			},
		},
	}
}

// GetAPISchema returns the database schema for the project
func (h *APIDiscoveryHandler) GetAPISchema(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Get project tables and schema
	tables := h.getProjectTables(projectID)
	
	schema := DBSchema{
		Tables: tables,
	}

	c.JSON(http.StatusOK, schema)
}

// RefreshAPIDiscovery triggers a refresh of API discovery for external apps
func (h *APIDiscoveryHandler) RefreshAPIDiscovery(c *gin.Context) {
	projectIDStr := c.Param("project_id")
	projectID, err := strconv.Atoi(projectIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	// Verify project exists
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Parse request body for refresh options
	var refreshRequest struct {
		Reason      string   `json:"reason"`        // Why the refresh was triggered
		Source      string   `json:"source"`       // What app/service triggered this
		Templates   []string `json:"templates"`    // Specific templates to refresh (optional)
		ForceRescan bool     `json:"forceRescan"`  // Force full database rescan
		Webhook     string   `json:"webhook"`      // Optional webhook to call when done
	}

	if err := c.ShouldBindJSON(&refreshRequest); err != nil {
		// Allow empty body - set defaults
		refreshRequest.Reason = "External trigger"
		refreshRequest.Source = "Unknown"
		refreshRequest.ForceRescan = true
	}

	// Perform discovery refresh
	baseURL := fmt.Sprintf("%s/p/%d/api", h.cfg.BaseURL, project.ID)
	
	// Get all routes (this will re-scan everything)
	var allRoutes []APIRoute
	var categories []string
	categorySet := make(map[string]bool)

	// Add core routes
	coreRoutes := h.getCoreRoutes()
	allRoutes = append(allRoutes, coreRoutes...)
	for _, route := range coreRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	// Add database-generated routes (force rescan)
	dbRoutes, dbSchema := h.getDatabaseRoutes(projectID)
	allRoutes = append(allRoutes, dbRoutes...)
	for _, route := range dbRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	// Add template-specific routes (force rescan)
	templateRoutes := h.getTemplateRoutes(projectID)
	allRoutes = append(allRoutes, templateRoutes...)
	for _, route := range templateRoutes {
		if !categorySet[route.Category] {
			categories = append(categories, route.Category)
			categorySet[route.Category] = true
		}
	}

	response := APIDiscoveryResponse{
		BaseURL:    baseURL,
		Routes:     allRoutes,
		Categories: categories,
		Schema:     dbSchema,
	}

	// Create scan report message in project inbox
	go func() {
		err := h.createScanReportMessage(projectID, response, refreshRequest)
		if err != nil {
			fmt.Printf("Failed to create scan report message: %v\n", err)
		}
	}()

	// Optional: Call webhook if provided
	if refreshRequest.Webhook != "" {
		go h.callWebhook(refreshRequest.Webhook, map[string]interface{}{
			"projectId":    projectID,
			"routeCount":   len(allRoutes),
			"categories":   categories,
			"refreshedAt":  time.Now().Unix(),
			"reason":       refreshRequest.Reason,
			"source":       refreshRequest.Source,
		})
	}

	// Return success with discovery data
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"message":     "API discovery refreshed successfully",
		"projectId":   projectID,
		"routeCount":  len(allRoutes),
		"categories":  categories,
		"triggeredBy": refreshRequest.Source,
		"reason":      refreshRequest.Reason,
		"discovery":   response,
	})
}

// callWebhook sends a POST request to the webhook URL (async)
func (h *APIDiscoveryHandler) callWebhook(webhookURL string, payload map[string]interface{}) {
	// This runs in a goroutine, so errors are just logged
	// In production, you might want more robust webhook handling
	// (retries, dead letter queues, etc.)
	
	// TODO: Implement webhook calling logic
	// For now, just log that we would call it
	fmt.Printf("Would call webhook %s with payload: %+v\n", webhookURL, payload)
}

// createScanReportMessage creates an inbox message with the API discovery scan report
func (h *APIDiscoveryHandler) createScanReportMessage(projectID int, discovery APIDiscoveryResponse, refreshRequest struct {
	Reason      string   `json:"reason"`
	Source      string   `json:"source"`
	Templates   []string `json:"templates"`
	ForceRescan bool     `json:"forceRescan"`
	Webhook     string   `json:"webhook"`
}) error {
	// First, ensure we have a system channel for this project
	channelID, err := h.ensureSystemChannel(projectID)
	if err != nil {
		return fmt.Errorf("failed to ensure system channel: %v", err)
	}

	// Create scan report content
	scanReport := h.generateScanReport(discovery, refreshRequest)
	
	// Create the message
	message := models.Message{
		ID:        uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Content:   scanReport,
		Type:      "system",
		ChannelID: channelID,
		UserID:    "system", // System-generated message
		ProjectID: uint(projectID),
		Metadata: map[string]interface{}{
			"message_type":   "api_discovery_scan",
			"route_count":    len(discovery.Routes),
			"category_count": len(discovery.Categories),
			"source":         refreshRequest.Source,
			"reason":         refreshRequest.Reason,
			"timestamp":      time.Now().Unix(),
		},
	}

	// Save to database
	if err := h.db.Create(&message).Error; err != nil {
		return fmt.Errorf("failed to create scan report message: %v", err)
	}

	return nil
}

// ensureSystemChannel ensures a system channel exists for project messages
func (h *APIDiscoveryHandler) ensureSystemChannel(projectID int) (string, error) {
	var channel models.Channel
	
	// Try to find existing system channel
	err := h.db.Where("project_id = ? AND type = ? AND name = ?", projectID, "system", "inbox").First(&channel).Error
	
	if err == gorm.ErrRecordNotFound {
		// Create system channel
		channel = models.Channel{
			ID:          uuid.New().String(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Name:        "inbox",
			Description: "System messages and notifications",
			Type:        "system",
			Topic:       "Project notifications and system messages",
			IsActive:    true,
			MaxMembers:  0, // Unlimited
			Settings: map[string]interface{}{
				"auto_created": true,
				"system_only":  true,
			},
			ProjectID:     uint(projectID),
			CreatedBy:     "system",
			MemberCount:   0,
			MessageCount:  0,
			LastActivity:  time.Now(),
		}
		
		if err := h.db.Create(&channel).Error; err != nil {
			return "", fmt.Errorf("failed to create system channel: %v", err)
		}
	} else if err != nil {
		return "", fmt.Errorf("failed to query system channel: %v", err)
	}
	
	return channel.ID, nil
}

// generateScanReport creates a formatted scan report message
func (h *APIDiscoveryHandler) generateScanReport(discovery APIDiscoveryResponse, refreshRequest struct {
	Reason      string   `json:"reason"`
	Source      string   `json:"source"`
	Templates   []string `json:"templates"`
	ForceRescan bool     `json:"forceRescan"`
	Webhook     string   `json:"webhook"`
}) string {
	// Count routes by source
	sourceCounts := make(map[string]int)
	for _, route := range discovery.Routes {
		sourceCounts[route.Source]++
	}

	// Generate report
	report := "🔍 **API Discovery Scan Report**\n\n"
	
	// Trigger information
	if refreshRequest.Source != "" {
		report += fmt.Sprintf("**Triggered by:** %s\n", refreshRequest.Source)
	}
	if refreshRequest.Reason != "" {
		report += fmt.Sprintf("**Reason:** %s\n", refreshRequest.Reason)
	}
	report += fmt.Sprintf("**Scan time:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))

	// Summary statistics
	report += "## 📊 Summary\n"
	report += fmt.Sprintf("- **Total API routes discovered:** %d\n", len(discovery.Routes))
	report += fmt.Sprintf("- **Categories:** %d (%s)\n", len(discovery.Categories), strings.Join(discovery.Categories, ", "))
	
	if len(sourceCounts) > 0 {
		report += "\n### Routes by Source:\n"
		for source, count := range sourceCounts {
			var icon string
			switch source {
			case "core":
				icon = "⚙️"
			case "database":
				icon = "🗄️"
			case "template":
				icon = "📦"
			default:
				icon = "❓"
			}
			report += fmt.Sprintf("- %s **%s:** %d routes\n", icon, title(source), count)
		}
	}

	// Template information
	if len(refreshRequest.Templates) > 0 {
		report += fmt.Sprintf("\n### Templates Processed:\n")
		for _, template := range refreshRequest.Templates {
			report += fmt.Sprintf("- 📦 %s\n", template)
		}
	}

	// Database schema info
	if discovery.Schema != nil && len(discovery.Schema.Tables) > 0 {
		report += fmt.Sprintf("\n### Database Schema:\n")
		report += fmt.Sprintf("- **Tables scanned:** %d\n", len(discovery.Schema.Tables))
		
		// Show first few tables as examples
		maxShow := 5
		if len(discovery.Schema.Tables) > 0 {
			report += "- **Example tables:** "
			examples := []string{}
			for i, table := range discovery.Schema.Tables {
				if i >= maxShow {
					examples = append(examples, "...")
					break
				}
				examples = append(examples, table.Name)
			}
			report += strings.Join(examples, ", ") + "\n"
		}
	}

	// Footer
	report += "\n---\n"
	report += "💡 **Tip:** Routes are automatically updated when your database schema or templates change. "
	report += "You can manually refresh using the \"API Routes Vernieuwen\" button in the API section.\n\n"
	report += fmt.Sprintf("🔗 **Base URL:** `%s`", discovery.BaseURL)

	return report
}

// Helper function to capitalize strings (replacement for deprecated strings.Title)
func title(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// ========== SUPABASE-INSPIRED INTELLIGENT API DISCOVERY ==========

// getTemplateDeployments retrieves all templates installed for a project (like Supabase's real-time detection)
func (h *APIDiscoveryHandler) getTemplateDeployments(projectID int) []struct {
	TemplateName string
	Variables    map[string]interface{}
} {
	var deployments []struct {
		TemplateName string
		Variables    map[string]interface{}
	}
	
	// Query template_deployments table to see what's actually installed
	query := `
		SELECT template_name, variables 
		FROM template_deployments 
		WHERE project_id = ? AND status = 'deployed'
	`
	
	rows, err := h.db.Raw(query, projectID).Rows()
	if err != nil {
		return deployments
	}
	defer rows.Close()
	
	for rows.Next() {
		var templateName string
		var variablesJSON string
		
		if err := rows.Scan(&templateName, &variablesJSON); err == nil {
			deployment := struct {
				TemplateName string
				Variables    map[string]interface{}
			}{
				TemplateName: templateName,
				Variables:    make(map[string]interface{}),
			}
			
			// Parse variables if present
			if variablesJSON != "" {
				json.Unmarshal([]byte(variablesJSON), &deployment.Variables)
			}
			
			deployments = append(deployments, deployment)
		}
	}
	
	return deployments
}

// generateSmartTemplateRoutes creates routes based on actual template deployment + database schema
func (h *APIDiscoveryHandler) generateSmartTemplateRoutes(projectID int, templateName string) []APIRoute {
	routes := []APIRoute{}
	
	// Get actual database tables for this project
	tables := h.getProjectTables(projectID)
	
	switch templateName {
	case "photoportfolio":
		routes = append(routes, h.generatePhotoPortfolioIntelligentRoutes(projectID, tables)...)
	case "ecommerce":
		routes = append(routes, h.generateEcommerceIntelligentRoutes(projectID, tables)...)
	case "blog":
		routes = append(routes, h.generateBlogIntelligentRoutes(projectID, tables)...)
	default:
		// Generic template - generate basic routes
		routes = append(routes, h.generateGenericTemplateRoutes(projectID, tables)...)
	}
	
	// Add custom functions/RPC endpoints (Supabase style)
	routes = append(routes, h.getTemplateCustomFunctions(projectID, templateName)...)
	
	return routes
}

// generateRoutesFromTablePatterns intelligently detects templates from table patterns (fallback method)
func (h *APIDiscoveryHandler) generateRoutesFromTablePatterns(projectID int) []APIRoute {
	routes := []APIRoute{}
	tables := h.getProjectTables(projectID)
	
	// Analyze table patterns to detect likely templates
	templateSignatures := h.detectTemplateSignatures(tables)
	
	for template, confidence := range templateSignatures {
		if confidence > 0.7 { // Only generate routes for high-confidence detections
			routes = append(routes, h.generateSmartTemplateRoutes(projectID, template)...)
		}
	}
	
	return routes
}

// detectTemplateSignatures analyzes table patterns to identify templates
func (h *APIDiscoveryHandler) detectTemplateSignatures(tables []TableSchema) map[string]float64 {
	signatures := make(map[string]float64)
	
	// PhotoPortfolio signature detection
	photoPortfolioScore := 0.0
	for _, table := range tables {
		tableName := strings.ToLower(table.Name)
		switch {
		case tableName == "albums" || strings.Contains(tableName, "album"):
			photoPortfolioScore += 0.3
		case tableName == "images" || strings.Contains(tableName, "photo"):
			photoPortfolioScore += 0.3
		case tableName == "pages" && h.hasColumn(table, "content"):
			photoPortfolioScore += 0.2
		case strings.Contains(tableName, "translation") && h.hasColumn(table, "language"):
			photoPortfolioScore += 0.2
		}
	}
	if photoPortfolioScore > 0 {
		signatures["photoportfolio"] = photoPortfolioScore
	}
	
	// Blog signature detection
	blogScore := 0.0
	for _, table := range tables {
		tableName := strings.ToLower(table.Name)
		switch {
		case tableName == "posts" && h.hasColumn(table, "content"):
			blogScore += 0.4
		case tableName == "categories" && h.hasColumn(table, "name"):
			blogScore += 0.3
		case tableName == "tags" && h.hasColumn(table, "name"):
			blogScore += 0.3
		}
	}
	if blogScore > 0 {
		signatures["blog"] = blogScore
	}
	
	// E-commerce signature detection
	ecommerceScore := 0.0
	for _, table := range tables {
		tableName := strings.ToLower(table.Name)
		switch {
		case tableName == "products" && h.hasColumn(table, "price"):
			ecommerceScore += 0.4
		case tableName == "orders" && h.hasColumn(table, "total"):
			ecommerceScore += 0.3
		case strings.Contains(tableName, "payment"):
			ecommerceScore += 0.3
		}
	}
	if ecommerceScore > 0 {
		signatures["ecommerce"] = ecommerceScore
	}
	
	return signatures
}

// generatePhotoPortfolioIntelligentRoutes creates PhotoPortfolio routes based on actual schema
func (h *APIDiscoveryHandler) generatePhotoPortfolioIntelligentRoutes(projectID int, tables []TableSchema) []APIRoute {
	routes := []APIRoute{}
	
	// Scan tables and generate appropriate routes
	for _, table := range tables {
		tableName := strings.ToLower(table.Name)
		
		switch {
		case tableName == "pages" || tableName == "portfolio_pages":
			routes = append(routes, h.generateContentRoutes(table, "PhotoPortfolio Content")...)
		case tableName == "albums" || tableName == "photo_albums":
			routes = append(routes, h.generateAlbumRoutes(table, "PhotoPortfolio Albums")...)
		case tableName == "images" || tableName == "portfolio_images":
			routes = append(routes, h.generateImageRoutes(table, "PhotoPortfolio Images")...)
		case tableName == "settings" || tableName == "portfolio_settings":
			routes = append(routes, h.generateSettingsRoutes(table, "PhotoPortfolio Settings")...)
		case strings.Contains(tableName, "translation") && h.hasColumn(table, "language"):
			routes = append(routes, h.generateTranslationRoutes(table, "PhotoPortfolio i18n")...)
		}
	}
	
	return routes
}

// generateContentRoutes creates intelligent routes for content tables (pages, posts, articles)
func (h *APIDiscoveryHandler) generateContentRoutes(table TableSchema, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Basic CRUD (always included)
	routes = append(routes, h.generateBasicCRUD(table, "content", category)...)
	
	// Content-specific intelligent routes
	if h.hasColumn(table, "published") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/published",
			Description: fmt.Sprintf("Alleen gepubliceerde %s ophalen", tableName),
			Category:    category,
			RequiresAuth: false, // Public content
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "language", Type: "string", Required: false, Description: "Language filter"},
				{Name: "limit", Type: "integer", Required: false, Description: "Maximum results"},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/published?language=en&limit=10\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/published?language=en&limit=10')", tableName),
			},
		})
	}
	
	if h.hasColumn(table, "path") || h.hasColumn(table, "slug") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/by-path/{path}",
			Description: fmt.Sprintf("%s ophalen via path/slug", title(h.getSingularName(tableName))),
			Category:    category,
			RequiresAuth: false,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "path", Type: "string", Required: true, Description: "Page path or slug"},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/by-path/about\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/by-path/about')", tableName),
			},
		})
	}
	
	return routes
}

// generateAlbumRoutes creates intelligent routes for album/collection tables
func (h *APIDiscoveryHandler) generateAlbumRoutes(table TableSchema, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Basic CRUD
	routes = append(routes, h.generateBasicCRUD(table, "collections", category)...)
	
	// Album-specific routes
	routes = append(routes, APIRoute{
		Method:      "GET",
		Path:        "/" + tableName + "/{id}/images",
		Description: fmt.Sprintf("Alle afbeeldingen in album ophalen"),
		Category:    category,
		RequiresAuth: false,
		Source:      "template",
		Parameters: []APIParameter{
			{Name: "id", Type: "integer", Required: true, Description: "Album ID"},
			{Name: "limit", Type: "integer", Required: false, Description: "Maximum results"},
		},
		Example: &APIExample{
			Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/{id}/images?limit=20\"", tableName),
			JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}/images?limit=20')", tableName),
		},
	})
	
	if h.hasColumn(table, "featured") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/featured",
			Description: fmt.Sprintf("Uitgelichte %s ophalen", tableName),
			Category:    category,
			RequiresAuth: false,
			Source:      "template",
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/featured\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/featured')", tableName),
			},
		})
	}
	
	return routes
}

// generateImageRoutes creates intelligent routes for image/media tables  
func (h *APIDiscoveryHandler) generateImageRoutes(table TableSchema, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Basic CRUD
	routes = append(routes, h.generateBasicCRUD(table, "media", category)...)
	
	// Image-specific routes
	if h.hasColumn(table, "file_id") || h.hasColumn(table, "url") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/{id}/url",
			Description: fmt.Sprintf("Publieke URL van afbeelding ophalen"),
			Category:    category,
			RequiresAuth: false,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "id", Type: "integer", Required: true, Description: "Image ID"},
				{Name: "size", Type: "string", Required: false, Description: "Image size (thumb, medium, large)"},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/{id}/url?size=medium\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/{id}/url?size=medium')", tableName),
			},
		})
	}
	
	return routes
}

// generateSettingsRoutes creates intelligent routes for settings tables
func (h *APIDiscoveryHandler) generateSettingsRoutes(table TableSchema, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Settings typically use key-value access patterns (like Supabase)
	if h.hasColumn(table, "key") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/{key}",
			Description: fmt.Sprintf("Instelling ophalen via key"),
			Category:    category,
			RequiresAuth: true,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "key", Type: "string", Required: true, Description: "Setting key"},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl -H \"X-API-Key: your-api-key\" \"{{baseURL}}/%s/site_title\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/site_title', { headers: { 'X-API-Key': 'your-api-key' } })", tableName),
			},
		})
	}
	
	// Basic CRUD for settings
	routes = append(routes, h.generateBasicCRUD(table, "settings", category)...)
	
	return routes
}

// generateTranslationRoutes creates intelligent routes for translation tables
func (h *APIDiscoveryHandler) generateTranslationRoutes(table TableSchema, category string) []APIRoute {
	routes := []APIRoute{}
	tableName := table.Name
	
	// Basic CRUD
	routes = append(routes, h.generateBasicCRUD(table, "i18n", category)...)
	
	// Translation-specific routes  
	if h.hasColumn(table, "language") && h.hasColumn(table, "page_id") {
		routes = append(routes, APIRoute{
			Method:      "GET",
			Path:        "/" + tableName + "/page/{pageId}/{language}",
			Description: fmt.Sprintf("Specifieke vertaling ophalen"),
			Category:    category,
			RequiresAuth: false,
			Source:      "template",
			Parameters: []APIParameter{
				{Name: "pageId", Type: "integer", Required: true, Description: "Page ID"},
				{Name: "language", Type: "string", Required: true, Description: "Language code"},
			},
			Example: &APIExample{
				Curl:       fmt.Sprintf("curl \"{{baseURL}}/%s/page/{pageId}/nl\"", tableName),
				JavaScript: fmt.Sprintf("fetch('{{baseURL}}/%s/page/{pageId}/nl')", tableName),
			},
		})
	}
	
	return routes
}

// Placeholder functions for other templates
func (h *APIDiscoveryHandler) generateEcommerceIntelligentRoutes(projectID int, tables []TableSchema) []APIRoute {
	// Would implement e-commerce specific route generation
	return []APIRoute{}
}

func (h *APIDiscoveryHandler) generateBlogIntelligentRoutes(projectID int, tables []TableSchema) []APIRoute {
	// Would implement blog specific route generation  
	return []APIRoute{}
}

func (h *APIDiscoveryHandler) generateGenericTemplateRoutes(projectID int, tables []TableSchema) []APIRoute {
	routes := []APIRoute{}
	// Generate basic CRUD for all tables
	for _, table := range tables {
		routes = append(routes, h.generateIntelligentRoutes(table)...)
	}
	return routes
}

// getTemplateCustomFunctions scans for custom functions/RPC installed by templates (Supabase-style)
func (h *APIDiscoveryHandler) getTemplateCustomFunctions(projectID int, templateName string) []APIRoute {
	// This would scan the database for stored procedures/functions that match the template
	// For now, return template-specific custom functions
	
	switch templateName {
	case "photoportfolio":
		return []APIRoute{
			{
				Method:      "POST",
				Path:        "/rpc/get_portfolio_stats",
				Description: "Custom functie: Portfolio statistieken ophalen",
				Category:    "PhotoPortfolio Functions", 
				RequiresAuth: false,
				Source:      "template",
				Parameters: []APIParameter{
					{Name: "period", Type: "string", Required: false, Description: "Time period (week, month, year)"},
				},
				Example: &APIExample{
					Curl:       "curl -X POST -H \"Content-Type: application/json\" -d '{\"period\":\"month\"}' \"{{baseURL}}/rpc/get_portfolio_stats\"",
					JavaScript: "fetch('{{baseURL}}/rpc/get_portfolio_stats', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({period:'month'}) })",
				},
			},
		}
	default:
		return []APIRoute{}
	}
}

// hasColumn helper function to check if table has a specific column
func (h *APIDiscoveryHandler) hasColumn(table TableSchema, columnName string) bool {
	for _, col := range table.Columns {
		if strings.ToLower(col.Name) == strings.ToLower(columnName) {
			return true
		}
	}
	return false
}