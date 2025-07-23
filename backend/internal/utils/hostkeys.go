package utils

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// HostKeyEntry represents a known host key
type HostKeyEntry struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Hostname    string `json:"hostname" gorm:"not null;index"`
	Port        int    `json:"port" gorm:"not null;default:22"`
	KeyType     string `json:"key_type" gorm:"not null"`
	PublicKey   string `json:"public_key" gorm:"not null"`
	Fingerprint string `json:"fingerprint" gorm:"not null"`
	FirstSeen   int64  `json:"first_seen" gorm:"not null"`
	LastSeen    int64  `json:"last_seen" gorm:"not null"`
	Verified    bool   `json:"verified" gorm:"default:false"`
	ProjectID   uint   `json:"project_id" gorm:"not null;index"`
}

// HostKeyManager manages SSH host keys for secure connections
type HostKeyManager struct {
	db *gorm.DB
}

// NewHostKeyManager creates a new host key manager
func NewHostKeyManager(db *gorm.DB) *HostKeyManager {
	return &HostKeyManager{db: db}
}

// CreateHostKeyCallback creates a SSH host key verification callback
func (hm *HostKeyManager) CreateHostKeyCallback(projectID uint, allowNewHosts bool) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Extract port from remote address
		_, portStr, err := net.SplitHostPort(remote.String())
		if err != nil {
			return fmt.Errorf("failed to parse remote address: %w", err)
		}
		
		// For hostname, use the provided hostname parameter (not the IP from remote)
		keyType := key.Type()
		publicKeyBytes := key.Marshal()
		publicKeyB64 := base64.StdEncoding.EncodeToString(publicKeyBytes)
		fingerprint := hm.generateFingerprint(key)
		
		// Look for existing host key
		var hostKey HostKeyEntry
		err = hm.db.Where("hostname = ? AND key_type = ? AND project_id = ?", 
			hostname, keyType, projectID).First(&hostKey).Error
		
		if err == nil {
			// Host key exists, verify it matches
			if hostKey.PublicKey == publicKeyB64 {
				// Update last seen
				hm.db.Model(&hostKey).Update("last_seen", getCurrentTimestamp())
				return nil // Key matches, connection allowed
			} else {
				// Key changed - security risk!
				return fmt.Errorf("HOST KEY VERIFICATION FAILED: The host key for %s has changed! "+
					"This could indicate a man-in-the-middle attack. "+
					"Expected fingerprint: %s, Got: %s", 
					hostname, hostKey.Fingerprint, fingerprint)
			}
		}
		
		// Host key not found
		if !allowNewHosts {
			return fmt.Errorf("unknown host %s. Host key verification failed. "+
				"Fingerprint: %s. "+
				"To accept this host, enable 'allow new hosts' in your project settings", 
				hostname, fingerprint)
		}
		
		// Add new host key (unverified initially)
		port := 22
		if portStr != "22" && portStr != "" {
			// Parse port if it's not the default
			fmt.Sscanf(portStr, "%d", &port)
		}
		
		newHostKey := HostKeyEntry{
			Hostname:    hostname,
			Port:        port,
			KeyType:     keyType,
			PublicKey:   publicKeyB64,
			Fingerprint: fingerprint,
			FirstSeen:   getCurrentTimestamp(),
			LastSeen:    getCurrentTimestamp(),
			Verified:    false, // Requires manual verification
			ProjectID:   projectID,
		}
		
		if err := hm.db.Create(&newHostKey).Error; err != nil {
			return fmt.Errorf("failed to store host key: %w", err)
		}
		
		// Log warning about unverified host
		return fmt.Errorf("NEW HOST DETECTED: %s (fingerprint: %s). "+
			"This host has been added but requires manual verification for security. "+
			"Please verify the fingerprint and mark as trusted in your project settings",
			hostname, fingerprint)
	}
}

// VerifyHostKey manually verifies a host key (to be called by admin)
func (hm *HostKeyManager) VerifyHostKey(projectID uint, hostname string, keyType string) error {
	var hostKey HostKeyEntry
	err := hm.db.Where("hostname = ? AND key_type = ? AND project_id = ?", 
		hostname, keyType, projectID).First(&hostKey).Error
	
	if err != nil {
		return fmt.Errorf("host key not found")
	}
	
	return hm.db.Model(&hostKey).Update("verified", true).Error
}

// RemoveHostKey removes a host key (for key rotation scenarios)
func (hm *HostKeyManager) RemoveHostKey(projectID uint, hostname string, keyType string) error {
	return hm.db.Where("hostname = ? AND key_type = ? AND project_id = ?", 
		hostname, keyType, projectID).Delete(&HostKeyEntry{}).Error
}

// ListHostKeys returns all host keys for a project
func (hm *HostKeyManager) ListHostKeys(projectID uint) ([]HostKeyEntry, error) {
	var hostKeys []HostKeyEntry
	err := hm.db.Where("project_id = ?", projectID).Find(&hostKeys).Error
	return hostKeys, err
}

// generateFingerprint creates SHA256 and MD5 fingerprints for a public key
func (hm *HostKeyManager) generateFingerprint(key ssh.PublicKey) string {
	// Generate SHA256 fingerprint (modern standard)
	sha256sum := sha256.Sum256(key.Marshal())
	sha256fp := base64.StdEncoding.EncodeToString(sha256sum[:])
	sha256fp = strings.TrimRight(sha256fp, "=")
	
	// Generate MD5 fingerprint (legacy, but still commonly used)
	md5sum := md5.Sum(key.Marshal())
	md5fp := fmt.Sprintf("%x", md5sum)
	
	// Format MD5 with colons
	md5formatted := ""
	for i := 0; i < len(md5fp); i += 2 {
		if i > 0 {
			md5formatted += ":"
		}
		md5formatted += md5fp[i : i+2]
	}
	
	return fmt.Sprintf("SHA256:%s MD5:%s", sha256fp, md5formatted)
}

// getCurrentTimestamp returns current Unix timestamp
func getCurrentTimestamp() int64 {
	return time.Now().Unix()
}