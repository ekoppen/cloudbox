package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/services"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// CompatibilityHandler handles CloudBox SDK compatibility checks
type CompatibilityHandler struct {
	db            *gorm.DB
	cfg           *config.Config
	githubService *services.GitHubService
}

// NewCompatibilityHandler creates a new compatibility handler
func NewCompatibilityHandler(db *gorm.DB, cfg *config.Config) *CompatibilityHandler {
	return &CompatibilityHandler{
		db:            db,
		cfg:           cfg,
		githubService: services.NewGitHubService(db),
	}
}

// CompatibilityCheckRequest represents a request to check repository compatibility
type CompatibilityCheckRequest struct {
	RepoURL  string `json:"repo_url" binding:"required"`
	Branch   string `json:"branch"`
	SSHKeyID *uint  `json:"ssh_key_id"`
}

// CompatibilityCheckResponse represents the compatibility check result
type CompatibilityCheckResponse struct {
	IsCompatible      bool                   `json:"is_compatible"`
	CompatibilityScore int                   `json:"compatibility_score"` // 0-100
	
	// CloudBox SDK detection
	HasCloudBoxSDK    bool   `json:"has_cloudbox_sdk"`
	SDKVersion        string `json:"sdk_version"`
	SDKPackageName    string `json:"sdk_package_name"`
	
	// Project analysis
	FrameworkType     string `json:"framework_type"`
	PackageManager    string `json:"package_manager"`
	
	// CloudBox configuration
	HasCloudBoxConfig bool                   `json:"has_cloudbox_config"`
	ConfigFile        string                 `json:"config_file"`
	DetectedConfig    map[string]interface{} `json:"detected_config"`
	
	// Environment variables
	EnvVariables      []string `json:"env_variables"`
	RequiredEnvVars   []string `json:"required_env_vars"`
	MissingEnvVars    []string `json:"missing_env_vars"`
	
	// Commands
	InstallCommands   []string `json:"install_commands"`
	BuildCommands     []string `json:"build_commands"`
	StartCommands     []string `json:"start_commands"`
	
	// Issues and recommendations
	Issues            []string `json:"issues"`
	Warnings          []string `json:"warnings"`
	Recommendations   []string `json:"recommendations"`
	
	// File analysis
	AnalyzedFiles     []string                      `json:"analyzed_files"`
	FileAnalysis      map[string]FileAnalysisResult `json:"file_analysis"`
	
	CheckedAt         time.Time `json:"checked_at"`
}

// FileAnalysisResult represents analysis of a specific file
type FileAnalysisResult struct {
	FileType        string                 `json:"file_type"`
	HasCloudBoxUsage bool                  `json:"has_cloudbox_usage"`
	ImportStatements []string              `json:"import_statements"`
	ConfigOptions   map[string]interface{} `json:"config_options"`
	Issues          []string               `json:"issues"`
}

// CheckRepositoryCompatibility checks if a repository is compatible with CloudBox
func (h *CompatibilityHandler) CheckRepositoryCompatibility(c *gin.Context) {
	var req CompatibilityCheckRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set default branch
	if req.Branch == "" {
		req.Branch = "main"
	}

	// Perform compatibility check
	result, err := h.performCompatibilityCheck(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check compatibility",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// CheckGitHubRepositoryCompatibility checks compatibility of an existing GitHub repository
func (h *CompatibilityHandler) CheckGitHubRepositoryCompatibility(c *gin.Context) {
	repoID := c.Param("id")
	
	var githubRepo models.GitHubRepository
	if err := h.db.First(&githubRepo, repoID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	// Check if we have a recent compatibility check
	var existingCheck models.CloudBoxCompatibility
	if err := h.db.Where("github_repository_id = ?", githubRepo.ID).
		Where("checked_at > ?", time.Now().Add(-24*time.Hour)).
		First(&existingCheck).Error; err == nil {
		// Return cached result if less than 24 hours old
		response := h.convertModelToResponse(existingCheck)
		c.JSON(http.StatusOK, response)
		return
	}

	// Perform new compatibility check
	req := CompatibilityCheckRequest{
		RepoURL: githubRepo.CloneURL,
		Branch:  githubRepo.Branch,
	}

	result, err := h.performCompatibilityCheck(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to check compatibility",
			"details": err.Error(),
		})
		return
	}

	// Save result to database
	h.saveCompatibilityResult(githubRepo.ID, result)

	c.JSON(http.StatusOK, result)
}

// performCompatibilityCheck performs the actual compatibility analysis
func (h *CompatibilityHandler) performCompatibilityCheck(req CompatibilityCheckRequest) (*CompatibilityCheckResponse, error) {
	// Clone repository (temporarily) for analysis
	repoFiles, err := h.cloneAndAnalyzeRepository(req.RepoURL, req.Branch, req.SSHKeyID)
	if err != nil {
		return nil, fmt.Errorf("failed to analyze repository: %v", err)
	}

	response := &CompatibilityCheckResponse{
		CheckedAt:       time.Now(),
		AnalyzedFiles:   []string{},
		FileAnalysis:    make(map[string]FileAnalysisResult),
		Issues:          []string{},
		Warnings:        []string{},
		Recommendations: []string{},
		EnvVariables:    []string{},
		RequiredEnvVars: []string{},
		MissingEnvVars:  []string{},
		InstallCommands: []string{},
		BuildCommands:   []string{},
		StartCommands:   []string{},
	}

	// Analyze package.json for CloudBox SDK
	h.analyzePackageJSON(repoFiles, response)
	
	// Analyze source code for CloudBox usage
	h.analyzeSourceCode(repoFiles, response)
	
	// Analyze configuration files
	h.analyzeConfigFiles(repoFiles, response)
	
	// Analyze environment variables
	h.analyzeEnvironmentVariables(repoFiles, response)
	
	// Generate compatibility score and recommendations
	h.calculateCompatibilityScore(response)
	h.generateRecommendations(response)

	return response, nil
}

// analyzePackageJSON analyzes package.json for CloudBox SDK dependency
func (h *CompatibilityHandler) analyzePackageJSON(files map[string]string, response *CompatibilityCheckResponse) {
	packageJSON, exists := files["package.json"]
	if !exists {
		response.Issues = append(response.Issues, "No package.json found - not a Node.js project")
		return
	}

	response.AnalyzedFiles = append(response.AnalyzedFiles, "package.json")

	var packageData map[string]interface{}
	if err := json.Unmarshal([]byte(packageJSON), &packageData); err != nil {
		response.Issues = append(response.Issues, "Invalid package.json format")
		return
	}

	// Check dependencies for CloudBox SDK
	if deps, ok := packageData["dependencies"].(map[string]interface{}); ok {
		for depName, version := range deps {
			if strings.Contains(depName, "cloudbox") {
				response.HasCloudBoxSDK = true
				response.SDKPackageName = depName
				if v, ok := version.(string); ok {
					response.SDKVersion = v
				}
				break
			}
		}
	}

	// Check devDependencies as well
	if devDeps, ok := packageData["devDependencies"].(map[string]interface{}); ok {
		for depName, version := range devDeps {
			if strings.Contains(depName, "cloudbox") {
				response.HasCloudBoxSDK = true
				response.SDKPackageName = depName
				if v, ok := version.(string); ok {
					response.SDKVersion = v
				}
				break
			}
		}
	}

	// Detect package manager
	if _, exists := files["yarn.lock"]; exists {
		response.PackageManager = "yarn"
	} else if _, exists := files["pnpm-lock.yaml"]; exists {
		response.PackageManager = "pnpm"
	} else {
		response.PackageManager = "npm"
	}

	// Detect framework type
	if deps, ok := packageData["dependencies"].(map[string]interface{}); ok {
		if _, exists := deps["react"]; exists {
			response.FrameworkType = "react"
			if _, exists := deps["next"]; exists {
				response.FrameworkType = "nextjs"
			}
		} else if _, exists := deps["vue"]; exists {
			response.FrameworkType = "vue"
			if _, exists := deps["nuxt"]; exists {
				response.FrameworkType = "nuxtjs"
			}
		} else if _, exists := deps["@angular/core"]; exists {
			response.FrameworkType = "angular"
		} else if _, exists := deps["svelte"]; exists {
			response.FrameworkType = "svelte"
			if _, exists := deps["@sveltejs/kit"]; exists {
				response.FrameworkType = "sveltekit"
			}
		}
	}

	// Extract scripts
	if scripts, ok := packageData["scripts"].(map[string]interface{}); ok {
		if build, exists := scripts["build"]; exists {
			if buildCmd, ok := build.(string); ok {
				response.BuildCommands = append(response.BuildCommands, buildCmd)
			}
		}
		if start, exists := scripts["start"]; exists {
			if startCmd, ok := start.(string); ok {
				response.StartCommands = append(response.StartCommands, startCmd)
			}
		}
		if dev, exists := scripts["dev"]; exists {
			if devCmd, ok := dev.(string); ok {
				response.StartCommands = append(response.StartCommands, devCmd)
			}
		}
	}

	// Install command based on package manager
	switch response.PackageManager {
	case "yarn":
		response.InstallCommands = append(response.InstallCommands, "yarn install")
	case "pnpm":
		response.InstallCommands = append(response.InstallCommands, "pnpm install")
	default:
		response.InstallCommands = append(response.InstallCommands, "npm install")
	}

	// Analyze package.json for CloudBox usage patterns
	response.FileAnalysis["package.json"] = FileAnalysisResult{
		FileType:         "package.json",
		HasCloudBoxUsage: response.HasCloudBoxSDK,
		ImportStatements: []string{},
		ConfigOptions:    packageData,
		Issues:           []string{},
	}
}

// analyzeSourceCode analyzes source code files for CloudBox SDK usage
func (h *CompatibilityHandler) analyzeSourceCode(files map[string]string, response *CompatibilityCheckResponse) {
	// Patterns to look for CloudBox usage
	cloudboxImportPatterns := []string{
		`import.*cloudbox`,
		`require.*cloudbox`,
		`from.*cloudbox`,
		`import.*@ekoppen/cloudbox-sdk`,
		`require.*@ekoppen/cloudbox-sdk`,
	}

	configPatterns := []string{
		`CLOUDBOX_ENDPOINT`,
		`CLOUDBOX_API_KEY`,
		`CLOUDBOX_PROJECT_ID`,
		`cloudbox\.init`,
		`new CloudBox`,
		`CloudBoxClient`,
	}

	// Analyze JavaScript/TypeScript files
	for filename, content := range files {
		if h.isSourceFile(filename) {
			response.AnalyzedFiles = append(response.AnalyzedFiles, filename)
			
			analysis := FileAnalysisResult{
				FileType:         h.getFileType(filename),
				HasCloudBoxUsage: false,
				ImportStatements: []string{},
				ConfigOptions:    make(map[string]interface{}),
				Issues:           []string{},
			}

			// Check for CloudBox imports
			for _, pattern := range cloudboxImportPatterns {
				if matched, _ := regexp.MatchString(pattern, content); matched {
					analysis.HasCloudBoxUsage = true
					// Extract import statement
					re := regexp.MustCompile(pattern)
					matches := re.FindAllString(content, -1)
					analysis.ImportStatements = append(analysis.ImportStatements, matches...)
				}
			}

			// Check for CloudBox configuration
			for _, pattern := range configPatterns {
				if matched, _ := regexp.MatchString(pattern, content); matched {
					analysis.HasCloudBoxUsage = true
					response.HasCloudBoxConfig = true
				}
			}

			// Extract environment variable references
			envRegex := regexp.MustCompile(`process\.env\.([A-Z_]+)`)
			envMatches := envRegex.FindAllStringSubmatch(content, -1)
			for _, match := range envMatches {
				if len(match) > 1 {
					envVar := match[1]
					if !containsString(response.EnvVariables, envVar) {
						response.EnvVariables = append(response.EnvVariables, envVar)
					}
				}
			}

			if analysis.HasCloudBoxUsage {
				response.HasCloudBoxConfig = true
			}

			response.FileAnalysis[filename] = analysis
		}
	}
}

// analyzeConfigFiles analyzes configuration files for CloudBox setup
func (h *CompatibilityHandler) analyzeConfigFiles(files map[string]string, response *CompatibilityCheckResponse) {
	configFiles := []string{
		".env",
		".env.local",
		".env.example",
		"cloudbox.config.js",
		"cloudbox.json",
		"next.config.js",
		"nuxt.config.js",
		"vite.config.js",
		"svelte.config.js",
	}

	for _, configFile := range configFiles {
		if content, exists := files[configFile]; exists {
			response.AnalyzedFiles = append(response.AnalyzedFiles, configFile)
			
			if strings.HasPrefix(configFile, ".env") {
				h.analyzeEnvFile(content, response)
			} else if strings.Contains(configFile, "cloudbox") {
				response.HasCloudBoxConfig = true
				response.ConfigFile = configFile
				h.parseCloudBoxConfig(content, response)
			}
		}
	}
}

// analyzeEnvFile analyzes environment files
func (h *CompatibilityHandler) analyzeEnvFile(content string, response *CompatibilityCheckResponse) {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}
		
		if strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				envVar := strings.TrimSpace(parts[0])
				if !containsString(response.EnvVariables, envVar) {
					response.EnvVariables = append(response.EnvVariables, envVar)
				}
				
				// Check for CloudBox-specific variables
				if strings.HasPrefix(envVar, "CLOUDBOX_") {
					response.HasCloudBoxConfig = true
					if !containsString(response.RequiredEnvVars, envVar) {
						response.RequiredEnvVars = append(response.RequiredEnvVars, envVar)
					}
				}
			}
		}
	}
}

// parseCloudBoxConfig parses CloudBox configuration files
func (h *CompatibilityHandler) parseCloudBoxConfig(content string, response *CompatibilityCheckResponse) {
	// Try to parse as JSON first
	var config map[string]interface{}
	if err := json.Unmarshal([]byte(content), &config); err == nil {
		response.DetectedConfig = config
		return
	}

	// If JSON parsing fails, try to extract configuration from JS file
	// This is a simplified parser for basic config extraction
	lines := strings.Split(content, "\n")
	config = make(map[string]interface{})
	
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Look for simple key-value pairs
		if strings.Contains(line, ":") && !strings.HasPrefix(line, "//") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.Trim(strings.TrimSpace(parts[0]), `"'`)
				value := strings.Trim(strings.TrimSpace(parts[1]), `"',`)
				config[key] = value
			}
		}
	}
	
	response.DetectedConfig = config
}

// analyzeEnvironmentVariables analyzes environment variable requirements
func (h *CompatibilityHandler) analyzeEnvironmentVariables(files map[string]string, response *CompatibilityCheckResponse) {
	// Standard CloudBox environment variables
	standardCloudBoxVars := []string{
		"CLOUDBOX_ENDPOINT",
		"CLOUDBOX_API_KEY", 
		"CLOUDBOX_PROJECT_ID",
		"CLOUDBOX_PROJECT_SLUG",
	}

	// Check which standard variables are used/configured
	for _, stdVar := range standardCloudBoxVars {
		if containsString(response.EnvVariables, stdVar) {
			if !containsString(response.RequiredEnvVars, stdVar) {
				response.RequiredEnvVars = append(response.RequiredEnvVars, stdVar)
			}
		}
	}

	// If CloudBox SDK is used but no env vars found, flag as missing
	if response.HasCloudBoxSDK && len(response.RequiredEnvVars) == 0 {
		response.MissingEnvVars = standardCloudBoxVars
		response.Issues = append(response.Issues, "CloudBox SDK found but no environment variables configured")
	}
}

// calculateCompatibilityScore calculates overall compatibility score
func (h *CompatibilityHandler) calculateCompatibilityScore(response *CompatibilityCheckResponse) {
	score := 0

	// Base score for having package.json
	if containsString(response.AnalyzedFiles, "package.json") {
		score += 10
	}

	// CloudBox SDK presence (major factor)
	if response.HasCloudBoxSDK {
		score += 40
		if response.SDKVersion != "" {
			score += 10
		}
	}

	// CloudBox configuration
	if response.HasCloudBoxConfig {
		score += 20
		if len(response.RequiredEnvVars) > 0 {
			score += 10
		}
	}

	// Framework detection
	if response.FrameworkType != "" {
		score += 10
	}

	// Package manager detection
	if response.PackageManager != "" {
		score += 5
	}

	// Build commands
	if len(response.BuildCommands) > 0 {
		score += 5
	}

	// Deduct points for issues
	score -= len(response.Issues) * 5

	// Ensure score is within bounds
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	response.CompatibilityScore = score
	response.IsCompatible = score >= 50 // 50% minimum for compatibility
}

// generateRecommendations generates recommendations for improving compatibility
func (h *CompatibilityHandler) generateRecommendations(response *CompatibilityCheckResponse) {
	if !response.HasCloudBoxSDK {
		response.Recommendations = append(response.Recommendations, 
			"Install CloudBox SDK: npm install @ekoppen/cloudbox-sdk")
	}

	if !response.HasCloudBoxConfig {
		response.Recommendations = append(response.Recommendations,
			"Create CloudBox configuration file or add environment variables")
	}

	if len(response.RequiredEnvVars) == 0 && response.HasCloudBoxSDK {
		response.Recommendations = append(response.Recommendations,
			"Configure CloudBox environment variables (CLOUDBOX_ENDPOINT, CLOUDBOX_API_KEY, CLOUDBOX_PROJECT_ID)")
	}

	if len(response.BuildCommands) == 0 {
		response.Recommendations = append(response.Recommendations,
			"Add build script to package.json")
	}

	if len(response.StartCommands) == 0 {
		response.Recommendations = append(response.Recommendations,
			"Add start script to package.json")
	}

	if response.FrameworkType == "" {
		response.Recommendations = append(response.Recommendations,
			"Consider using a supported framework (React, Vue, Next.js, Nuxt, Svelte, Angular)")
	}

	if len(response.Issues) > 0 {
		response.Recommendations = append(response.Recommendations,
			"Fix identified issues to improve compatibility")
	}
}

// Helper functions
func (h *CompatibilityHandler) isSourceFile(filename string) bool {
	extensions := []string{".js", ".ts", ".jsx", ".tsx", ".vue", ".svelte"}
	for _, ext := range extensions {
		if strings.HasSuffix(filename, ext) {
			return true
		}
	}
	return false
}

func (h *CompatibilityHandler) getFileType(filename string) string {
	if strings.HasSuffix(filename, ".ts") || strings.HasSuffix(filename, ".tsx") {
		return "typescript"
	}
	if strings.HasSuffix(filename, ".js") || strings.HasSuffix(filename, ".jsx") {
		return "javascript"
	}
	if strings.HasSuffix(filename, ".vue") {
		return "vue"
	}
	if strings.HasSuffix(filename, ".svelte") {
		return "svelte"
	}
	return "unknown"
}

func containsString(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// cloneAndAnalyzeRepository clones repository and extracts files for analysis
func (h *CompatibilityHandler) cloneAndAnalyzeRepository(repoURL, branch string, sshKeyID *uint) (map[string]string, error) {
	// TODO: Implement actual repository cloning and file extraction
	// This would typically involve:
	// 1. Cloning the repository to a temporary directory
	// 2. Reading relevant files
	// 3. Cleaning up temporary directory
	
	// For now, return mock data
	mockFiles := map[string]string{
		"package.json": `{
			"name": "example-app",
			"dependencies": {
				"@ekoppen/cloudbox-sdk": "^1.0.0",
				"react": "^18.0.0"
			},
			"scripts": {
				"build": "npm run build",
				"start": "npm start"
			}
		}`,
		".env.example": `CLOUDBOX_ENDPOINT=http://localhost:8080
CLOUDBOX_API_KEY=your-api-key
CLOUDBOX_PROJECT_ID=1`,
		"src/main.js": `import { CloudBoxClient } from '@ekoppen/cloudbox-sdk';
const cloudbox = new CloudBoxClient({
	projectId: process.env.CLOUDBOX_PROJECT_ID,
	apiKey: process.env.CLOUDBOX_API_KEY
});`,
	}
	
	return mockFiles, nil
}

// saveCompatibilityResult saves the compatibility check result to database
func (h *CompatibilityHandler) saveCompatibilityResult(githubRepoID uint, result *CompatibilityCheckResponse) {
	compatibility := models.CloudBoxCompatibility{
		GitHubRepositoryID: githubRepoID,
		IsCompatible:       result.IsCompatible,
		HasCloudBoxSDK:     result.HasCloudBoxSDK,
		SDKVersion:         result.SDKVersion,
		PackageManager:     result.PackageManager,
		FrameworkType:      result.FrameworkType,
		HasCloudBoxConfig:  result.HasCloudBoxConfig,
		ConfigFile:         result.ConfigFile,
		DetectedConfig:     result.DetectedConfig,
		EnvVariables:       pq.StringArray(result.EnvVariables),
		RequiredEnvVars:    pq.StringArray(result.RequiredEnvVars),
		InstallCommands:    pq.StringArray(result.InstallCommands),
		BuildCommands:      pq.StringArray(result.BuildCommands),
		StartCommands:      pq.StringArray(result.StartCommands),
		Issues:             result.Issues,
		Recommendations:    result.Recommendations,
		CheckedAt:          result.CheckedAt,
		CheckVersion:       "1.0.0",
	}

	// Upsert the compatibility record
	h.db.Where("github_repository_id = ?", githubRepoID).
		Assign(compatibility).
		FirstOrCreate(&compatibility)
}

// convertModelToResponse converts database model to API response
func (h *CompatibilityHandler) convertModelToResponse(model models.CloudBoxCompatibility) *CompatibilityCheckResponse {
	return &CompatibilityCheckResponse{
		IsCompatible:      model.IsCompatible,
		CompatibilityScore: h.calculateScoreFromModel(model),
		HasCloudBoxSDK:    model.HasCloudBoxSDK,
		SDKVersion:        model.SDKVersion,
		FrameworkType:     model.FrameworkType,
		PackageManager:    model.PackageManager,
		HasCloudBoxConfig: model.HasCloudBoxConfig,
		ConfigFile:        model.ConfigFile,
		DetectedConfig:    model.DetectedConfig,
		EnvVariables:      []string(model.EnvVariables),
		RequiredEnvVars:   []string(model.RequiredEnvVars),
		InstallCommands:   []string(model.InstallCommands),
		BuildCommands:     []string(model.BuildCommands),
		StartCommands:     []string(model.StartCommands),
		Issues:            model.Issues,
		Recommendations:   model.Recommendations,
		CheckedAt:         model.CheckedAt,
	}
}

// calculateScoreFromModel calculates compatibility score from database model
func (h *CompatibilityHandler) calculateScoreFromModel(model models.CloudBoxCompatibility) int {
	score := 0
	
	if model.HasCloudBoxSDK {
		score += 50
	}
	if model.HasCloudBoxConfig {
		score += 20
	}
	if model.FrameworkType != "" {
		score += 10
	}
	if len(model.BuildCommands) > 0 {
		score += 10
	}
	if len(model.StartCommands) > 0 {
		score += 10
	}
	
	// Deduct for issues
	score -= len(model.Issues) * 5
	
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	
	return score
}