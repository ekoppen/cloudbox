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
		
		// Check if origin is allowed
		if isOriginAllowed(origin, cfg.AllowedOrigins) {
			c.Header("Access-Control-Allow-Origin", origin)
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

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
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