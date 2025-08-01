package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/middleware"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(db *gorm.DB, cfg *config.Config) *AuthHandler {
	return &AuthHandler{db: db, cfg: cfg}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Name     string `json:"name" binding:"required"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	RememberMe bool   `json:"remember_me"`
}

// Register handles user registration
func (h *AuthHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	// Create user with default admin role
	user := models.User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		Name:         req.Name,
		Role:         models.RoleAdmin, // Default role is admin (project admin)
		IsActive:     true,
	}

	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, string(user.Role), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user":  user,
		"token": token,
	})
}

// Login handles user login
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find user
	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	h.db.Save(&user)

	// Generate JWT token
	token, err := middleware.GenerateJWT(user.ID, user.Email, string(user.Role), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate token"})
		return
	}

	response := gin.H{
		"user":  user,
		"token": token,
	}

	// Generate refresh token if remember me is enabled
	if req.RememberMe {
		refreshToken, err := h.generateRefreshToken(user.ID, c.ClientIP(), c.GetHeader("User-Agent"))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
			return
		}
		response["refresh_token"] = refreshToken
	}

	c.JSON(http.StatusOK, response)
}

// generateRefreshToken creates a new refresh token for the user
func (h *AuthHandler) generateRefreshToken(userID uint, ipAddress, userAgent string) (string, error) {
	// Generate random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", err
	}
	
	token := hex.EncodeToString(tokenBytes)
	
	// Hash the token for storage
	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])
	
	// Create refresh token record
	refreshToken := models.RefreshToken{
		Token:     token,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		IsActive:  true,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		UserID:    userID,
	}
	
	// Save to database
	if err := h.db.Create(&refreshToken).Error; err != nil {
		return "", err
	}
	
	return token, nil
}

// RefreshToken handles token refresh
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Hash the provided token to compare with stored hash
	hash := sha256.Sum256([]byte(req.RefreshToken))
	tokenHash := hex.EncodeToString(hash[:])
	
	// Find refresh token
	var refreshToken models.RefreshToken
	if err := h.db.Where("token_hash = ? AND is_active = true", tokenHash).Preload("User").First(&refreshToken).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid refresh token"})
		return
	}
	
	// Check if token is expired
	if time.Now().After(refreshToken.ExpiresAt) {
		// Deactivate expired token
		h.db.Model(&refreshToken).Update("is_active", false)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Refresh token expired"})
		return
	}
	
	// Generate new JWT token
	newToken, err := middleware.GenerateJWT(refreshToken.User.ID, refreshToken.User.Email, string(refreshToken.User.Role), h.cfg)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate new token"})
		return
	}
	
	// Update user's last login time
	now := time.Now()
	refreshToken.User.LastLoginAt = &now
	h.db.Save(&refreshToken.User)
	
	c.JSON(http.StatusOK, gin.H{
		"user":  refreshToken.User,
		"token": newToken,
	})
}

// Logout handles user logout
func (h *AuthHandler) Logout(c *gin.Context) {
	// Get optional refresh token from request
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	c.ShouldBindJSON(&req)
	
	// If refresh token provided, invalidate it
	if req.RefreshToken != "" {
		hash := sha256.Sum256([]byte(req.RefreshToken))
		tokenHash := hex.EncodeToString(hash[:])
		
		// Deactivate the refresh token
		h.db.Model(&models.RefreshToken{}).Where("token_hash = ?", tokenHash).Update("is_active", false)
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Logged out successfully"})
}

// GetProfile returns the current user's profile
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// UpdateProfile updates the current user's profile
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var req struct {
		Name string `json:"name"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update user
	if err := h.db.Model(&models.User{}).Where("id = ?", userID).Update("name", req.Name).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Profile updated successfully"})
}

// ChangePassword allows a user to change their own password
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID := c.GetUint("user_id")
	
	var req struct {
		CurrentPassword string `json:"current_password" binding:"required"`
		NewPassword     string `json:"new_password" binding:"required,min=6"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Get current user
	var user models.User
	if err := h.db.First(&user, userID).Error; err != nil {
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash new password"})
		return
	}
	
	// Update password
	if err := h.db.Model(&user).Update("password_hash", string(hashedPassword)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Password changed successfully"})
}

// ListUsers returns all users (superadmin only)
func (h *AuthHandler) ListUsers(c *gin.Context) {
	var users []models.User
	// Try with preload first, fallback to simple query if it fails
	if err := h.db.Preload("OrganizationAdmins.Organization").Find(&users).Error; err != nil {
		// Fallback: load users without preloading relationships
		if err := h.db.Find(&users).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
			return
		}
	}
	
	c.JSON(http.StatusOK, users)
}

// CreateUser creates a new user (superadmin only)
func (h *AuthHandler) CreateUser(c *gin.Context) {
	var req struct {
		Email     string `json:"email" binding:"required,email"`
		Name      string `json:"name" binding:"required"`
		Password  string `json:"password" binding:"required,min=8"`
		Role      string `json:"role" binding:"required"`
		IsActive  bool   `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate role and convert to UserRole type
	var userRole models.UserRole
	switch req.Role {
	case "user":
		userRole = models.RoleUser
	case "admin":
		userRole = models.RoleAdmin
	case "superadmin":
		userRole = models.RoleSuperAdmin
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	
	// Check if user already exists
	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	
	// Create user
	user := models.User{
		Email:        req.Email,
		Name:         req.Name,
		PasswordHash: string(hashedPassword),
		Role:         userRole,
		IsActive:     req.IsActive,
	}
	
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	
	// Don't return password hash
	user.PasswordHash = ""
	
	c.JSON(http.StatusCreated, user)
}

// UpdateUser updates a user's details (superadmin only)
func (h *AuthHandler) UpdateUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Role     string `json:"role" binding:"required"`
		IsActive bool   `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate role and convert to UserRole type
	var userRole models.UserRole
	switch req.Role {
	case "user":
		userRole = models.RoleUser
	case "admin":
		userRole = models.RoleAdmin
	case "superadmin":
		userRole = models.RoleSuperAdmin
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	
	// Check if email is already taken by another user
	var existingUser models.User
	if err := h.db.Where("email = ? AND id != ?", req.Email, uint(userID)).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Email is already taken by another user"})
		return
	}
	
	// Update user - use Select to ensure boolean false values are updated
	updates := map[string]interface{}{
		"email":     req.Email,
		"name":      req.Name,
		"role":      userRole,
		"is_active": req.IsActive,
	}
	
	result := h.db.Model(&models.User{}).Select("email", "name", "role", "is_active").Where("id = ?", uint(userID)).Updates(updates)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	// Fetch and return updated user
	var updatedUser models.User
	if err := h.db.First(&updatedUser, uint(userID)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch updated user"})
		return
	}
	
	// Don't return password hash
	updatedUser.PasswordHash = ""
	
	c.JSON(http.StatusOK, updatedUser)
}

// UpdateUserRole updates a user's role (superadmin only)
func (h *AuthHandler) UpdateUserRole(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	var req struct {
		Role string `json:"role" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate role and convert to UserRole type
	var userRole models.UserRole
	switch req.Role {
	case "user":
		userRole = models.RoleUser
	case "admin":
		userRole = models.RoleAdmin
	case "superadmin":
		userRole = models.RoleSuperAdmin
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role"})
		return
	}
	
	// Update user role
	result := h.db.Model(&models.User{}).Where("id = ?", uint(userID)).Update("role", userRole)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user role"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User role updated successfully"})
}

// DeleteUser deletes a user (superadmin only)
func (h *AuthHandler) DeleteUser(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	// Prevent superadmin from deleting themselves
	currentUserID := c.GetUint("user_id")
	if uint(userID) == currentUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete your own account"})
		return
	}
	
	result := h.db.Delete(&models.User{}, uint(userID))
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete user"})
		return
	}
	
	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "User deleted successfully"})
}

// AssignOrganizationAdmin assigns a user as admin to an organization (superadmin only)
func (h *AuthHandler) AssignOrganizationAdmin(c *gin.Context) {
	var req struct {
		UserID         uint   `json:"user_id" binding:"required"`
		OrganizationID uint   `json:"organization_id" binding:"required"`
		Role           string `json:"role"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	if req.Role == "" {
		req.Role = "admin"
	}
	
	// Get current user (superadmin)
	currentUserID := c.GetUint("user_id")
	
	// Check if user exists and is admin role
	var user models.User
	if err := h.db.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	
	if user.Role != models.RoleAdmin {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only admin users can be assigned to organizations"})
		return
	}
	
	// Check if organization exists
	var org models.Organization
	if err := h.db.First(&org, req.OrganizationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization not found"})
		return
	}
	
	// Check if assignment already exists
	var existingAdmin models.OrganizationAdmin
	err := h.db.Where("user_id = ? AND organization_id = ? AND is_active = true", req.UserID, req.OrganizationID).First(&existingAdmin).Error
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already admin of this organization"})
		return
	}
	
	// Create organization admin assignment
	orgAdmin := models.OrganizationAdmin{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		Role:           req.Role,
		IsActive:       true,
		AssignedBy:     currentUserID,
		AssignedAt:     time.Now(),
	}
	
	if err := h.db.Create(&orgAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign organization admin"})
		return
	}
	
	// Load the created assignment with relations
	if err := h.db.Preload("User").Preload("Organization").First(&orgAdmin, orgAdmin.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load assignment"})
		return
	}
	
	c.JSON(http.StatusCreated, orgAdmin)
}

// RevokeOrganizationAdmin revokes a user's admin access to an organization (superadmin only)
func (h *AuthHandler) RevokeOrganizationAdmin(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("user_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}
	
	orgID, err := strconv.ParseUint(c.Param("org_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid organization ID"})
		return
	}
	
	// Get current user (superadmin)
	currentUserID := c.GetUint("user_id")
	
	// Find active organization admin assignment
	var orgAdmin models.OrganizationAdmin
	if err := h.db.Where("user_id = ? AND organization_id = ? AND is_active = true", uint(userID), uint(orgID)).First(&orgAdmin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Organization admin assignment not found"})
		return
	}
	
	// Update assignment to revoked
	now := time.Now()
	orgAdmin.IsActive = false
	orgAdmin.RevokedBy = &currentUserID
	orgAdmin.RevokedAt = &now
	
	if err := h.db.Save(&orgAdmin).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to revoke organization admin"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Organization admin access revoked successfully"})
}

// ListOrganizationAdmins lists all organization admin assignments (superadmin only)
func (h *AuthHandler) ListOrganizationAdmins(c *gin.Context) {
	var orgAdmins []models.OrganizationAdmin
	
	query := h.db.Preload("User").Preload("Organization")
	
	// Optional filtering by organization
	if orgID := c.Query("organization_id"); orgID != "" {
		query = query.Where("organization_id = ?", orgID)
	}
	
	// Optional filtering by user
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	// Filter by active status (default: only active)
	if active := c.Query("active"); active != "false" {
		query = query.Where("is_active = true")
	}
	
	if err := query.Order("created_at DESC").Find(&orgAdmins).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch organization admins"})
		return
	}
	
	c.JSON(http.StatusOK, orgAdmins)
}