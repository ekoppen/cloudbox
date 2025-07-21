package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/gin-gonic/gin"
)

// RateLimit middleware implements basic rate limiting
func RateLimit(cfg *config.Config) gin.HandlerFunc {
	// Simple in-memory store for demo purposes
	// In production, use Redis or another distributed store
	clients := make(map[string][]time.Time)

	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		// Get current time
		now := time.Now()

		// Clean old entries (older than window)
		window := 1 * time.Hour // Default window
		if cfg.RateLimitWindow != "" {
			if d, err := time.ParseDuration(cfg.RateLimitWindow); err == nil {
				window = d
			}
		}

		// Get or create client request history
		requests, exists := clients[clientIP]
		if !exists {
			requests = []time.Time{}
		}

		// Remove old requests outside the window
		var validRequests []time.Time
		cutoff := now.Add(-window)
		for _, reqTime := range requests {
			if reqTime.After(cutoff) {
				validRequests = append(validRequests, reqTime)
			}
		}

		// Check if client has exceeded rate limit
		if len(validRequests) >= cfg.RateLimitRequests {
			c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitRequests))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(window).Unix(), 10))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Rate limit exceeded",
				"retry_after": window.Seconds(),
			})
			c.Abort()
			return
		}

		// Add current request
		validRequests = append(validRequests, now)
		clients[clientIP] = validRequests

		// Set rate limit headers
		remaining := cfg.RateLimitRequests - len(validRequests)
		c.Header("X-RateLimit-Limit", strconv.Itoa(cfg.RateLimitRequests))
		c.Header("X-RateLimit-Remaining", strconv.Itoa(remaining))
		c.Header("X-RateLimit-Reset", strconv.FormatInt(now.Add(window).Unix(), 10))

		c.Next()
	}
}