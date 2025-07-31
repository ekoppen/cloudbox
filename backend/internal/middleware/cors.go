package middleware

import (
	"net/http"
	"strings"

	"github.com/cloudbox/backend/internal/config"
	"github.com/gin-gonic/gin"
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
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
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
			c.Header("Access-Control-Allow-Headers", strings.Join(cfg.AllowedHeaders, ", "))
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