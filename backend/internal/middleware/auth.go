package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Claims represents JWT claims
type Claims struct {
	UserID uint   `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// RequireAuth middleware validates JWT tokens
func RequireAuth(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header required"})
			c.Abort()
			return
		}

		// Extract token from "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			c.Abort()
			return
		}
		tokenString := parts[1]

		// Parse and validate token
		token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
			return []byte(cfg.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(*Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}

		// Store user info in context
		c.Set("user_id", claims.UserID)
		c.Set("user_email", claims.Email)
		c.Set("user_role", claims.Role)

		c.Next()
	}
}

// ProjectAuth middleware validates project access via API key
func ProjectAuth(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectSlug := c.Param("project_slug")
		if projectSlug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug required"})
			c.Abort()
			return
		}

		// Get API key from header
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key required"})
			c.Abort()
			return
		}

		// Find project and validate API key
		var project models.Project
		var key models.APIKey

		// Try to parse as ID first (numeric), then by slug
		var err error
		if projectID, parseErr := strconv.ParseUint(projectSlug, 10, 32); parseErr == nil {
			err = db.Where("id = ?", uint(projectID)).First(&project).Error
		}
		if err != nil {
			err = db.Where("slug = ?", projectSlug).First(&project).Error
		}
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			c.Abort()
			return
		}

		// Find API key by project - we need to check hash since keys are not stored in plain text
		var apiKeys []models.APIKey
		err = db.Where("project_id = ? AND is_active = true", project.ID).Find(&apiKeys).Error
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
			c.Abort()
			return
		}

		// Check each API key hash
		var validKey *models.APIKey
		for _, k := range apiKeys {
			if err := bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(apiKey)); err == nil {
				validKey = &k
				break
			}
		}

		if validKey == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid API key"})
			c.Abort()
			return
		}

		key = *validKey

		// Check if key is expired
		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
			c.Abort()
			return
		}

		// Update last used timestamp in transaction for consistency
		tx := db.Begin()
		if err := tx.Model(&key).Update("last_used_at", time.Now()).Error; err != nil {
			log.Printf("Failed to update API key last_used_at: %v", err)
			tx.Rollback()
		} else {
			tx.Commit()
		}

		// Store project and key info in context
		c.Set("project", project)
		c.Set("api_key", key)

		c.Next()
	}
}

// GenerateJWT generates a new JWT token for a user
func GenerateJWT(userID uint, email string, role string, cfg *config.Config) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    "cloudbox",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(cfg.JWTSecret))
}

// RequireSuperAdmin middleware checks if user has superadmin role
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Super admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdminOrSuperAdmin middleware checks if user has admin or superadmin role
func RequireAdminOrSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		userRole := c.GetString("user_role")
		if userRole != "admin" && userRole != "superadmin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ProjectOnly middleware validates project exists without requiring authentication
func ProjectOnly(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectSlug := c.Param("project_slug")
		if projectSlug == "" {
			log.Printf("ProjectOnly: No project_slug in URL path")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug required"})
			c.Abort()
			return
		}

		log.Printf("ProjectOnly: Looking for project with slug: %s", projectSlug)

		// Test database connection
		var testCount int64
		if err := db.Model(&models.Project{}).Count(&testCount).Error; err != nil {
			log.Printf("ProjectOnly: Database connection error: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Database connection error"})
			c.Abort()
			return
		}
		log.Printf("ProjectOnly: Database has %d projects total", testCount)

		// Find the project - try by ID first, then by slug
		var project models.Project
		var err error
		
		// Try to parse as ID first (numeric) - photoportfolio is not numeric so skip this
		if projectID, parseErr := strconv.ParseUint(projectSlug, 10, 32); parseErr == nil {
			log.Printf("ProjectOnly: Trying to find project by ID: %d", uint(projectID))
			err = db.Where("id = ?", uint(projectID)).First(&project).Error
			if err != nil {
				log.Printf("ProjectOnly: Error finding by ID: %v", err)
			} else {
				log.Printf("ProjectOnly: Found by ID - ID: %d, Name: %s, Slug: %s", project.ID, project.Name, project.Slug)
			}
		} else {
			log.Printf("ProjectOnly: '%s' is not numeric, will search by slug", projectSlug)
		}
		
		// If not found by ID (or slug is not numeric), try by slug
		if err != nil || project.ID == 0 {
			log.Printf("ProjectOnly: Trying to find project by slug: %s", projectSlug)
			err = db.Where("slug = ?", projectSlug).First(&project).Error
			if err != nil {
				log.Printf("ProjectOnly: Error finding by slug: %v", err)
			} else {
				log.Printf("ProjectOnly: Found by slug - ID: %d, Name: %s, Slug: %s", project.ID, project.Name, project.Slug)
			}
		}
		
		if err != nil {
			log.Printf("ProjectOnly: Project not found for slug '%s', error: %v", projectSlug, err)
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			c.Abort()
			return
		}

		log.Printf("ProjectOnly: Found project ID: %d, Name: %s, Slug: %s", project.ID, project.Name, project.Slug)
		
		// Store project info in context
		c.Set("project", project)
		c.Set("project_id", project.ID)
		c.Next()
	}
}

// ProjectAuthOrJWT middleware accepts both JWT tokens and API keys for project access
func ProjectAuthOrJWT(cfg *config.Config, db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectSlug := c.Param("project_slug")
		if projectSlug == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Project slug required"})
			c.Abort()
			return
		}

		// Find the project first - try by ID first, then by slug
		var project models.Project
		var err error
		
		// Try to parse as ID first (numeric)
		if projectID, parseErr := strconv.ParseUint(projectSlug, 10, 32); parseErr == nil {
			err = db.Where("id = ?", uint(projectID)).First(&project).Error
		} else {
			err = fmt.Errorf("not numeric") // Ensure err is set so we try slug lookup
		}
		
		// If not found by ID, try by slug
		if err != nil {
			err = db.Where("slug = ?", projectSlug).First(&project).Error
		}
		
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Project not found"})
			c.Abort()
			return
		}

		// Try API key authentication first
		apiKey := c.GetHeader("X-API-Key")
		if apiKey != "" {
			// Find API key by project - we need to check hash since keys are not stored in plain text
			var apiKeys []models.APIKey
			err = db.Where("project_id = ? AND is_active = true", project.ID).Find(&apiKeys).Error
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
				c.Abort()
				return
			}

			// Check each API key hash
			var validKey *models.APIKey
			for _, k := range apiKeys {
				if err := bcrypt.CompareHashAndPassword([]byte(k.KeyHash), []byte(apiKey)); err == nil {
					validKey = &k
					break
				}
			}

			if validKey != nil {
				key := *validKey
				// Check if key is expired
				if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
					c.JSON(http.StatusUnauthorized, gin.H{"error": "API key expired"})
					c.Abort()
					return
				}

				// Update last used timestamp
				tx := db.Begin()
				if err := tx.Model(&key).Update("last_used_at", time.Now()).Error; err != nil {
					log.Printf("Failed to update API key last_used_at: %v", err)
					tx.Rollback()
				} else {
					tx.Commit()
				}

				// Store project and key info in context
				c.Set("project", project)
				c.Set("project_id", project.ID)
				c.Set("api_key", key)
				c.Next()
				return
			}
		}

		// If no valid API key, try JWT authentication
		authHeader := c.GetHeader("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && parts[0] == "Bearer" {
				tokenString := parts[1]

				// Parse and validate token
				token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
					return []byte(cfg.JWTSecret), nil
				})

				if err == nil && token.Valid {
					// Extract claims
					claims, ok := token.Claims.(*Claims)
					if ok {
						// For JWT access, check if user has admin access to the project
						userRole := claims.Role
						if userRole == "admin" || userRole == "superadmin" {
							// Store user and project info in context
							c.Set("user_id", claims.UserID)
							c.Set("user_email", claims.Email)
							c.Set("user_role", claims.Role)
							c.Set("project", project)
							c.Set("project_id", project.ID)
							c.Next()
							return
						}
					}
				}
			}
		}

		// Neither authentication method worked
		c.JSON(http.StatusUnauthorized, gin.H{"error": "API key or valid authorization required"})
		c.Abort()
	}
}