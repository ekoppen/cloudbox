package middleware

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CORSEnvironment represents different deployment environments
type CORSEnvironment string

const (
	Development CORSEnvironment = "development"
	Staging     CORSEnvironment = "staging"
	Production  CORSEnvironment = "production"
)

// EnhancedCORSConfig represents intelligent CORS configuration
type EnhancedCORSConfig struct {
	Environment     CORSEnvironment `json:"environment"`
	AllowedOrigins  []string        `json:"allowed_origins"`
	AllowedHeaders  []string        `json:"allowed_headers"`
	AllowedMethods  []string        `json:"allowed_methods"`
	DynamicPatterns []string        `json:"dynamic_patterns"`
	AutoDetect      bool            `json:"auto_detect"`
	DebugMode       bool            `json:"debug_mode"`
	MaxAge          int             `json:"max_age"`
}

// SmartCORS provides intelligent CORS handling with environment awareness
func SmartCORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		corsConfig := getEnhancedCORSConfig(cfg)
		
		// Add debugging headers if in debug mode
		if corsConfig.DebugMode {
			c.Header("X-CORS-Debug", "true")
			c.Header("X-CORS-Origin-Received", origin)
			c.Header("X-CORS-Environment", string(corsConfig.Environment))
		}
		
		// Always set CORS headers for OPTIONS requests (preflight)
		if c.Request.Method == "OPTIONS" {
			// Multi-tier origin validation for preflight
			if origin != "" {
				if allowed, reason := isOriginAllowedEnhanced(origin, corsConfig); allowed {
					c.Header("Access-Control-Allow-Origin", origin)
					if corsConfig.DebugMode {
						c.Header("X-CORS-Allow-Reason", reason)
					}
				} else {
					// Provide helpful error headers
					setCORSErrorHeaders(c, origin, corsConfig, reason)
				}
			}
			
			// Set comprehensive allowed methods
			allowedMethods := ensureRequiredMethods(corsConfig.AllowedMethods)
			c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			
			// Set comprehensive allowed headers
			allowedHeaders := getUniversalAllowedHeaders(corsConfig)
			c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", corsConfig.MaxAge))
			
			// Add helpful development headers
			if corsConfig.DebugMode {
				c.Header("X-CORS-Preflight", "success")
			}
			
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// For non-OPTIONS requests, check if origin is allowed
		if origin != "" {
			if allowed, reason := isOriginAllowedEnhanced(origin, corsConfig); allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				
				// Set comprehensive headers
				allowedMethods := ensureRequiredMethods(corsConfig.AllowedMethods)
				c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
				
				allowedHeaders := getUniversalAllowedHeaders(corsConfig)
				c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
				
				c.Header("Access-Control-Allow-Credentials", "true")
				c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", corsConfig.MaxAge))
				
				if corsConfig.DebugMode {
					c.Header("X-CORS-Request-Allow-Reason", reason)
				}
			} else {
				// Log and set error headers for debugging
				logCORSViolation(origin, c.Request.URL.Path, reason)
				setCORSErrorHeaders(c, origin, corsConfig, reason)
			}
		}

		c.Next()
	}
}

// getEnhancedCORSConfig creates enhanced CORS configuration from base config
func getEnhancedCORSConfig(cfg *config.Config) *EnhancedCORSConfig {
	// Detect environment from NODE_ENV or default to development
	env := Development
	if nodeEnv := os.Getenv("NODE_ENV"); nodeEnv != "" {
		switch nodeEnv {
		case "production":
			env = Production
		case "staging":
			env = Staging
		default:
			env = Development
		}
	}
	
	// Enable debug mode for development
	debugMode := env == Development || os.Getenv("CORS_DEBUG") == "true"
	
	return &EnhancedCORSConfig{
		Environment:     env,
		AllowedOrigins:  cfg.AllowedOrigins,
		AllowedHeaders:  cfg.AllowedHeaders,
		AllowedMethods:  cfg.AllowedMethods,
		DynamicPatterns: []string{},
		AutoDetect:      env == Development,
		DebugMode:       debugMode,
		MaxAge:          3600,
	}
}

// isOriginAllowedEnhanced provides multi-tier origin validation
func isOriginAllowedEnhanced(origin string, corsConfig *EnhancedCORSConfig) (bool, string) {
	// Check explicit origins first
	for _, allowed := range corsConfig.AllowedOrigins {
		if allowed == "*" || allowed == origin {
			return true, fmt.Sprintf("exact match: %s", allowed)
		}
		
		// Handle wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true, fmt.Sprintf("wildcard subdomain match: %s", allowed)
			}
		}
		
		// Handle localhost wildcard patterns (e.g., http://localhost:*)
		if matchesLocalhostPattern(origin, allowed) {
			return true, fmt.Sprintf("localhost pattern match: %s", allowed)
		}
	}
	
	// Development environment auto-detection
	if corsConfig.Environment == Development && corsConfig.AutoDetect {
		if isLocalhostOrigin(origin) {
			return true, "development auto-detect: localhost origin"
		}
	}
	
	// Check if any allowed origins suggest localhost development is allowed
	if isLocalhostOrigin(origin) {
		for _, allowed := range corsConfig.AllowedOrigins {
			if isLocalhostOrigin(allowed) || isLocalhostPattern(allowed) {
				return true, fmt.Sprintf("localhost development match: %s", allowed)
			}
		}
	}
	
	return false, fmt.Sprintf("origin '%s' not in allowed origins: %v", origin, corsConfig.AllowedOrigins)
}

// getUniversalAllowedHeaders returns comprehensive headers for all client applications
func getUniversalAllowedHeaders(corsConfig *EnhancedCORSConfig) []string {
	// Check if wildcard is specified
	if len(corsConfig.AllowedHeaders) == 1 && corsConfig.AllowedHeaders[0] == "*" {
		return getAllPossibleHeaders()
	}
	
	// Merge configured headers with essential headers
	essentialHeaders := getEssentialHeaders()
	allHeaders := make(map[string]bool)
	
	// Add essential headers
	for _, header := range essentialHeaders {
		allHeaders[header] = true
	}
	
	// Add configured headers
	for _, header := range corsConfig.AllowedHeaders {
		allHeaders[header] = true
	}
	
	// Convert back to slice
	result := make([]string, 0, len(allHeaders))
	for header := range allHeaders {
		result = append(result, header)
	}
	
	return result
}

// getAllPossibleHeaders returns all headers that might be needed by client applications
func getAllPossibleHeaders() []string {
	return []string{
		// Standard HTTP headers
		"Accept", "Content-Type", "Content-Length", "Accept-Encoding",
		"Cache-Control", "X-Requested-With", "User-Agent",
		
		// Authentication headers (all variations for maximum compatibility)
		"Authorization", "Bearer",
		"Session-Token", "session-token",
		"X-Session-Token", "x-session-token", 
		"X-Auth-Token", "X-API-Key", "API-Key",
		"X-Access-Token", "Access-Token",
		
		// Project and application headers
		"X-Project-ID", "X-Project-Token", "Project-ID", "Project-Token",
		"X-Application-ID", "Application-ID",
		
		// Security headers
		"X-CSRF-Token", "X-XSRF-TOKEN", "CSRF-Token",
		
		// Framework-specific headers
		"X-Requested-With", "X-HTTP-Method-Override",
		
		// Development and debugging headers
		"X-Dev-Mode", "X-Debug", "X-Source", "X-Client-Version",
		
		// CloudBox-specific headers
		"X-CloudBox-Client", "X-CloudBox-Version", "X-CloudBox-SDK",
		
		// Additional common headers
		"X-Custom-Header", "X-Client-Info",
	}
}

// getEssentialHeaders returns headers that are always required
func getEssentialHeaders() []string {
	return []string{
		"Accept", "Content-Type", "Authorization",
		"Session-Token", "X-API-Key", "X-Session-Token",
		"X-Project-ID", "X-Requested-With",
	}
}

// ensureRequiredMethods ensures all required HTTP methods are included
func ensureRequiredMethods(configuredMethods []string) []string {
	requiredMethods := []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	methodSet := make(map[string]bool)
	
	// Add configured methods
	for _, method := range configuredMethods {
		methodSet[method] = true
	}
	
	// Add required methods if missing
	for _, method := range requiredMethods {
		methodSet[method] = true
	}
	
	// Convert back to slice
	result := make([]string, 0, len(methodSet))
	for method := range methodSet {
		result = append(result, method)
	}
	
	return result
}

// setCORSErrorHeaders sets helpful error headers for debugging
func setCORSErrorHeaders(c *gin.Context, origin string, corsConfig *EnhancedCORSConfig, reason string) {
	c.Header("X-CORS-Error", "Origin not allowed")
	c.Header("X-CORS-Error-Reason", reason)
	c.Header("X-CORS-Allowed-Origins", strings.Join(corsConfig.AllowedOrigins, ", "))
	c.Header("X-CORS-Help", fmt.Sprintf("Run: node scripts/setup-cors.js --origin=%s", origin))
	
	if corsConfig.Environment == Development {
		c.Header("X-CORS-Quick-Fix", fmt.Sprintf("Add CORS_ORIGINS=%s,http://localhost:* to .env", origin))
	}
}

// logCORSViolation logs CORS violations for monitoring
func logCORSViolation(origin, path, reason string) {
	log.Printf("CORS violation: origin=%s, path=%s, reason=%s", origin, path, reason)
}

// isLocalhostOrigin checks if an origin is a localhost-based development origin
func isLocalhostOrigin(origin string) bool {
	return strings.Contains(origin, "localhost") || 
		   strings.Contains(origin, "127.0.0.1") || 
		   strings.Contains(origin, "[::1]")
}

// isLocalhostPattern checks if an allowed origin pattern matches localhost development
func isLocalhostPattern(pattern string) bool {
	// Support patterns like:
	// "http://localhost:*" 
	// "https://localhost:*"
	// "localhost:*"
	// "*://localhost:*"
	if strings.Contains(pattern, "localhost") && strings.Contains(pattern, "*") {
		return true
	}
	
	// Support IP-based patterns
	if (strings.Contains(pattern, "127.0.0.1") || strings.Contains(pattern, "[::1]")) && 
	   strings.Contains(pattern, "*") {
		return true
	}
	
	return false
}

// matchesLocalhostPattern checks if origin matches a localhost wildcard pattern
func matchesLocalhostPattern(origin, pattern string) bool {
	if !isLocalhostPattern(pattern) {
		return false
	}
	
	// Extract the base pattern without wildcards
	if strings.Contains(pattern, "localhost") {
		// Match patterns like "http://localhost:*" with "http://localhost:4041"
		basePattern := strings.Replace(pattern, "*", "", -1)
		if strings.HasPrefix(origin, basePattern) {
			return true
		}
	}
	
	return false
}

// contains checks if a slice contains a specific string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// ProjectSmartCORS middleware handles project-specific CORS with enhanced features
func ProjectSmartCORS(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		corsConfig := getEnhancedCORSConfig(cfg)
		
		// Get project ID from context (set by ProjectAuthOrJWT middleware)
		projectID, exists := c.Get("project_id")
		if !exists {
			// Fall back to global SmartCORS if no project context
			SmartCORS(cfg)(c)
			return
		}
		
		// Load project-specific CORS configuration
		var projectCorsConfig models.CORSConfig
		err := db.Where("project_id = ?", projectID).First(&projectCorsConfig).Error
		if err != nil {
			// Fall back to global SmartCORS if no project CORS config found
			SmartCORS(cfg)(c)
			return
		}
		
		// Create enhanced config from project config
		enhancedProjectConfig := &EnhancedCORSConfig{
			Environment:    corsConfig.Environment,
			AllowedOrigins: projectCorsConfig.AllowedOrigins,
			AllowedHeaders: projectCorsConfig.AllowedHeaders,
			AllowedMethods: projectCorsConfig.AllowedMethods,
			AutoDetect:     corsConfig.AutoDetect,
			DebugMode:      corsConfig.DebugMode,
			MaxAge:         projectCorsConfig.MaxAge,
		}
		
		// For OPTIONS requests (preflight)
		if c.Request.Method == "OPTIONS" {
			if origin != "" {
				if allowed, reason := isOriginAllowedEnhanced(origin, enhancedProjectConfig); allowed {
					c.Header("Access-Control-Allow-Origin", origin)
					if enhancedProjectConfig.DebugMode {
						c.Header("X-CORS-Allow-Reason", fmt.Sprintf("project-%v: %s", projectID, reason))
					}
				} else {
					setCORSErrorHeaders(c, origin, enhancedProjectConfig, fmt.Sprintf("project-%v: %s", projectID, reason))
				}
			}
			
			c.Header("Access-Control-Allow-Methods", strings.Join(ensureRequiredMethods(enhancedProjectConfig.AllowedMethods), ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(getUniversalAllowedHeaders(enhancedProjectConfig), ", "))
			
			if projectCorsConfig.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", enhancedProjectConfig.MaxAge))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// For non-OPTIONS requests
		if origin != "" {
			if allowed, reason := isOriginAllowedEnhanced(origin, enhancedProjectConfig); allowed {
				c.Header("Access-Control-Allow-Origin", origin)
				c.Header("Access-Control-Allow-Methods", strings.Join(ensureRequiredMethods(enhancedProjectConfig.AllowedMethods), ", "))
				c.Header("Access-Control-Allow-Headers", strings.Join(getUniversalAllowedHeaders(enhancedProjectConfig), ", "))
				
				if projectCorsConfig.AllowCredentials {
					c.Header("Access-Control-Allow-Credentials", "true")
				}
				
				if enhancedProjectConfig.DebugMode {
					c.Header("X-CORS-Request-Allow-Reason", fmt.Sprintf("project-%v: %s", projectID, reason))
				}
			} else {
				logCORSViolation(origin, c.Request.URL.Path, fmt.Sprintf("project-%v: %s", projectID, reason))
				setCORSErrorHeaders(c, origin, enhancedProjectConfig, fmt.Sprintf("project-%v: %s", projectID, reason))
			}
		}

		c.Next()
	}
}