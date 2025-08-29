package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// CORS middleware handles Cross-Origin Resource Sharing
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Always set CORS headers for OPTIONS requests (preflight)
		if c.Request.Method == "OPTIONS" {
			// For preflight, check if origin is allowed, but be more permissive
			if origin != "" && (isOriginAllowed(origin, cfg.AllowedOrigins) || contains(cfg.AllowedOrigins, "*")) {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if origin != "" {
				// For development, allow localhost origins even if not explicitly configured
				if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
			
			// Ensure PATCH is always included
			allowedMethods := cfg.AllowedMethods
			hasPatch := false
			for _, method := range allowedMethods {
				if method == "PATCH" {
					hasPatch = true
					break
				}
			}
			if !hasPatch {
				allowedMethods = append(allowedMethods, "PATCH")
			}
			
			c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			
			// Handle AllowedHeaders - if wildcard, allow comprehensive headers including Authorization and session tokens
			if len(cfg.AllowedHeaders) == 1 && cfg.AllowedHeaders[0] == "*" {
				allowedHeaders := getDefaultAllowedHeaders(cfg)
				c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			} else {
				c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}
			
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// For non-OPTIONS requests, check if origin is allowed
		if isOriginAllowed(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			
			// Ensure PATCH is always included
			allowedMethods := cfg.AllowedMethods
			hasPatch := false
			for _, method := range allowedMethods {
				if method == "PATCH" {
					hasPatch = true
					break
				}
			}
			if !hasPatch {
				allowedMethods = append(allowedMethods, "PATCH")
			}
			c.Header("Access-Control-Allow-Methods", strings.Join(allowedMethods, ", "))
			
			// Handle AllowedHeaders - if wildcard, allow comprehensive headers including Authorization and session tokens
			if len(cfg.AllowedHeaders) == 1 && cfg.AllowedHeaders[0] == "*" {
				allowedHeaders := getDefaultAllowedHeaders(cfg)
				c.Header("Access-Control-Allow-Headers", strings.Join(allowedHeaders, ", "))
			} else {
				c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
			}
			
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Max-Age", "86400")
		}

		c.Next()
	}
}

// isOriginAllowed checks if the origin is in the allowed origins list
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	for _, allowed := range allowedOrigins {
		if allowed == "*" || allowed == origin {
			return true
		}
		
		// Handle wildcard subdomains (e.g., *.example.com)
		if strings.HasPrefix(allowed, "*.") {
			domain := strings.TrimPrefix(allowed, "*.")
			if strings.HasSuffix(origin, domain) {
				return true
			}
		}
		
		// Handle localhost wildcard patterns (e.g., http://localhost:*)
		if matchesLocalhostPattern(origin, allowed) {
			return true
		}
	}
	
	// Fall back to localhost detection for development environments
	// This provides extra flexibility for development setups
	if isLocalhostOrigin(origin) {
		// Check if any allowed origins suggest localhost development is allowed
		for _, allowed := range allowedOrigins {
			if isLocalhostOrigin(allowed) || isLocalhostPattern(allowed) {
				return true
			}
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

// getDefaultAllowedHeaders returns comprehensive headers for wildcard configuration
// This can be customized via environment variables or extended for specific use cases
func getDefaultAllowedHeaders(cfg *config.Config) []string {
	baseHeaders := []string{
		"Accept",
		"Content-Type", 
		"Content-Length",
		"Accept-Encoding",
		"Authorization",
		"X-CSRF-Token",
		"X-API-Key",
		"Cache-Control",
		"X-Requested-With",
	}
	
	// Add session token headers (both variations for compatibility)
	sessionHeaders := []string{
		"Session-Token",
		"session-token", 
		"X-Session-Token",
		"x-session-token",
	}
	
	// Add project-specific headers if needed
	projectHeaders := []string{
		"X-Project-ID",
		"X-Project-Token",
		"Project-ID",
		"Project-Token",
	}
	
	// Combine all headers
	allHeaders := append(baseHeaders, sessionHeaders...)
	allHeaders = append(allHeaders, projectHeaders...)
	
	return allHeaders
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

// ProjectCORS middleware handles project-specific CORS configuration
func ProjectCORS(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// Get project ID from context (set by ProjectAuthOrJWT middleware)
		projectID, exists := c.Get("project_id")
		if !exists {
			// Fall back to global CORS if no project context
			CORS(cfg)(c)
			return
		}
		
		// Load project-specific CORS configuration
		var corsConfig models.CORSConfig
		err := db.Where("project_id = ?", projectID).First(&corsConfig).Error
		if err != nil {
			// Fall back to global CORS if no project CORS config found
			CORS(cfg)(c)
			return
		}
		
		// For OPTIONS requests (preflight)
		if c.Request.Method == "OPTIONS" {
			if origin != "" && (isOriginAllowed(origin, corsConfig.AllowedOrigins) || contains(corsConfig.AllowedOrigins, "*")) {
				c.Header("Access-Control-Allow-Origin", origin)
			} else if origin != "" {
				// For development, allow localhost origins even if not explicitly configured
				if strings.Contains(origin, "localhost") || strings.Contains(origin, "127.0.0.1") {
					c.Header("Access-Control-Allow-Origin", origin)
				}
			}
			
			c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
			if corsConfig.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
			c.Header("Access-Control-Max-Age", fmt.Sprintf("%d", corsConfig.MaxAge))
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		
		// For non-OPTIONS requests
		if isOriginAllowed(origin, corsConfig.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", strings.Join(corsConfig.AllowedMethods, ", "))
			c.Header("Access-Control-Allow-Headers", strings.Join(corsConfig.AllowedHeaders, ", "))
			if corsConfig.AllowCredentials {
				c.Header("Access-Control-Allow-Credentials", "true")
			}
		}

		c.Next()
	}
}