package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Standard error messages
const (
	ErrInvalidProjectID     = "Invalid project ID"
	ErrInvalidOrganizationID = "Invalid organization ID"
	ErrInvalidUserID        = "Invalid user ID"
	ErrInvalidFunctionID    = "Invalid function ID"
	ErrInvalidDeploymentID  = "Invalid deployment ID"
	ErrInvalidAPIKeyID      = "Invalid API key ID"
	ErrInvalidSSHKeyID      = "Invalid SSH key ID"
	ErrInvalidServerID      = "Invalid server ID"
	ErrProjectNotFound      = "Project not found"
	ErrOrganizationNotFound = "Organization not found"
	ErrUserNotFound         = "User not found"
	ErrFunctionNotFound     = "Function not found"
	ErrDeploymentNotFound   = "Deployment not found"
	ErrAPIKeyNotFound       = "API key not found"
	ErrSSHKeyNotFound       = "SSH key not found"
	ErrServerNotFound       = "Server not found"
	ErrAccessDenied         = "Access denied"
	ErrInternalServer       = "Internal server error"
	ErrBadRequest           = "Bad request"
	ErrUnauthorized         = "Unauthorized"
	ErrForbidden            = "Forbidden"
)

// ErrorResponse represents a standard error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Code    string `json:"code,omitempty"`
	Details string `json:"details,omitempty"`
}

// ResponseBadRequest sends a 400 Bad Request response
func ResponseBadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, ErrorResponse{Error: message})
}

// ResponseUnauthorized sends a 401 Unauthorized response
func ResponseUnauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, ErrorResponse{Error: message})
}

// ResponseForbidden sends a 403 Forbidden response
func ResponseForbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, ErrorResponse{Error: message})
}

// ResponseNotFound sends a 404 Not Found response
func ResponseNotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, ErrorResponse{Error: message})
}

// ResponseConflict sends a 409 Conflict response
func ResponseConflict(c *gin.Context, message string) {
	c.JSON(http.StatusConflict, ErrorResponse{Error: message})
}

// ResponseInternalError sends a 500 Internal Server Error response
func ResponseInternalError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, ErrorResponse{Error: message})
}

// ResponseInvalidProjectID sends a standardized invalid project ID error
func ResponseInvalidProjectID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidProjectID)
}

// ResponseInvalidOrganizationID sends a standardized invalid organization ID error
func ResponseInvalidOrganizationID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidOrganizationID)
}

// ResponseInvalidUserID sends a standardized invalid user ID error
func ResponseInvalidUserID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidUserID)
}

// ResponseInvalidFunctionID sends a standardized invalid function ID error
func ResponseInvalidFunctionID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidFunctionID)
}

// ResponseInvalidDeploymentID sends a standardized invalid deployment ID error
func ResponseInvalidDeploymentID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidDeploymentID)
}

// ResponseInvalidAPIKeyID sends a standardized invalid API key ID error
func ResponseInvalidAPIKeyID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidAPIKeyID)
}

// ResponseInvalidSSHKeyID sends a standardized invalid SSH key ID error
func ResponseInvalidSSHKeyID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidSSHKeyID)
}

// ResponseInvalidServerID sends a standardized invalid server ID error
func ResponseInvalidServerID(c *gin.Context) {
	ResponseBadRequest(c, ErrInvalidServerID)
}

// ResponseProjectNotFound sends a standardized project not found error
func ResponseProjectNotFound(c *gin.Context) {
	ResponseNotFound(c, ErrProjectNotFound)
}

// ResponseOrganizationNotFound sends a standardized organization not found error
func ResponseOrganizationNotFound(c *gin.Context) {
	ResponseNotFound(c, ErrOrganizationNotFound)
}

// ResponseUserNotFound sends a standardized user not found error
func ResponseUserNotFound(c *gin.Context) {
	ResponseNotFound(c, ErrUserNotFound)
}

// ResponseSuccess sends a successful response with data
func ResponseSuccess(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}

// ResponseCreated sends a 201 Created response with data
func ResponseCreated(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, data)
}

// ResponseMessage sends a simple message response
func ResponseMessage(c *gin.Context, message string) {
	c.JSON(http.StatusOK, gin.H{"message": message})
}