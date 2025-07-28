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
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserHandler handles app user management requests
type UserHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewUserHandler creates a new user handler
func NewUserHandler(db *gorm.DB, cfg *config.Config) *UserHandler {
	return &UserHandler{db: db, cfg: cfg}
}

// User Management

// ListUsers returns all users for a project
func (h *UserHandler) ListUsers(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	// Parse query parameters
	limit := 25 // Default limit
	offset := 0
	orderBy := "created_at DESC"
	
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}
	
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	
	if order := c.Query("orderBy"); order != "" {
		if order == "email" || order == "name" || order == "created_at" || order == "last_login_at" {
			orderBy = order
			if dir := c.Query("orderDir"); dir == "desc" {
				orderBy += " DESC"
			} else {
				orderBy += " ASC"
			}
		}
	}
	
	var users []models.AppUser
	query := h.db.Where("project_id = ?", project.ID).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	// Optional filters
	if email := c.Query("email"); email != "" {
		query = query.Where("email ILIKE ?", "%"+email+"%")
	}
	
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	
	if active := c.Query("active"); active != "" {
		if active == "true" {
			query = query.Where("is_active = ?", true)
		} else if active == "false" {
			query = query.Where("is_active = ?", false)
		}
	}
	
	if err := query.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}
	
	// Get total count
	var total int64
	countQuery := h.db.Model(&models.AppUser{}).Where("project_id = ?", project.ID)
	
	if email := c.Query("email"); email != "" {
		countQuery = countQuery.Where("email ILIKE ?", "%"+email+"%")
	}
	if name := c.Query("name"); name != "" {
		countQuery = countQuery.Where("name ILIKE ?", "%"+name+"%")
	}
	if active := c.Query("active"); active != "" {
		if active == "true" {
			countQuery = countQuery.Where("is_active = ?", true)
		} else if active == "false" {
			countQuery = countQuery.Where("is_active = ?", false)
		}
	}
	
	countQuery.Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"users":  users,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// CreateUser creates a new app user
func (h *UserHandler) CreateUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var req struct {
		Email       string                 `json:"email" binding:"required,email"`
		Password    string                 `json:"password" binding:"required,min=8"`
		Name        string                 `json:"name"`
		Username    string                 `json:"username"`
		ProfileData map[string]interface{} `json:"profile_data"`
		IsActive    *bool                  `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Check if user already exists in this project
	var existingUser models.AppUser
	if err := h.db.Where("project_id = ? AND email = ?", project.ID, req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	
	// Check username uniqueness if provided
	if req.Username != "" {
		if err := h.db.Where("project_id = ? AND username = ?", project.ID, req.Username).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "User with this username already exists"})
			return
		}
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// Set defaults
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	
	// Create user
	user := models.AppUser{
		ID:           uuid.New().String(),
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Username:     req.Username,
		ProfileData:  req.ProfileData,
		IsActive:     isActive,
		ProjectID:    project.ID,
	}
	
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	
	c.JSON(http.StatusCreated, user)
}

// GetUser returns a specific user
func (h *UserHandler) GetUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	c.JSON(http.StatusOK, user)
}

// UpdateUser updates a user
func (h *UserHandler) UpdateUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	
	var req struct {
		Name        *string                `json:"name"`
		Username    *string                `json:"username"`
		ProfileData map[string]interface{} `json:"profile_data"`
		Preferences map[string]interface{} `json:"preferences"`
		IsActive    *bool                  `json:"is_active"`
		Status      *string                `json:"status"` // Support frontend 'status' field
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Debug logging
	fmt.Printf("UpdateUser: userID=%s, IsActive=%v, Status=%v\n", userID, req.IsActive, req.Status)
	if req.IsActive != nil {
		fmt.Printf("UpdateUser: IsActive value=%t\n", *req.IsActive)
	}
	if req.Status != nil {
		fmt.Printf("UpdateUser: Status value=%s\n", *req.Status)
	}
	
	// Find user
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Check username uniqueness if changed
	if req.Username != nil && *req.Username != user.Username {
		var existingUser models.AppUser
		if err := h.db.Where("project_id = ? AND username = ? AND id != ?", project.ID, *req.Username, userID).First(&existingUser).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Username already taken"})
			return
		}
	}
	
	// Update fields - using Select to ensure boolean false values are updated
	updates := make(map[string]interface{})
	selectFields := []string{}
	
	if req.Name != nil {
		updates["name"] = *req.Name
		selectFields = append(selectFields, "name")
	}
	if req.Username != nil {
		updates["username"] = *req.Username
		selectFields = append(selectFields, "username")
	}
	if req.ProfileData != nil {
		updates["profile_data"] = req.ProfileData
		selectFields = append(selectFields, "profile_data")
	}
	if req.Preferences != nil {
		updates["preferences"] = req.Preferences
		selectFields = append(selectFields, "preferences")
	}
	// Handle both is_active boolean and status string
	var isActiveValue *bool
	if req.IsActive != nil {
		isActiveValue = req.IsActive
	} else if req.Status != nil {
		// Convert status string to is_active boolean
		active := (*req.Status == "active")
		isActiveValue = &active
	}
	
	if isActiveValue != nil {
		updates["is_active"] = *isActiveValue
		selectFields = append(selectFields, "is_active")
	}
	
	// Use Select to force update of specified fields, including boolean false values
	if err := h.db.Model(&user).Select(selectFields).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	// Return updated user
	h.db.Where("project_id = ? AND id = ?", project.ID, userID).First(&user)
	c.JSON(http.StatusOK, user)
}

// DeleteUser deletes a user
func (h *UserHandler) DeleteUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	
	result := h.db.Where("project_id = ? AND id = ?", project.ID, userID).Delete(&models.AppUser{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Also delete user sessions
	h.db.Where("project_id = ? AND user_id = ?", project.ID, userID).Delete(&models.AppSession{})
	
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// Authentication

// LoginUser authenticates a user and creates a session
func (h *UserHandler) LoginUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("LoginUser: JSON binding error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	fmt.Printf("LoginUser: attempting login for email=%s, project=%s\n", req.Email, project.Slug)
	
	// Find user
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND email = ?", project.ID, req.Email).First(&user).Error; err != nil {
		fmt.Printf("LoginUser: user not found for email=%s, project_id=%d, error=%v\n", req.Email, project.ID, err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	fmt.Printf("LoginUser: found user id=%s, is_active=%t\n", user.ID, user.IsActive)
	
	// Check if user is active
	if !user.IsActive {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is disabled"})
		return
	}
	
	// Check if account is locked
	if user.LockedUntil != nil && user.LockedUntil.After(time.Now()) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Account is temporarily locked"})
		return
	}
	
	// Verify password
	fmt.Printf("LoginUser: Comparing password for user %s\n", user.Email)
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		fmt.Printf("LoginUser: Password verification failed for user %s: %v\n", user.Email, err)
		// Increment login attempts
		user.LoginAttempts++
		if user.LoginAttempts >= 5 {
			lockUntil := time.Now().Add(15 * time.Minute)
			user.LockedUntil = &lockUntil
		}
		h.db.Save(&user)
		
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}
	
	fmt.Printf("LoginUser: Password verification successful for user %s\n", user.Email)
	
	// Reset login attempts on successful login
	user.LoginAttempts = 0
	user.LockedUntil = nil
	now := time.Now()
	user.LastLoginAt = &now
	user.LastSeenAt = &now
	h.db.Save(&user)
	
	// Create session
	sessionToken, err := generateSecureToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate session token"})
		return
	}
	
	session := models.AppSession{
		ID:        uuid.New().String(),
		UserID:    user.ID,
		Token:     sessionToken,
		ExpiresAt: time.Now().Add(24 * time.Hour),
		IPAddress: c.ClientIP(),
		UserAgent: c.GetHeader("User-Agent"),
		ProjectID: project.ID,
		IsActive:  true,
	}
	
	if err := h.db.Create(&session).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"user": user,
		"session": gin.H{
			"token":      sessionToken,
			"expires_at": session.ExpiresAt,
		},
	})
}

// LogoutUser invalidates a user session
func (h *UserHandler) LogoutUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	sessionToken := c.GetHeader("Session-Token")
	
	if sessionToken == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Session token required"})
		return
	}
	
	result := h.db.Where("project_id = ? AND token = ?", project.ID, sessionToken).Delete(&models.AppSession{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to logout"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetCurrentUser returns the current authenticated user
func (h *UserHandler) GetCurrentUser(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	sessionToken := c.GetHeader("Session-Token")
	
	if sessionToken == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Session token required"})
		return
	}
	
	// Find valid session
	var session models.AppSession
	if err := h.db.Where("project_id = ? AND token = ? AND expires_at > ? AND is_active = ?", 
		project.ID, sessionToken, time.Now(), true).First(&session).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired session"})
		return
	}
	
	// Get user
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, session.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	
	// Update last activity
	now := time.Now()
	session.LastActivity = &now
	user.LastSeenAt = &now
	h.db.Save(&session)
	h.db.Save(&user)
	
	c.JSON(http.StatusOK, user)
}

// ChangePassword changes a user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=8"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find user
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, userID).First(&user).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Current password is incorrect"})
		return
	}
	
	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// Update password
	user.PasswordHash = string(hashedPassword)
	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	
	// Invalidate all sessions for this user
	h.db.Where("project_id = ? AND user_id = ?", project.ID, userID).Delete(&models.AppSession{})
	
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// Session Management

// ListSessions returns all sessions for a user
func (h *UserHandler) ListSessions(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	
	var sessions []models.AppSession
	if err := h.db.Where("project_id = ? AND user_id = ? AND is_active = ?", 
		project.ID, userID, true).Find(&sessions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sessions"})
		return
	}
	
	c.JSON(http.StatusOK, sessions)
}

// RevokeSession revokes a specific session
func (h *UserHandler) RevokeSession(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	userID := c.Param("user_id")
	sessionID := c.Param("session_id")
	
	result := h.db.Where("project_id = ? AND user_id = ? AND id = ?", 
		project.ID, userID, sessionID).Delete(&models.AppSession{})
	
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke session"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Session not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Session revoked successfully"})
}

// GetAuthSettings returns authentication settings for the project
func (h *UserHandler) GetAuthSettings(c *gin.Context) {
	_ = c.MustGet("project").(models.Project)
	
	// Return default auth settings - in a real implementation, 
	// these would be stored in the database per project
	settings := gin.H{
		"email_verification":   true,
		"password_min_length":  8,
		"session_duration":     24,
		"max_login_attempts":   5,
		"lockout_duration":     30,
		"providers": []gin.H{
			{"id": "email", "name": "Email/Password", "enabled": true, "icon": "✉️"},
			{"id": "google", "name": "Google OAuth", "enabled": false, "icon": "🌐"},
			{"id": "github", "name": "GitHub OAuth", "enabled": false, "icon": "⚫"},
			{"id": "apple", "name": "Apple ID", "enabled": false, "icon": "🍎"},
		},
	}
	
	c.JSON(http.StatusOK, settings)
}

// UpdateAuthSettings updates authentication settings for the project
func (h *UserHandler) UpdateAuthSettings(c *gin.Context) {
	_ = c.MustGet("project").(models.Project)
	
	var req struct {
		EmailVerification  bool `json:"email_verification"`
		PasswordMinLength  int  `json:"password_min_length"`
		SessionDuration    int  `json:"session_duration"`
		MaxLoginAttempts   int  `json:"max_login_attempts"`
		LockoutDuration    int  `json:"lockout_duration"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// In a real implementation, you would save these settings to the database
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{"message": "Settings updated successfully"})
}

// GetAuthProviders returns available authentication providers
func (h *UserHandler) GetAuthProviders(c *gin.Context) {
	_ = c.MustGet("project").(models.Project)
	
	providers := []gin.H{
		{"id": "email", "name": "Email/Password", "enabled": true, "icon": "✉️"},
		{"id": "google", "name": "Google OAuth", "enabled": false, "icon": "🌐"},
		{"id": "github", "name": "GitHub OAuth", "enabled": false, "icon": "⚫"},
		{"id": "apple", "name": "Apple ID", "enabled": false, "icon": "🍎"},
	}
	
	c.JSON(http.StatusOK, providers)
}

// UpdateAuthProvider updates an authentication provider's settings
func (h *UserHandler) UpdateAuthProvider(c *gin.Context) {
	_ = c.MustGet("project").(models.Project)
	providerID := c.Param("provider_id")
	
	var req struct {
		Enabled bool `json:"enabled"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// In a real implementation, you would save provider settings to the database
	// For now, we'll just return success
	c.JSON(http.StatusOK, gin.H{
		"message": "Provider settings updated successfully",
		"provider_id": providerID,
		"enabled": req.Enabled,
	})
}

// Helper functions

// generateSecureToken generates a cryptographically secure token
func generateSecureToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}