package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParseProjectID extracts and validates project ID from URL parameter
func ParseProjectID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseOrganizationID extracts and validates organization ID from URL parameter
func ParseOrganizationID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseUserID extracts and validates user ID from URL parameter
func ParseUserID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseFunctionID extracts and validates function ID from URL parameter
func ParseFunctionID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseDeploymentID extracts and validates deployment ID from URL parameter
func ParseDeploymentID(c *gin.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// ParseAPIKeyID extracts and validates API key ID from URL parameter
func ParseAPIKeyID(c *gin.Context) (uint, error) {
	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(keyID), nil
}

// ParseSSHKeyID extracts and validates SSH key ID from URL parameter
func ParseSSHKeyID(c *gin.Context) (uint, error) {
	keyID, err := strconv.ParseUint(c.Param("key_id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(keyID), nil
}

// ParseServerID extracts and validates server ID from URL parameter
func ParseServerID(c *gin.Context) (uint, error) {
	serverID, err := strconv.ParseUint(c.Param("server_id"), 10, 32)
	if err != nil {
		return 0, err
	}
	return uint(serverID), nil
}