package handlers

import (
	"crypto/md5"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/utils"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// SSHKeyHandler handles SSH key management
type SSHKeyHandler struct {
	db  *gorm.DB
	cfg *config.Config
}

// NewSSHKeyHandler creates a new SSH key handler
func NewSSHKeyHandler(db *gorm.DB, cfg *config.Config) *SSHKeyHandler {
	return &SSHKeyHandler{db: db, cfg: cfg}
}

// CreateSSHKeyRequest represents a request to create an SSH key
type CreateSSHKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	KeyType     string `json:"key_type"` // rsa, ed25519
	KeySize     int    `json:"key_size"` // 2048, 4096
}

// UpdateSSHKeyRequest represents a request to update an SSH key
type UpdateSSHKeyRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

// ListSSHKeys returns all SSH keys for a project
func (h *SSHKeyHandler) ListSSHKeys(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var sshKeys []models.SSHKey
	if err := h.db.Where("project_id = ?", uint(projectID)).Find(&sshKeys).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SSH keys"})
		return
	}

	// Remove private keys from response for security
	for i := range sshKeys {
		sshKeys[i].PrivateKey = "" // Never return private keys in list
	}

	c.JSON(http.StatusOK, sshKeys)
}

// CreateSSHKey generates a new SSH key pair
func (h *SSHKeyHandler) CreateSSHKey(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.KeyType == "" {
		req.KeyType = "rsa"
	}
	if req.KeySize == 0 {
		req.KeySize = 2048
	}

	// Generate key pair
	privateKey, publicKey, fingerprint, err := h.generateSSHKeyPair(req.KeyType, req.KeySize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate SSH key pair"})
		return
	}

	// Encrypt private key before storing
	masterPassword := h.cfg.MasterKey // Should be set in config
	if masterPassword == "" {
		log.Printf("Warning: No master key configured, generating temporary key")
		var err error
		masterPassword, err = utils.GenerateMasterKey()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate encryption key"})
			return
		}
	}

	encryptedPrivateKey, err := utils.EncryptPrivateKey(privateKey, masterPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to encrypt private key"})
		return
	}

	// Create SSH key record with encrypted private key
	sshKey := models.SSHKey{
		Name:        req.Name,
		Description: req.Description,
		PublicKey:   publicKey,
		PrivateKey:  encryptedPrivateKey, // Now properly encrypted
		Fingerprint: fingerprint,
		KeyType:     req.KeyType,
		KeySize:     req.KeySize,
		ProjectID:   uint(projectID),
		IsActive:    true,
	}

	if err := h.db.Create(&sshKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create SSH key"})
		return
	}

	// Remove private key from response for security
	sshKey.PrivateKey = "" // Don't return private key in API response
	c.JSON(http.StatusCreated, sshKey)
}

// GetSSHKey returns a specific SSH key
func (h *SSHKeyHandler) GetSSHKey(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	var sshKey models.SSHKey
	if err := h.db.Where("id = ? AND project_id = ?", uint(keyID), uint(projectID)).First(&sshKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SSH key"})
		}
		return
	}

	// Remove private key from response for security
	sshKey.PrivateKey = "" // Never return private key in API response
	c.JSON(http.StatusOK, sshKey)
}

// GetDecryptedPrivateKey returns the decrypted private key for internal use only
// This function should NEVER be exposed as an API endpoint
func (h *SSHKeyHandler) GetDecryptedPrivateKey(keyID uint) (string, error) {
	var sshKey models.SSHKey
	if err := h.db.Where("id = ?", keyID).First(&sshKey).Error; err != nil {
		return "", err
	}

	if h.cfg.MasterKey == "" {
		return "", fmt.Errorf("master key not configured")
	}

	decryptedKey, err := utils.DecryptPrivateKey(sshKey.PrivateKey, h.cfg.MasterKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return decryptedKey, nil
}

// UpdateSSHKey updates an SSH key
func (h *SSHKeyHandler) UpdateSSHKey(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	var req UpdateSSHKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the SSH key
	var sshKey models.SSHKey
	if err := h.db.Where("id = ? AND project_id = ?", uint(keyID), uint(projectID)).First(&sshKey).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch SSH key"})
		}
		return
	}

	// Update the SSH key
	sshKey.Name = req.Name
	sshKey.Description = req.Description

	if err := h.db.Save(&sshKey).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update SSH key"})
		return
	}

	// Remove private key from response for security
	sshKey.PrivateKey = ""
	c.JSON(http.StatusOK, sshKey)
}

// DeleteSSHKey deletes an SSH key
func (h *SSHKeyHandler) DeleteSSHKey(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid key ID"})
		return
	}

	// Check if key is being used by any servers
	var serverCount int64
	if err := h.db.Model(&models.WebServer{}).Where("ssh_key_id = ?", uint(keyID)).Count(&serverCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to check key usage"})
		return
	}

	if serverCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete SSH key that is being used by servers"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(keyID), uint(projectID)).Delete(&models.SSHKey{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete SSH key"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "SSH key not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "SSH key deleted successfully"})
}

// generateSSHKeyPair generates a new SSH key pair
func (h *SSHKeyHandler) generateSSHKeyPair(keyType string, keySize int) (privateKey, publicKey, fingerprint string, err error) {
	switch keyType {
	case "rsa":
		return h.generateRSAKeyPair(keySize)
	default:
		return "", "", "", fmt.Errorf("unsupported key type: %s", keyType)
	}
}

// generateRSAKeyPair generates an RSA key pair
func (h *SSHKeyHandler) generateRSAKeyPair(keySize int) (privateKey, publicKey, fingerprint string, err error) {
	// Generate private key
	rsaPrivateKey, err := rsa.GenerateKey(rand.Reader, keySize)
	if err != nil {
		return "", "", "", err
	}

	// Encode private key to PEM format
	privateKeyPEM := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(rsaPrivateKey),
	}
	privateKeyBytes := pem.EncodeToMemory(privateKeyPEM)

	// Generate public key
	rsaPublicKey := &rsaPrivateKey.PublicKey
	sshPublicKey, err := ssh.NewPublicKey(rsaPublicKey)
	if err != nil {
		return "", "", "", err
	}

	// Encode public key to SSH format
	publicKeyBytes := ssh.MarshalAuthorizedKey(sshPublicKey)

	// Generate fingerprint
	fingerprintBytes := md5.Sum(sshPublicKey.Marshal())
	fingerprintStr := ""
	for i, b := range fingerprintBytes {
		if i > 0 {
			fingerprintStr += ":"
		}
		fingerprintStr += fmt.Sprintf("%02x", b)
	}

	return string(privateKeyBytes), string(publicKeyBytes), fingerprintStr, nil
}