package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// TemplateDeploymentHandler handles template-based project deployments
type TemplateDeploymentHandler struct {
	db             *gorm.DB
	cfg            *config.Config
	githubService  *services.GitHubService
	templateHandler *TemplateHandler
}

// NewTemplateDeploymentHandler creates a new template deployment handler
func NewTemplateDeploymentHandler(db *gorm.DB, cfg *config.Config) *TemplateDeploymentHandler {
	return &TemplateDeploymentHandler{
		db:              db,
		cfg:             cfg,
		githubService:   services.NewGitHubService(db),
		templateHandler: NewTemplateHandler(db, cfg),
	}
}

// TemplateDeploymentRequest represents a request to deploy a template to GitHub
type TemplateDeploymentRequest struct {
	// Template configuration
	TemplateName string                 `json:"template_name" binding:"required"` // photoportfolio, blog, ecommerce
	Variables    map[string]interface{} `json:"variables"`                        // Template variables for customization
	
	// GitHub repository configuration
	GitHubRepo struct {
		Name        string `json:"name" binding:"required"`        // Repository name
		Description string `json:"description"`                    // Repository description
		IsPrivate   bool   `json:"is_private"`                     // Private or public repository
		Owner       string `json:"owner"`                          // GitHub username/org (optional, defaults to authenticated user)
	} `json:"github_repo" binding:"required"`
	
	// Deployment configuration
	DeploymentConfig struct {
		AppPort        int                    `json:"app_port"`        // Application port (default: 3000)
		BuildCommand   string                 `json:"build_command"`   // Build command
		StartCommand   string                 `json:"start_command"`   // Start command
		Environment    map[string]interface{} `json:"environment"`     // Environment variables
		AutoDeploy     bool                   `json:"auto_deploy"`     // Auto-deploy on push
		CustomDomain   string                 `json:"custom_domain"`   // Custom domain for deployment
		SSLEnabled     bool                   `json:"ssl_enabled"`     // Enable SSL/HTTPS
	} `json:"deployment_config"`
}

// TemplateDeploymentResponse represents the deployment result
type TemplateDeploymentResponse struct {
	// Template setup results
	TemplateResults map[string]interface{} `json:"template_results"`
	
	// GitHub repository information
	GitHubRepo struct {
		ID          uint   `json:"id"`
		Name        string `json:"name"`
		FullName    string `json:"full_name"`
		CloneURL    string `json:"clone_url"`
		IsPrivate   bool   `json:"is_private"`
		Description string `json:"description"`
	} `json:"github_repo"`
	
	// Deployment information
	Deployment struct {
		ID           uint      `json:"id"`
		Status       string    `json:"status"`
		Domain       string    `json:"domain"`
		BuildLogs    string    `json:"build_logs,omitempty"`
		CreatedAt    time.Time `json:"created_at"`
	} `json:"deployment"`
	
	// Project configuration
	ProjectConfig struct {
		APIKey      string                 `json:"api_key"`
		ProjectID   uint                   `json:"project_id"`
		ProjectSlug string                 `json:"project_slug"`
		Environment map[string]interface{} `json:"environment"`
	} `json:"project_config"`
}

// ListTemplateDeployments returns template deployments for a project
func (h *TemplateDeploymentHandler) ListTemplateDeployments(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var deployments []models.TemplateDeployment
	if err := h.db.Where("project_id = ?", projectID).
		Preload("GitHubRepository").
		Preload("Deployment").
		Order("created_at DESC").
		Find(&deployments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch deployments"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

// CreateTemplateDeployment creates a new template-based deployment
func (h *TemplateDeploymentHandler) CreateTemplateDeployment(c *gin.Context) {
	projectID := c.GetUint("project_id")
	if projectID == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req TemplateDeploymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get project details
	var project models.Project
	if err := h.db.First(&project, projectID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
		return
	}

	// Step 1: Setup CloudBox template (collections, buckets, etc.)
	templateResults, err := h.setupCloudBoxTemplate(projectID, req.TemplateName, req.Variables)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to setup CloudBox template",
			"details": err.Error(),
		})
		return
	}

	// Step 2: Create GitHub repository from template
	githubRepo, err := h.createGitHubRepositoryFromTemplate(
		req.TemplateName,
		req.GitHubRepo,
		req.Variables,
		project,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create GitHub repository",
			"details": err.Error(),
		})
		return
	}

	// Step 3: Create deployment configuration
	deployment, err := h.createDeploymentConfig(
		projectID,
		githubRepo.ID,
		req.DeploymentConfig,
		project,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create deployment",
			"details": err.Error(),
		})
		return
	}

	// Step 4: Save template deployment record
	templateDeployment := models.TemplateDeployment{
		ProjectID:          projectID,
		TemplateName:       req.TemplateName,
		Variables:          req.Variables,
		GitHubRepositoryID: githubRepo.ID,
		DeploymentID:       deployment.ID,
		Status:             "created",
		CreatedAt:          time.Now(),
	}
	
	if err := h.db.Create(&templateDeployment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save deployment record"})
		return
	}

	// Step 5: Generate API key for the deployment
	apiKey, err := h.generateDeploymentAPIKey(projectID, githubRepo.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to generate API key",
			"details": err.Error(),
		})
		return
	}

	// Step 6: Trigger initial deployment if auto-deploy is enabled
	if req.DeploymentConfig.AutoDeploy {
		go h.triggerInitialDeployment(deployment.ID, githubRepo.ID)
	}

	// Return response
	response := TemplateDeploymentResponse{
		TemplateResults: templateResults,
		GitHubRepo: struct {
			ID          uint   `json:"id"`
			Name        string `json:"name"`
			FullName    string `json:"full_name"`
			CloneURL    string `json:"clone_url"`
			IsPrivate   bool   `json:"is_private"`
			Description string `json:"description"`
		}{
			ID:          githubRepo.ID,
			Name:        githubRepo.Name,
			FullName:    githubRepo.FullName,
			CloneURL:    githubRepo.CloneURL,
			IsPrivate:   githubRepo.IsPrivate,
			Description: githubRepo.Description,
		},
		Deployment: struct {
			ID           uint      `json:"id"`
			Status       string    `json:"status"`
			Domain       string    `json:"domain"`
			BuildLogs    string    `json:"build_logs,omitempty"`
			CreatedAt    time.Time `json:"created_at"`
		}{
			ID:        deployment.ID,
			Status:    deployment.Status,
			Domain:    deployment.Domain,
			CreatedAt: deployment.CreatedAt,
		},
		ProjectConfig: struct {
			APIKey      string                 `json:"api_key"`
			ProjectID   uint                   `json:"project_id"`
			ProjectSlug string                 `json:"project_slug"`
			Environment map[string]interface{} `json:"environment"`
		}{
			APIKey:      apiKey,
			ProjectID:   projectID,
			ProjectSlug: project.Slug,
			Environment: h.generateEnvironmentVariables(project, apiKey, req.DeploymentConfig),
		},
	}

	c.JSON(http.StatusCreated, response)
}

// setupCloudBoxTemplate sets up the CloudBox template (collections, buckets, etc.)
func (h *TemplateDeploymentHandler) setupCloudBoxTemplate(projectID uint, templateName string, variables map[string]interface{}) (map[string]interface{}, error) {
	// Get the template definition
	var templateDef TemplateDefinition
	switch templateName {
	case "photoportfolio":
		templateDef = h.templateHandler.getPhotoPortfolioTemplate()
	case "blog":
		templateDef = h.getBlogTemplate()
	case "ecommerce":
		templateDef = h.getEcommerceTemplate()
	case "saas":
		templateDef = h.getSaaSTemplate()
	case "portfolio":
		templateDef = h.getPortfolioTemplate()
	default:
		return nil, fmt.Errorf("unknown template: %s", templateName)
	}

	// Apply variables to template
	if variables != nil {
		templateDef = h.applyVariablesToTemplate(templateDef, variables)
	}

	// Setup collections
	results := make(map[string]interface{})
	
	for _, collectionTemplate := range templateDef.Collections {
		result, err := h.templateHandler.setupCollection(projectID, collectionTemplate)
		if err != nil {
			results[collectionTemplate.Name] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
			continue
		}
		results[collectionTemplate.Name] = map[string]interface{}{
			"status":     "success",
			"collection": result,
		}
	}
	
	// Setup storage buckets
	for _, bucketTemplate := range templateDef.Buckets {
		result, err := h.templateHandler.setupBucket(projectID, bucketTemplate)
		if err != nil {
			results[bucketTemplate.Name] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
			continue
		}
		results[bucketTemplate.Name] = map[string]interface{}{
			"status": "success",
			"bucket":  result,
		}
	}
	
	// Setup CORS configuration
	if templateDef.CORS != nil {
		result, err := h.templateHandler.setupCORS(projectID, *templateDef.CORS)
		if err != nil {
			results["cors"] = map[string]interface{}{
				"status": "error",
				"error":  err.Error(),
			}
		} else {
			results["cors"] = map[string]interface{}{
				"status": "success",
				"cors":   result,
			}
		}
	}

	return results, nil
}

// createGitHubRepositoryFromTemplate creates a GitHub repository with template code
func (h *TemplateDeploymentHandler) createGitHubRepositoryFromTemplate(
	templateName string,
	githubConfig struct {
		Name        string `json:"name" binding:"required"`
		Description string `json:"description"`
		IsPrivate   bool   `json:"is_private"`
		Owner       string `json:"owner"`
	},
	variables map[string]interface{},
	project models.Project,
) (*models.GitHubRepository, error) {
	// Generate template repository content
	templateContent, err := h.generateTemplateContent(templateName, variables, project)
	if err != nil {
		return nil, fmt.Errorf("failed to generate template content: %v", err)
	}

	// GitHub repository data would be used for API calls
	// (removed unused variable)

	// TODO: Implement actual GitHub API calls to create repository
	// For now, create a local record
	githubRepo := &models.GitHubRepository{
		Name:        githubConfig.Name,
		FullName:    fmt.Sprintf("%s/%s", githubConfig.Owner, githubConfig.Name),
		CloneURL:    fmt.Sprintf("https://github.com/%s/%s.git", githubConfig.Owner, githubConfig.Name),
		IsPrivate:   githubConfig.IsPrivate,
		Description: githubConfig.Description,
		Branch:      "main",
	}

	if err := h.db.Create(githubRepo).Error; err != nil {
		return nil, fmt.Errorf("failed to save repository record: %v", err)
	}

	// TODO: Push template content to GitHub repository
	_ = templateContent // Placeholder for actual implementation

	return githubRepo, nil
}

// createDeploymentConfig creates deployment configuration
func (h *TemplateDeploymentHandler) createDeploymentConfig(
	projectID uint,
	githubRepoID uint,
	deployConfig struct {
		AppPort        int                    `json:"app_port"`
		BuildCommand   string                 `json:"build_command"`
		StartCommand   string                 `json:"start_command"`
		Environment    map[string]interface{} `json:"environment"`
		AutoDeploy     bool                   `json:"auto_deploy"`
		CustomDomain   string                 `json:"custom_domain"`
		SSLEnabled     bool                   `json:"ssl_enabled"`
	},
	project models.Project,
) (*models.Deployment, error) {
	// Set defaults
	appPort := deployConfig.AppPort
	if appPort == 0 {
		appPort = 3000
	}

	buildCommand := deployConfig.BuildCommand
	if buildCommand == "" {
		buildCommand = "npm run build"
	}

	startCommand := deployConfig.StartCommand
	if startCommand == "" {
		startCommand = "npm start"
	}

	// Create deployment record
	deployment := &models.Deployment{
		Name:             fmt.Sprintf("%s-deployment", project.Slug),
		Description:      fmt.Sprintf("Template deployment for %s", project.Name),
		Version:          "1.0.0",
		Domain:           fmt.Sprintf("%s.cloudbox.dev", project.Slug),
		Subdomain:        project.Slug,
		Port:             3000, // Default port
		Status:           "pending",
		Branch:           "main",
		BuildCommand:     buildCommand,
		StartCommand:     startCommand,
		Environment:      deployConfig.Environment,
	}

	if err := h.db.Create(deployment).Error; err != nil {
		return nil, fmt.Errorf("failed to create deployment: %v", err)
	}

	return deployment, nil
}

// generateDeploymentAPIKey generates an API key for the deployment
func (h *TemplateDeploymentHandler) generateDeploymentAPIKey(projectID uint, repoName string) (string, error) {
	// Generate API key
	apiKey := fmt.Sprintf("cb_%s_%d", generateTemplateRandomString(32), time.Now().Unix())
	
	// Hash the API key
	hashedKey, err := hashAPIKey(apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to hash API key: %v", err)
	}

	// Create API key record
	apiKeyRecord := models.APIKey{
		Name:      fmt.Sprintf("Template Deployment - %s", repoName),
		KeyHash:   hashedKey,
		ProjectID: projectID,
		Permissions: []string{"read", "write"},
		LastUsedAt: nil,
		CreatedAt:  time.Now(),
	}

	if err := h.db.Create(&apiKeyRecord).Error; err != nil {
		return "", fmt.Errorf("failed to save API key: %v", err)
	}

	return apiKey, nil
}

// generateEnvironmentVariables generates environment variables for the deployment
func (h *TemplateDeploymentHandler) generateEnvironmentVariables(
	project models.Project,
	apiKey string,
	deployConfig struct {
		AppPort        int                    `json:"app_port"`
		BuildCommand   string                 `json:"build_command"`
		StartCommand   string                 `json:"start_command"`
		Environment    map[string]interface{} `json:"environment"`
		AutoDeploy     bool                   `json:"auto_deploy"`
		CustomDomain   string                 `json:"custom_domain"`
		SSLEnabled     bool                   `json:"ssl_enabled"`
	},
) map[string]interface{} {
	env := map[string]interface{}{
		"CLOUDBOX_ENDPOINT":     h.cfg.BaseURL,
		"CLOUDBOX_PROJECT_ID":   strconv.Itoa(int(project.ID)),
		"CLOUDBOX_PROJECT_SLUG": project.Slug,
		"CLOUDBOX_API_KEY":      apiKey,
		"NODE_ENV":              "production",
		"PORT":                  strconv.Itoa(deployConfig.AppPort),
	}

	// Add custom environment variables
	if deployConfig.Environment != nil {
		for key, value := range deployConfig.Environment {
			env[key] = value
		}
	}

	return env
}

// triggerInitialDeployment triggers the initial deployment
func (h *TemplateDeploymentHandler) triggerInitialDeployment(deploymentID, githubRepoID uint) {
	// TODO: Implement actual deployment trigger
	// This would typically involve:
	// 1. Cloning the repository
	// 2. Running build commands
	// 3. Deploying to container/server
	// 4. Updating deployment status
	
	// For now, just update status
	h.db.Model(&models.Deployment{}).
		Where("id = ?", deploymentID).
		Update("status", "deploying")
}

// generateTemplateContent generates the actual code content for the template
func (h *TemplateDeploymentHandler) generateTemplateContent(
	templateName string,
	variables map[string]interface{},
	project models.Project,
) (map[string]string, error) {
	content := make(map[string]string)

	switch templateName {
	case "photoportfolio":
		content = h.generatePhotoPortfolioContent(variables, project)
	case "blog":
		content = h.generateBlogContent(variables, project)
	case "ecommerce":
		content = h.generateEcommerceContent(variables, project)
	default:
		return nil, fmt.Errorf("template content not implemented for: %s", templateName)
	}

	return content, nil
}

// Helper functions for other templates (blog, ecommerce, etc.)
func (h *TemplateDeploymentHandler) getBlogTemplate() TemplateDefinition {
	// TODO: Implement blog template definition
	return TemplateDefinition{
		Name:        "blog",
		Version:     "1.0.0",
		Description: "Blog Template with posts, categories, and comments",
		Collections: []CollectionTemplate{
			{
				Name:        "posts",
				Description: "Blog posts",
				Schema: map[string]interface{}{
					"title":     "string",
					"content":   "text",
					"published": "boolean",
					"author":    "string",
				},
			},
		},
		Buckets: []BucketTemplate{
			{
				Name:        "uploads",
				Description: "Blog media uploads",
				MaxFileSize: 10485760,
				IsPublic:    true,
			},
		},
	}
}

func (h *TemplateDeploymentHandler) getEcommerceTemplate() TemplateDefinition {
	// TODO: Implement ecommerce template definition
	return TemplateDefinition{
		Name:        "ecommerce",
		Version:     "1.0.0",
		Description: "E-commerce Template with products, orders, and customers",
		Collections: []CollectionTemplate{
			{
				Name:        "products",
				Description: "Product catalog",
				Schema: map[string]interface{}{
					"name":  "string",
					"price": "number",
					"stock": "number",
				},
			},
		},
		Buckets: []BucketTemplate{
			{
				Name:        "product-images",
				Description: "Product images",
				MaxFileSize: 5242880,
				IsPublic:    true,
			},
		},
	}
}

func (h *TemplateDeploymentHandler) getSaaSTemplate() TemplateDefinition {
	// TODO: Implement SaaS template definition
	return TemplateDefinition{}
}

func (h *TemplateDeploymentHandler) getPortfolioTemplate() TemplateDefinition {
	// TODO: Implement portfolio template definition
	return TemplateDefinition{}
}

func (h *TemplateDeploymentHandler) applyVariablesToTemplate(template TemplateDefinition, variables map[string]interface{}) TemplateDefinition {
	// TODO: Implement variable substitution in template
	// This would replace placeholders like {{site_name}} with actual values
	return template
}

func (h *TemplateDeploymentHandler) generatePhotoPortfolioContent(variables map[string]interface{}, project models.Project) map[string]string {
	// TODO: Generate actual PhotoPortfolio code files
	// This would generate:
	// - package.json
	// - src/main.js
	// - src/components/*
	// - README.md
	// - docker-compose.yml
	// etc.
	return map[string]string{
		"package.json": `{
			"name": "` + project.Slug + `-photoportfolio",
			"version": "1.0.0",
			"dependencies": {
				"@ekoppen/cloudbox-sdk": "^1.0.0"
			}
		}`,
		"README.md": "# " + project.Name + " Photo Portfolio\n\nGenerated by CloudBox Template System",
	}
}

func (h *TemplateDeploymentHandler) generateBlogContent(variables map[string]interface{}, project models.Project) map[string]string {
	// TODO: Implement blog content generation
	return map[string]string{}
}

func (h *TemplateDeploymentHandler) generateEcommerceContent(variables map[string]interface{}, project models.Project) map[string]string {
	// TODO: Implement ecommerce content generation
	return map[string]string{}
}

// Helper functions
func generateTemplateRandomString(length int) string {
	bytes := make([]byte, length)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

func hashAPIKey(key string) (string, error) {
	// TODO: Implement proper API key hashing (bcrypt)
	return fmt.Sprintf("hashed_%s", key), nil
}