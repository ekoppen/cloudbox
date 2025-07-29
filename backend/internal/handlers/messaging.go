package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MessagingHandler handles messaging requests
type MessagingHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewMessagingHandler creates a new messaging handler
func NewMessagingHandler(db *gorm.DB, cfg *config.Config) *MessagingHandler {
	return &MessagingHandler{db: db, cfg: cfg}
}

// Channel Management

// ListChannels returns all channels for a project
func (h *MessagingHandler) ListChannels(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Parse query parameters
	limit := 25
	offset := 0
	orderBy := "last_activity DESC"
	
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
	
	var channels []models.Channel
	query := h.db.Where("project_id = ?", uint(projectID)).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	// Optional filters
	if channelType := c.Query("type"); channelType != "" {
		query = query.Where("type = ?", channelType)
	}
	
	if name := c.Query("name"); name != "" {
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}
	
	if active := c.Query("active"); active == "true" {
		query = query.Where("is_active = ?", true)
	} else if active == "false" {
		query = query.Where("is_active = ?", false)
	}
	
	if err := query.Find(&channels).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch channels"})
		return
	}
	
	// Get total count
	var total int64
	countQuery := h.db.Model(&models.Channel{}).Where("project_id = ?", project.ID)
	if channelType := c.Query("type"); channelType != "" {
		countQuery = countQuery.Where("type = ?", channelType)
	}
	if name := c.Query("name"); name != "" {
		countQuery = countQuery.Where("name ILIKE ?", "%"+name+"%")
	}
	if active := c.Query("active"); active == "true" {
		countQuery = countQuery.Where("is_active = ?", true)
	} else if active == "false" {
		countQuery = countQuery.Where("is_active = ?", false)
	}
	countQuery.Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"channels": channels,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// CreateChannel creates a new channel
func (h *MessagingHandler) CreateChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	
	var req struct {
		Name        string                 `json:"name" binding:"required"`
		Description string                 `json:"description"`
		Type        string                 `json:"type"`
		Topic       string                 `json:"topic"`
		MaxMembers  *int                   `json:"max_members"`
		Settings    map[string]interface{} `json:"settings"`
		CreatedBy   string                 `json:"created_by" binding:"required"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Validate channel type
	if req.Type == "" {
		req.Type = "public"
	}
	if req.Type != "public" && req.Type != "private" && req.Type != "direct" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid channel type"})
		return
	}
	
	// Validate creator exists
	var creator models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, req.CreatedBy).First(&creator).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Creator user not found"})
		return
	}
	
	// Check if channel name already exists for public channels
	if req.Type == "public" || req.Type == "private" {
		var existingChannel models.Channel
		if err := h.db.Where("project_id = ? AND name = ? AND type != 'direct'", project.ID, req.Name).First(&existingChannel).Error; err == nil {
			c.JSON(http.StatusConflict, gin.H{"error": "Channel with this name already exists"})
			return
		}
	}
	
	// Set defaults
	maxMembers := 0
	if req.MaxMembers != nil {
		maxMembers = *req.MaxMembers
	}
	
	// Create channel
	channel := models.Channel{
		ID:           uuid.New().String(),
		Name:         req.Name,
		Description:  req.Description,
		Type:         req.Type,
		Topic:        req.Topic,
		MaxMembers:   maxMembers,
		Settings:     req.Settings,
		ProjectID:    project.ID,
		CreatedBy:    req.CreatedBy,
		MemberCount:  1, // Creator is automatically a member
		MessageCount: 0,
		LastActivity: time.Now(),
	}
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	if err := tx.Create(&channel).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create channel"})
		return
	}
	
	// Add creator as channel owner
	membership := models.ChannelMember{
		ChannelID:    channel.ID,
		UserID:       req.CreatedBy,
		Role:         "owner",
		ProjectID:    project.ID,
		IsActive:     true,
		JoinedAt:     time.Now(),
		CanRead:      true,
		CanWrite:     true,
		CanInvite:    true,
		CanModerate:  true,
	}
	
	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add creator membership"})
		return
	}
	
	tx.Commit()
	c.JSON(http.StatusCreated, channel)
}

// GetChannel returns a specific channel
func (h *MessagingHandler) GetChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	var channel models.Channel
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	c.JSON(http.StatusOK, channel)
}

// UpdateChannel updates a channel
func (h *MessagingHandler) UpdateChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	var req struct {
		Name        *string                `json:"name"`
		Description *string                `json:"description"`
		Topic       *string                `json:"topic"`
		MaxMembers  *int                   `json:"max_members"`
		Settings    map[string]interface{} `json:"settings"`
		IsActive    *bool                  `json:"is_active"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find channel
	var channel models.Channel
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// Update fields
	updates := make(map[string]interface{})
	
	if req.Name != nil {
		// Check name uniqueness for non-direct channels
		if channel.Type != "direct" {
			var existingChannel models.Channel
			if err := h.db.Where("project_id = ? AND name = ? AND type != 'direct' AND id != ?", 
				project.ID, *req.Name, channelID).First(&existingChannel).Error; err == nil {
				c.JSON(http.StatusConflict, gin.H{"error": "Channel name already taken"})
				return
			}
		}
		updates["name"] = *req.Name
	}
	
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	
	if req.Topic != nil {
		updates["topic"] = *req.Topic
	}
	
	if req.MaxMembers != nil {
		updates["max_members"] = *req.MaxMembers
	}
	
	if req.Settings != nil {
		updates["settings"] = req.Settings
	}
	
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}
	
	if err := h.db.Model(&channel).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update channel"})
		return
	}
	
	// Return updated channel
	h.db.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel)
	c.JSON(http.StatusOK, channel)
}

// DeleteChannel deletes a channel and all its messages
func (h *MessagingHandler) DeleteChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	// Start transaction
	tx := h.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	
	// Delete all channel members
	if err := tx.Where("project_id = ? AND channel_id = ?", project.ID, channelID).Delete(&models.ChannelMember{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel members"})
		return
	}
	
	// Delete all messages in channel
	if err := tx.Where("project_id = ? AND channel_id = ?", project.ID, channelID).Delete(&models.Message{}).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete messages"})
		return
	}
	
	// Delete channel
	result := tx.Where("project_id = ? AND id = ?", project.ID, channelID).Delete(&models.Channel{})
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete channel"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Channel deleted successfully"})
}

// Channel Membership

// ListChannelMembers returns all members of a channel
func (h *MessagingHandler) ListChannelMembers(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	// Verify channel exists
	if !h.channelExists(project.ID, channelID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	var members []models.ChannelMember
	if err := h.db.Where("project_id = ? AND channel_id = ? AND is_active = ?", 
		project.ID, channelID, true).Find(&members).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch members"})
		return
	}
	
	c.JSON(http.StatusOK, members)
}

// JoinChannel adds a user to a channel
func (h *MessagingHandler) JoinChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	var req struct {
		UserID string `json:"user_id" binding:"required"`
		Role   string `json:"role"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Verify channel exists
	var channel models.Channel
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// Verify user exists
	var user models.AppUser
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, req.UserID).First(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User not found"})
		return
	}
	
	// Check if user is already a member
	var existingMember models.ChannelMember
	if err := h.db.Where("project_id = ? AND channel_id = ? AND user_id = ?", 
		project.ID, channelID, req.UserID).First(&existingMember).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User is already a member"})
		return
	}
	
	// Check max members limit
	if channel.MaxMembers > 0 && channel.MemberCount >= channel.MaxMembers {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Channel is full"})
		return
	}
	
	// Set default role
	role := "member"
	if req.Role != "" {
		if req.Role == "admin" || req.Role == "member" {
			role = req.Role
		}
	}
	
	// Create membership
	membership := models.ChannelMember{
		ChannelID:   channelID,
		UserID:      req.UserID,
		Role:        role,
		ProjectID:   project.ID,
		IsActive:    true,
		JoinedAt:    time.Now(),
		CanRead:     true,
		CanWrite:    true,
		CanInvite:   role == "admin",
		CanModerate: role == "admin",
	}
	
	// Start transaction
	tx := h.db.Begin()
	
	if err := tx.Create(&membership).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to join channel"})
		return
	}
	
	// Update channel member count
	tx.Model(&channel).Update("member_count", gorm.Expr("member_count + 1"))
	
	tx.Commit()
	c.JSON(http.StatusCreated, membership)
}

// LeaveChannel removes a user from a channel
func (h *MessagingHandler) LeaveChannel(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	userID := c.Param("user_id")
	
	// Start transaction
	tx := h.db.Begin()
	
	result := tx.Where("project_id = ? AND channel_id = ? AND user_id = ?", 
		project.ID, channelID, userID).Delete(&models.ChannelMember{})
	
	if result.Error != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to leave channel"})
		return
	}
	
	if result.RowsAffected == 0 {
		tx.Rollback()
		c.JSON(http.StatusNotFound, gin.H{"error": "Membership not found"})
		return
	}
	
	// Update channel member count
	var channel models.Channel
	if err := tx.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel).Error; err == nil {
		tx.Model(&channel).Update("member_count", gorm.Expr("member_count - 1"))
	}
	
	tx.Commit()
	c.JSON(http.StatusOK, gin.H{"message": "Left channel successfully"})
}

// Message Management

// ListMessages returns messages in a channel
func (h *MessagingHandler) ListMessages(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	// Verify channel exists
	if !h.channelExists(project.ID, channelID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// Parse query parameters
	limit := 50
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
	
	var messages []models.Message
	query := h.db.Where("project_id = ? AND channel_id = ? AND is_deleted = ?", 
		project.ID, channelID, false).
		Limit(limit).
		Offset(offset).
		Order(orderBy)
	
	// Optional filters
	if messageType := c.Query("type"); messageType != "" {
		query = query.Where("type = ?", messageType)
	}
	
	if userID := c.Query("user_id"); userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	
	if threadID := c.Query("thread_id"); threadID != "" {
		query = query.Where("thread_id = ?", threadID)
	}
	
	if err := query.Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	
	// Get total count
	var total int64
	countQuery := h.db.Model(&models.Message{}).Where("project_id = ? AND channel_id = ? AND is_deleted = ?", 
		project.ID, channelID, false)
	if messageType := c.Query("type"); messageType != "" {
		countQuery = countQuery.Where("type = ?", messageType)
	}
	if userID := c.Query("user_id"); userID != "" {
		countQuery = countQuery.Where("user_id = ?", userID)
	}
	if threadID := c.Query("thread_id"); threadID != "" {
		countQuery = countQuery.Where("thread_id = ?", threadID)
	}
	countQuery.Count(&total)
	
	c.JSON(http.StatusOK, gin.H{
		"messages": messages,
		"total":    total,
		"limit":    limit,
		"offset":   offset,
	})
}

// SendMessage sends a new message to a channel
func (h *MessagingHandler) SendMessage(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	
	var req struct {
		Content  string                 `json:"content" binding:"required"`
		Type     string                 `json:"type"`
		UserID   string                 `json:"user_id" binding:"required"`
		ParentID *string                `json:"parent_id"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Verify channel exists
	var channel models.Channel
	if err := h.db.Where("project_id = ? AND id = ?", project.ID, channelID).First(&channel).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Channel not found"})
		return
	}
	
	// Verify user exists and is channel member
	if !h.isChannelMember(project.ID, channelID, req.UserID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "User is not a member of this channel"})
		return
	}
	
	// Set default type
	messageType := "text"
	if req.Type != "" {
		messageType = req.Type
	}
	
	// Handle thread logic
	var threadID *string
	if req.ParentID != nil {
		// Verify parent message exists
		var parentMessage models.Message
		if err := h.db.Where("project_id = ? AND channel_id = ? AND id = ?", 
			project.ID, channelID, *req.ParentID).First(&parentMessage).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Parent message not found"})
			return
		}
		
		// Use parent's thread_id or parent's id as thread_id
		if parentMessage.ThreadID != nil {
			threadID = parentMessage.ThreadID
		} else {
			threadID = req.ParentID
		}
	}
	
	// Create message
	message := models.Message{
		ID:        uuid.New().String(),
		Content:   strings.TrimSpace(req.Content),
		Type:      messageType,
		Metadata:  req.Metadata,
		ChannelID: channelID,
		UserID:    req.UserID,
		ParentID:  req.ParentID,
		ThreadID:  threadID,
		ProjectID: project.ID,
	}
	
	// Start transaction
	tx := h.db.Begin()
	
	if err := tx.Create(&message).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}
	
	// Update channel stats
	now := time.Now()
	tx.Model(&channel).Updates(map[string]interface{}{
		"message_count":  gorm.Expr("message_count + 1"),
		"last_activity": now,
	})
	
	// Update parent message reply count if this is a reply
	if req.ParentID != nil {
		tx.Model(&models.Message{}).Where("id = ?", *req.ParentID).
			Update("reply_count", gorm.Expr("reply_count + 1"))
	}
	
	tx.Commit()
	c.JSON(http.StatusCreated, message)
}

// GetMessage returns a specific message
func (h *MessagingHandler) GetMessage(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	messageID := c.Param("message_id")
	
	var message models.Message
	if err := h.db.Where("project_id = ? AND channel_id = ? AND id = ? AND is_deleted = ?", 
		project.ID, channelID, messageID, false).First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	
	c.JSON(http.StatusOK, message)
}

// UpdateMessage updates a message
func (h *MessagingHandler) UpdateMessage(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	messageID := c.Param("message_id")
	
	var req struct {
		Content  string                 `json:"content" binding:"required"`
		Metadata map[string]interface{} `json:"metadata"`
	}
	
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	
	// Find message
	var message models.Message
	if err := h.db.Where("project_id = ? AND channel_id = ? AND id = ? AND is_deleted = ?", 
		project.ID, channelID, messageID, false).First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	
	// Update message
	now := time.Now()
	updates := map[string]interface{}{
		"content":   strings.TrimSpace(req.Content),
		"is_edited": true,
		"edited_at": &now,
	}
	
	if req.Metadata != nil {
		updates["metadata"] = req.Metadata
	}
	
	if err := h.db.Model(&message).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update message"})
		return
	}
	
	// Return updated message
	h.db.Where("project_id = ? AND channel_id = ? AND id = ?", project.ID, channelID, messageID).First(&message)
	c.JSON(http.StatusOK, message)
}

// DeleteMessage deletes a message (soft delete)
func (h *MessagingHandler) DeleteMessage(c *gin.Context) {
	project := c.MustGet("project").(models.Project)
	channelID := c.Param("channel_id")
	messageID := c.Param("message_id")
	
	// Find message
	var message models.Message
	if err := h.db.Where("project_id = ? AND channel_id = ? AND id = ? AND is_deleted = ?", 
		project.ID, channelID, messageID, false).First(&message).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}
	
	// Soft delete message
	now := time.Now()
	updates := map[string]interface{}{
		"is_deleted": true,
		"message_deleted_at": &now,
		"content":    "[Message deleted]",
	}
	
	if err := h.db.Model(&message).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{"message": "Message deleted successfully"})
}

// Helper functions

// channelExists checks if a channel exists
func (h *MessagingHandler) channelExists(projectID uint, channelID string) bool {
	var channel models.Channel
	err := h.db.Where("project_id = ? AND id = ?", projectID, channelID).First(&channel).Error
	return err == nil
}

// isChannelMember checks if a user is a member of a channel
func (h *MessagingHandler) isChannelMember(projectID uint, channelID, userID string) bool {
	var member models.ChannelMember
	err := h.db.Where("project_id = ? AND channel_id = ? AND user_id = ? AND is_active = ?", 
		projectID, channelID, userID, true).First(&member).Error
	return err == nil
}

// ListAllMessages returns all messages across all channels for the project
func (h *MessagingHandler) ListAllMessages(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	var messages []models.Message
	err = h.db.Where("project_id = ?", uint(projectID)).
		Order("created_at DESC").
		Limit(100). // Limit to recent 100 messages
		Find(&messages).Error
	
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}
	
	c.JSON(http.StatusOK, messages)
}

// ListTemplates returns message templates for the project
func (h *MessagingHandler) ListTemplates(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	_ = projectID // Use projectID if needed for database queries
	
	// For now, return default templates. In a real implementation,
	// these would be stored in the database per project
	templates := []gin.H{
		{
			"id": "welcome",
			"name": "Welcome Message",
			"content": "Welcome to {{project_name}}!",
			"type": "system",
		},
		{
			"id": "notification",
			"name": "Notification",
			"content": "You have a new notification: {{message}}",
			"type": "notification",
		},
	}
	
	c.JSON(http.StatusOK, templates)
}

// GetMessagingStats returns messaging statistics for the project
func (h *MessagingHandler) GetMessagingStats(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}
	
	// Count total channels
	var channelCount int64
	h.db.Model(&models.Channel{}).Where("project_id = ?", uint(projectID)).Count(&channelCount)
	
	// Count total messages
	var messageCount int64
	h.db.Model(&models.Message{}).Where("project_id = ?", uint(projectID)).Count(&messageCount)
	
	// Count active members across all channels
	var memberCount int64
	h.db.Model(&models.ChannelMember{}).Where("project_id = ? AND is_active = ?", uint(projectID), true).Count(&memberCount)
	
	stats := gin.H{
		"total_channels": channelCount,
		"total_messages": messageCount,
		"active_members": memberCount,
		"messages_today": 0, // Would need more complex query for this
		// Frontend expects these properties for the messaging dashboard
		"total_sent": messageCount,
		"total_delivered": messageCount,
		"total_opened": 0,
		"total_clicked": 0,
		"bounce_rate": 0,
		"open_rate": 0,
		"click_rate": 0,
	}
	
	c.JSON(http.StatusOK, stats)
}