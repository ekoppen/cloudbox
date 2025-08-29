package middleware

import (
	"strings"
	"strconv"
	"time"
	"net"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/cloudbox/backend/internal/models"
)

// APILoggingConfig holds configuration for API logging middleware
type APILoggingConfig struct {
	DB           *gorm.DB
	SkipPaths    []string  // Paths to skip logging (e.g., health checks)
	LogOnlyAPI   bool      // Only log API calls (starting with /p/ or /api/)
}

// APILoggingMiddleware creates a middleware that logs API requests for statistics
func APILoggingMiddleware(config APILoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		path := c.Request.URL.Path
		
		// Debug: log every request the middleware sees
		println("API Logging Middleware - Path:", path)
		
		// Skip if this path should not be logged
		if shouldSkipLogging(path, config.SkipPaths) {
			println("API Logging - Skipping path:", path)
			c.Next()
			return
		}
		
		// Skip if LogOnlyAPI is true and this is not an API call
		if config.LogOnlyAPI && !isAPIPath(c.Request.URL.Path) {
			c.Next()
			return
		}
		
		// Create a custom response writer to capture response size
		crw := &customResponseWriter{ResponseWriter: c.Writer, size: 0}
		c.Writer = crw
		
		// Process request
		c.Next()
		
		// Calculate response time
		duration := time.Since(startTime)
		responseTimeMs := int(duration.Nanoseconds() / 1000000) // Convert to milliseconds
		
		// Extract project ID from path if available
		var projectID *uint
		if pid := extractProjectIDFromPath(c.Request.URL.Path); pid != 0 {
			projectID = &pid
		}
		
		// Skip logging if no project ID and LogOnlyAPI is true (system calls)
		if config.LogOnlyAPI && projectID == nil {
			return
		}
		
		// Get authentication info
		var apiKeyID *uint
		var userID *uint
		
		if keyID, exists := c.Get("api_key_id"); exists {
			if id, ok := keyID.(uint); ok {
				apiKeyID = &id
			}
		}
		
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}
		
		// Extract IP address
		ipAddress := getClientIP(c)
		
		// Normalize endpoint for grouping (remove IDs and query params)
		normalizedEndpoint := normalizeEndpoint(c.Request.URL.Path)
		
		// Create log entry
		logEntry := models.APIRequestLog{
			ProjectID:        projectID,
			Method:           c.Request.Method,
			Endpoint:         normalizedEndpoint,
			FullPath:         c.Request.URL.Path,
			UserAgent:        c.Request.UserAgent(),
			IPAddress:        ipAddress,
			StatusCode:       c.Writer.Status(),
			ResponseTimeMs:   responseTimeMs,
			ResponseSizeBytes: crw.size,
			APIKeyID:         apiKeyID,
			UserID:           userID,
		}
		
		// Save to database (sync for debugging)
		if err := config.DB.Create(&logEntry).Error; err != nil {
			println("Failed to save API request log:", err.Error())
		} else {
			println("Successfully saved API request log for:", logEntry.Method, logEntry.Endpoint)
		}
	}
}

// customResponseWriter wraps gin.ResponseWriter to capture response size
type customResponseWriter struct {
	gin.ResponseWriter
	size int
}

func (w *customResponseWriter) Write(b []byte) (int, error) {
	size, err := w.ResponseWriter.Write(b)
	w.size += size
	return size, err
}

// shouldSkipLogging checks if a path should be skipped from logging
func shouldSkipLogging(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}

// isAPIPath checks if the path is an API call
func isAPIPath(path string) bool {
	return strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/p/")
}

// extractProjectIDFromPath extracts project ID from paths like /p/{id}/... or /api/v1/projects/{id}/...
func extractProjectIDFromPath(path string) uint {
	// Handle /p/{id}/... format
	if strings.HasPrefix(path, "/p/") {
		parts := strings.Split(path, "/")
		if len(parts) >= 3 {
			if id, err := strconv.ParseUint(parts[2], 10, 32); err == nil {
				return uint(id)
			}
		}
	}
	
	// Handle /api/v1/projects/{id}/... format
	if strings.Contains(path, "/projects/") {
		parts := strings.Split(path, "/")
		for i, part := range parts {
			if part == "projects" && i+1 < len(parts) {
				if id, err := strconv.ParseUint(parts[i+1], 10, 32); err == nil {
					return uint(id)
				}
			}
		}
	}
	
	return 0
}

// getClientIP extracts the real client IP from the request
func getClientIP(c *gin.Context) string {
	// Check for X-Forwarded-For header
	if forwarded := c.GetHeader("X-Forwarded-For"); forwarded != "" {
		ips := strings.Split(forwarded, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}
	
	// Check for X-Real-IP header
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}
	
	// Fall back to remote address
	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}
	
	return c.Request.RemoteAddr
}

// normalizeEndpoint normalizes API endpoints for grouping statistics
// Replaces IDs and UUIDs with placeholders for better grouping
func normalizeEndpoint(path string) string {
	parts := strings.Split(path, "/")
	
	for i, part := range parts {
		// Replace numeric IDs with {id}
		if isNumeric(part) {
			parts[i] = "{id}"
		}
		// Replace UUIDs with {uuid}
		if isUUID(part) {
			parts[i] = "{uuid}"
		}
		// Replace common ID patterns
		if strings.HasSuffix(part, ".jpg") || strings.HasSuffix(part, ".png") || strings.HasSuffix(part, ".pdf") {
			parts[i] = "{file}"
		}
	}
	
	normalized := strings.Join(parts, "/")
	
	// Remove query parameters
	if idx := strings.Index(normalized, "?"); idx != -1 {
		normalized = normalized[:idx]
	}
	
	return normalized
}

// isNumeric checks if a string is numeric
func isNumeric(s string) bool {
	_, err := strconv.ParseUint(s, 10, 64)
	return err == nil
}

// isUUID checks if a string looks like a UUID
func isUUID(s string) bool {
	return len(s) == 36 && 
		   s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-'
}