package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// WebServerHandler handles web server management
type WebServerHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewWebServerHandler creates a new web server handler
func NewWebServerHandler(db *gorm.DB, cfg *config.Config) *WebServerHandler {
	return &WebServerHandler{db: db, cfg: cfg}
}

// CreateWebServerRequest represents a request to create a web server
type CreateWebServerRequest struct {
	Name          string `json:"name" binding:"required"`
	Hostname      string `json:"hostname" binding:"required"`
	Port          int    `json:"port"`
	Username      string `json:"username" binding:"required"`
	Description   string `json:"description"`
	ServerType    string `json:"server_type"`
	OS            string `json:"os"`
	DockerEnabled bool   `json:"docker_enabled"`
	NginxEnabled  bool   `json:"nginx_enabled"`
	DeployPath    string `json:"deploy_path"`
	BackupPath    string `json:"backup_path"`
	LogPath       string `json:"log_path"`
	SSHKeyID      uint   `json:"ssh_key_id" binding:"required"`
}

// UpdateWebServerRequest represents a request to update a web server
type UpdateWebServerRequest struct {
	Name          string `json:"name"`
	Hostname      string `json:"hostname"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	Description   string `json:"description"`
	ServerType    string `json:"server_type"`
	OS            string `json:"os"`
	DockerEnabled *bool  `json:"docker_enabled"`
	NginxEnabled  *bool  `json:"nginx_enabled"`
	DeployPath    string `json:"deploy_path"`
	BackupPath    string `json:"backup_path"`
	LogPath       string `json:"log_path"`
	SSHKeyID      *uint  `json:"ssh_key_id"`
	IsActive      *bool  `json:"is_active"`
}

// ListWebServers returns all web servers for a project
func (h *WebServerHandler) ListWebServers(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var webServers []models.WebServer
	if err := h.db.Where("project_id = ?", uint(projectID)).Preload("SSHKey").Find(&webServers).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch web servers"})
		return
	}

	c.JSON(http.StatusOK, webServers)
}

// CreateWebServer creates a new web server
func (h *WebServerHandler) CreateWebServer(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateWebServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify SSH key exists and belongs to the project
	var sshKey models.SSHKey
	if err := h.db.Where("id = ? AND project_id = ?", req.SSHKeyID, uint(projectID)).First(&sshKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{"error": "SSH key not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify SSH key"})
		}
		return
	}

	// Set defaults
	if req.Port == 0 {
		req.Port = 22
	}
	if req.ServerType == "" {
		req.ServerType = "vps"
	}
	if req.OS == "" {
		req.OS = "ubuntu"
	}
	if req.DeployPath == "" {
		req.DeployPath = "/var/www"
	}
	if req.BackupPath == "" {
		req.BackupPath = "/var/backups"
	}
	if req.LogPath == "" {
		req.LogPath = "/var/log/deployments"
	}

	// Create web server record
	webServer := models.WebServer{
		Name:          req.Name,
		Hostname:      req.Hostname,
		Port:          req.Port,
		Username:      req.Username,
		Description:   req.Description,
		ServerType:    req.ServerType,
		OS:            req.OS,
		DockerEnabled: req.DockerEnabled,
		NginxEnabled:  req.NginxEnabled,
		DeployPath:    req.DeployPath,
		BackupPath:    req.BackupPath,
		LogPath:       req.LogPath,
		ProjectID:     uint(projectID),
		SSHKeyID:      req.SSHKeyID,
		IsActive:      true,
		ConnectionStatus: "unknown",
	}

	if err := h.db.Create(&webServer).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create web server"})
		return
	}

	// Load SSH key for response
	if err := h.db.Preload("SSHKey").First(&webServer, webServer.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load web server details"})
		return
	}

	c.JSON(http.StatusCreated, webServer)
}

// GetWebServer returns a specific web server
func (h *WebServerHandler) GetWebServer(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var webServer models.WebServer
	if err := h.db.Where("id = ? AND project_id = ?", uint(serverID), uint(projectID)).Preload("SSHKey").First(&webServer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Web server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch web server"})
		}
		return
	}

	c.JSON(http.StatusOK, webServer)
}

// UpdateWebServer updates a web server
func (h *WebServerHandler) UpdateWebServer(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	var req UpdateWebServerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find web server
	var webServer models.WebServer
	if err := h.db.Where("id = ? AND project_id = ?", uint(serverID), uint(projectID)).First(&webServer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Web server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch web server"})
		}
		return
	}

	// Update fields
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Hostname != "" {
		updates["hostname"] = req.Hostname
	}
	if req.Port != 0 {
		updates["port"] = req.Port
	}
	if req.Username != "" {
		updates["username"] = req.Username
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.ServerType != "" {
		updates["server_type"] = req.ServerType
	}
	if req.OS != "" {
		updates["os"] = req.OS
	}
	if req.DockerEnabled != nil {
		updates["docker_enabled"] = *req.DockerEnabled
	}
	if req.NginxEnabled != nil {
		updates["nginx_enabled"] = *req.NginxEnabled
	}
	if req.DeployPath != "" {
		updates["deploy_path"] = req.DeployPath
	}
	if req.BackupPath != "" {
		updates["backup_path"] = req.BackupPath
	}
	if req.LogPath != "" {
		updates["log_path"] = req.LogPath
	}
	if req.SSHKeyID != nil {
		// Verify SSH key exists and belongs to the project
		var sshKey models.SSHKey
		if err := h.db.Where("id = ? AND project_id = ?", *req.SSHKeyID, uint(projectID)).First(&sshKey).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusBadRequest, gin.H{"error": "SSH key not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to verify SSH key"})
			}
			return
		}
		updates["ssh_key_id"] = *req.SSHKeyID
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.db.Model(&webServer).Updates(updates).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update web server"})
		return
	}

	// Reload with SSH key
	if err := h.db.Preload("SSHKey").First(&webServer, webServer.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to load updated web server"})
		return
	}

	c.JSON(http.StatusOK, webServer)
}

// DeleteWebServer deletes a web server
func (h *WebServerHandler) DeleteWebServer(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	// Check if server is being used by any deployments
	var deploymentCount int64
	if err := h.db.Model(&models.Deployment{}).Where("web_server_id = ?", uint(serverID)).Count(&deploymentCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check server usage"})
		return
	}

	if deploymentCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete web server that has active deployments"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(serverID), uint(projectID)).Delete(&models.WebServer{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete web server"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Web server not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Web server deleted successfully"})
}

// TestConnection tests the SSH connection to a web server
func (h *WebServerHandler) TestConnection(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid server ID"})
		return
	}

	// Find web server with SSH key
	var webServer models.WebServer
	if err := h.db.Where("id = ? AND project_id = ?", uint(serverID), uint(projectID)).Preload("SSHKey").First(&webServer).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Web server not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch web server"})
		}
		return
	}

	// Test SSH connection
	connected, err := h.testSSHConnection(webServer)
	if err != nil {
		// Update connection status
		h.db.Model(&webServer).Updates(map[string]interface{}{
			"connection_status": "error",
			"last_connected_at": time.Now(),
		})

		c.JSON(http.StatusOK, gin.H{
			"connected": false,
			"error":     err.Error(),
		})
		return
	}

	// Update connection status
	status := "disconnected"
	if connected {
		status = "connected"
	}

	h.db.Model(&webServer).Updates(map[string]interface{}{
		"connection_status": status,
		"last_connected_at": time.Now(),
	})

	c.JSON(http.StatusOK, gin.H{
		"connected": connected,
		"message":   fmt.Sprintf("Connection test %s", status),
	})
}

// testSSHConnection tests SSH connection to a web server
func (h *WebServerHandler) testSSHConnection(webServer models.WebServer) (bool, error) {
	// Parse private key
	signer, err := ssh.ParsePrivateKey([]byte(webServer.SSHKey.PrivateKey))
	if err != nil {
		return false, fmt.Errorf("failed to parse private key: %w", err)
	}

	// Configure SSH client
	config := &ssh.ClientConfig{
		User: webServer.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         10 * time.Second,
	}

	// Connect to server
	addr := fmt.Sprintf("%s:%d", webServer.Hostname, webServer.Port)
	client, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return false, fmt.Errorf("failed to connect: %w", err)
	}
	defer client.Close()

	// Test simple command
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Run a simple command to verify connection
	err = session.Run("echo 'connection test'")
	if err != nil {
		return false, fmt.Errorf("failed to run test command: %w", err)
	}

	return true, nil
}