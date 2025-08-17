package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"gorm.io/gorm"
)

type PluginRegistryHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

func NewPluginRegistryHandler(db *gorm.DB, cfg *config.Config) *PluginRegistryHandler {
	return &PluginRegistryHandler{
		db:  db,
		cfg: cfg,
	}
}

type PluginSubmission struct {
	PluginName    string                 `json:"plugin_name" binding:"required"`
	Repository    string                 `json:"repository" binding:"required"`
	Version       string                 `json:"version" binding:"required"`
	Description   string                 `json:"description" binding:"required"`
	Author        string                 `json:"author" binding:"required"`
	Category      string                 `json:"category" binding:"required"`
	Tags          []string               `json:"tags"`
	License       string                 `json:"license" binding:"required"`
	Website       string                 `json:"website,omitempty"`
	SupportEmail  string                 `json:"support_email" binding:"required,email"`
	Screenshots   []string               `json:"screenshots,omitempty"`
	DemoURL       string                 `json:"demo_url,omitempty"`
	Permissions   []string               `json:"permissions" binding:"required"`
	Dependencies  map[string]string      `json:"dependencies,omitempty"`
	Configuration map[string]interface{} `json:"configuration,omitempty"`
}

type ApprovedRepositoryRequest struct {
	RepositoryURL   string `json:"repository_url" binding:"required"`
	OrganizationName string `json:"organization_name" binding:"required"`
	ContactEmail    string `json:"contact_email" binding:"required,email"`
	SecurityLevel   string `json:"security_level"` // high, medium, low
	AutoApprove     bool   `json:"auto_approve"`
	Verified        bool   `json:"verified"`
	Reason          string `json:"reason" binding:"required"`
}

// SubmitPlugin handles new plugin submissions
func (h *PluginRegistryHandler) SubmitPlugin(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	// Allow developers to submit plugins
	if userRole != "admin" && userRole != "superadmin" && userRole != "developer" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Developer, admin or superadmin access required",
		})
		return
	}

	var submission PluginSubmission
	if err := c.ShouldBindJSON(&submission); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid submission data: " + err.Error(),
		})
		return
	}

	// Validate repository URL
	if !h.isValidRepositoryURL(submission.Repository) {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid repository URL format",
		})
		return
	}

	// Parse user ID
	userIDInt, _ := strconv.ParseUint(userID, 10, 32)

	// Create plugin submission record
	pluginSubmission := models.PluginSubmission{
		PluginName:    submission.PluginName,
		Repository:    submission.Repository,
		Version:       submission.Version,
		Description:   submission.Description,
		Author:        submission.Author,
		Category:      submission.Category,
		Tags:          submission.Tags,
		License:       submission.License,
		Website:       submission.Website,
		SupportEmail:  submission.SupportEmail,
		Screenshots:   submission.Screenshots,
		DemoURL:       submission.DemoURL,
		Permissions:   submission.Permissions,
		Dependencies:  convertStringMapToInterface(submission.Dependencies),
		Configuration: submission.Configuration,
		Status:        "submitted",
		SubmittedBy:   uint(userIDInt),
		SubmittedAt:   time.Now(),
	}

	if err := h.db.Create(&pluginSubmission).Error; err != nil {
		log.Printf("Failed to create plugin submission: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to submit plugin",
		})
		return
	}

	// Send notification to review team
	h.notifyReviewTeam(pluginSubmission)

	// Log submission
	log.Printf("Plugin submission received: %s by %s (%s)", submission.PluginName, userEmail, submission.Repository)

	c.JSON(http.StatusOK, gin.H{
		"success":       true,
		"submission_id": pluginSubmission.ID,
		"message":       "Plugin submitted successfully for review",
		"timeline":      "Review typically takes 10-18 business days",
	})
}

// GetSubmissions lists plugin submissions (admin only)
func (h *PluginRegistryHandler) GetSubmissions(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	// Query parameters
	status := c.Query("status")
	limit := c.DefaultQuery("limit", "50")
	offset := c.DefaultQuery("offset", "0")

	var submissions []models.PluginSubmission
	query := h.db.Model(&models.PluginSubmission{})

	// Filter by status
	if status != "" {
		query = query.Where("status = ?", status)
	}

	// Apply pagination
	limitInt, _ := strconv.Atoi(limit)
	offsetInt, _ := strconv.Atoi(offset)
	if limitInt > 100 {
		limitInt = 100
	}

	query = query.Order("submitted_at DESC").Limit(limitInt).Offset(offsetInt)

	if err := query.Find(&submissions).Error; err != nil {
		log.Printf("Failed to fetch submissions: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch submissions",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"submissions": submissions,
		"count":       len(submissions),
	})
}

// ReviewSubmission handles plugin submission review decisions
func (h *PluginRegistryHandler) ReviewSubmission(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	submissionIDStr := c.Param("submissionId")
	submissionID, err := strconv.ParseUint(submissionIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid submission ID",
		})
		return
	}

	var reviewData struct {
		Action   string `json:"action" binding:"required"` // approve, reject, request_changes
		Comments string `json:"comments"`
		Score    int    `json:"score"` // 1-10
	}

	if err := c.ShouldBindJSON(&reviewData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid review data: " + err.Error(),
		})
		return
	}

	// Get submission
	var submission models.PluginSubmission
	if err := h.db.First(&submission, submissionID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Submission not found",
		})
		return
	}

	// Parse user ID
	userIDInt, _ := strconv.ParseUint(userID, 10, 32)

	// Update submission status
	submission.Status = reviewData.Action + "d" // approved, rejected, request_changesd
	submission.ReviewComments = reviewData.Comments
	submission.ReviewScore = reviewData.Score
	submission.ReviewedBy = uint(userIDInt)
	submission.ReviewedAt = time.Now()

	if err := h.db.Save(&submission).Error; err != nil {
		log.Printf("Failed to update submission: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update submission",
		})
		return
	}

	// If approved, add to marketplace
	if reviewData.Action == "approve" {
		h.addToMarketplace(submission)
	}

	// Notify submitter
	h.notifySubmitter(submission, reviewData.Action)

	// Log review action
	log.Printf("Plugin submission %s %s by %s: %s", submission.PluginName, reviewData.Action, userEmail, reviewData.Comments)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Submission %s successfully", reviewData.Action+"d"),
	})
}

// RequestRepositoryApproval handles requests for repository approval
func (h *PluginRegistryHandler) RequestRepositoryApproval(c *gin.Context) {
	_ = c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	var request ApprovedRepositoryRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request data: " + err.Error(),
		})
		return
	}

	// Parse user ID
	userIDInt, _ := strconv.ParseUint(userID, 10, 32)

	// Create repository approval request
	repoRequest := models.RepositoryApprovalRequest{
		RepositoryURL:    request.RepositoryURL,
		OrganizationName: request.OrganizationName,
		ContactEmail:     request.ContactEmail,
		SecurityLevel:    request.SecurityLevel,
		AutoApprove:      request.AutoApprove,
		Verified:         false, // Always false for requests
		Reason:           request.Reason,
		Status:           "pending",
		RequestedBy:      uint(userIDInt),
		RequestedAt:      time.Now(),
	}

	if err := h.db.Create(&repoRequest).Error; err != nil {
		log.Printf("Failed to create repository request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to submit repository approval request",
		})
		return
	}

	// Notify admin team
	h.notifyAdminTeam(repoRequest)

	log.Printf("Repository approval requested: %s by %s", request.RepositoryURL, userEmail)

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"request_id": repoRequest.ID,
		"message":    "Repository approval request submitted successfully",
		"timeline":   "Review typically takes 5-10 business days",
	})
}

// GetApprovedRepositories returns list of approved repositories
func (h *PluginRegistryHandler) GetApprovedRepositories(c *gin.Context) {
	userRole := c.GetString("user_role")
	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	var repositories []models.ApprovedRepository
	err := h.db.Where("is_active = ?", true).Find(&repositories).Error
	if err != nil {
		log.Printf("Error fetching approved repositories: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to fetch approved repositories",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"repositories": repositories,
		"count":        len(repositories),
	})
}

// ApproveRepository approves a repository for plugin submissions
func (h *PluginRegistryHandler) ApproveRepository(c *gin.Context) {
	userRole := c.GetString("user_role")
	userID := c.GetString("user_id")
	userEmail := c.GetString("user_email")

	if userRole != "admin" && userRole != "superadmin" {
		c.JSON(http.StatusForbidden, gin.H{
			"success": false,
			"error":   "Admin access required",
		})
		return
	}

	requestIDStr := c.Param("requestId")
	requestID, err := strconv.ParseUint(requestIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid request ID",
		})
		return
	}

	var approvalData struct {
		Action        string `json:"action" binding:"required"` // approve, reject
		SecurityLevel string `json:"security_level"`
		AutoApprove   bool   `json:"auto_approve"`
		Verified      bool   `json:"verified"`
		Comments      string `json:"comments"`
	}

	if err := c.ShouldBindJSON(&approvalData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "Invalid approval data: " + err.Error(),
		})
		return
	}

	// Get request
	var request models.RepositoryApprovalRequest
	if err := h.db.First(&request, requestID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "Request not found",
		})
		return
	}

	// Parse user ID
	userIDInt, _ := strconv.ParseUint(userID, 10, 32)

	// Update request status
	request.Status = approvalData.Action + "d"
	request.ReviewComments = approvalData.Comments
	request.ReviewedBy = uint(userIDInt)
	request.ReviewedAt = time.Now()

	if err := h.db.Save(&request).Error; err != nil {
		log.Printf("Failed to update repository request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "Failed to update request",
		})
		return
	}

	// If approved, add to approved repositories
	if approvalData.Action == "approve" {
		approvedRepo := models.ApprovedRepository{
			RepositoryURL:   request.RepositoryURL,
			RepositoryOwner: request.OrganizationName,
			RepositoryName:  strings.Split(request.RepositoryURL, "/")[len(strings.Split(request.RepositoryURL, "/"))-1],
			ApprovedBy:      uint(userIDInt),
			ApprovedAt:      time.Now(),
			ApprovalReason:  approvalData.Action,
		}

		if err := h.db.Create(&approvedRepo).Error; err != nil {
			log.Printf("Failed to create approved repository: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "Failed to add approved repository",
			})
			return
		}
	}

	// Notify requester
	h.notifyRequester(request, approvalData.Action)

	log.Printf("Repository approval %s: %s by %s", approvalData.Action, request.RepositoryURL, userEmail)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": fmt.Sprintf("Repository %s successfully", approvalData.Action+"d"),
	})
}

// Helper functions

func (h *PluginRegistryHandler) isValidRepositoryURL(url string) bool {
	// Basic GitHub URL validation
	return strings.HasPrefix(url, "https://github.com/") && len(strings.Split(url, "/")) >= 5
}

func (h *PluginRegistryHandler) notifyReviewTeam(submission models.PluginSubmission) {
	// TODO: Implement email notification to review team
	log.Printf("NOTIFICATION: New plugin submission - %s", submission.PluginName)
}

func (h *PluginRegistryHandler) notifySubmitter(submission models.PluginSubmission, action string) {
	// TODO: Implement email notification to submitter
	log.Printf("NOTIFICATION: Plugin %s %s - %s", submission.PluginName, action, submission.SupportEmail)
}

func (h *PluginRegistryHandler) notifyAdminTeam(request models.RepositoryApprovalRequest) {
	// TODO: Implement email notification to admin team
	log.Printf("NOTIFICATION: Repository approval request - %s", request.RepositoryURL)
}

func (h *PluginRegistryHandler) notifyRequester(request models.RepositoryApprovalRequest, action string) {
	// TODO: Implement email notification to requester
	log.Printf("NOTIFICATION: Repository %s %s - %s", request.RepositoryURL, action, request.ContactEmail)
}

func (h *PluginRegistryHandler) addToMarketplace(submission models.PluginSubmission) {
	// Create marketplace entry
	marketplaceEntry := models.PluginMarketplace{
		PluginName:   submission.PluginName,
		Repository:   submission.Repository,
		Version:      submission.Version,
		Description:  submission.Description,
		Author:       submission.Author,
		Category:     submission.Category,
		Tags:         submission.Tags,
		License:      submission.License,
		Website:      submission.Website,
		SupportEmail: submission.SupportEmail,
		Screenshots:  submission.Screenshots,
		DemoURL:      submission.DemoURL,
		Permissions:  submission.Permissions,
		Dependencies: submission.Dependencies,
		Verified:     false, // Manual verification required
		Official:     false, // Only CloudBox team can mark as official
		Featured:     false, // Admin decision
		Status:       "active",
		PublishedAt:  time.Now(),
	}

	if err := h.db.Create(&marketplaceEntry).Error; err != nil {
		log.Printf("Failed to add to marketplace: %v", err)
	} else {
		log.Printf("Plugin added to marketplace: %s", submission.PluginName)
	}
}

// Helper function to convert map[string]string to map[string]interface{}
func convertStringMapToInterface(input map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range input {
		result[k] = v
	}
	return result
}