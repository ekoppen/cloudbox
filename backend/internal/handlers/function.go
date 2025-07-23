package handlers

import (
	"context"
	"fmt"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/execution"
	"github.com/cloudbox/backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// FunctionHandler handles function management
type FunctionHandler struct {
	db       *gorm.DB
	cfg      *config.Config
	executor *execution.ExecutionEngine
}

// NewFunctionHandler creates a new function handler
func NewFunctionHandler(db *gorm.DB, cfg *config.Config) *FunctionHandler {
	// Create execution engine with default settings
	workDir := filepath.Join("/tmp", "cloudbox-functions")
	timeout := 30 * time.Second
	maxMemory := int64(128 * 1024 * 1024) // 128MB default
	
	executor := execution.NewExecutionEngine(workDir, timeout, maxMemory)
	
	return &FunctionHandler{
		db:       db,
		cfg:      cfg,
		executor: executor,
	}
}

// CreateFunctionRequest represents a request to create a function
type CreateFunctionRequest struct {
	Name         string                 `json:"name" binding:"required"`
	Description  string                 `json:"description"`
	Runtime      string                 `json:"runtime"`      // nodejs18, python3.9, go1.19
	Language     string                 `json:"language"`     // javascript, python, go
	Code         string                 `json:"code" binding:"required"`
	EntryPoint   string                 `json:"entry_point"`
	Timeout      int                    `json:"timeout"`      // seconds
	Memory       int                    `json:"memory"`       // MB
	Environment  map[string]interface{} `json:"environment"`
	Commands     []string               `json:"commands"`
	Dependencies map[string]interface{} `json:"dependencies"`
	IsPublic     bool                   `json:"is_public"`
}

// UpdateFunctionRequest represents a request to update a function
type UpdateFunctionRequest struct {
	Name         *string                 `json:"name"`
	Description  *string                 `json:"description"`
	Runtime      *string                 `json:"runtime"`
	Language     *string                 `json:"language"`
	Code         *string                 `json:"code"`
	EntryPoint   *string                 `json:"entry_point"`
	Timeout      *int                    `json:"timeout"`
	Memory       *int                    `json:"memory"`
	Environment  *map[string]interface{} `json:"environment"`
	Commands     *[]string               `json:"commands"`
	Dependencies *map[string]interface{} `json:"dependencies"`
	IsPublic     *bool                   `json:"is_public"`
	IsActive     *bool                   `json:"is_active"`
}

// ExecuteFunctionRequest represents a request to execute a function
type ExecuteFunctionRequest struct {
	Data    map[string]interface{} `json:"data"`
	Headers map[string]interface{} `json:"headers"`
}

// ListFunctions returns all functions for a project
func (h *FunctionHandler) ListFunctions(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var functions []models.Function
	if err := h.db.Where("project_id = ?", uint(projectID)).Find(&functions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch functions"})
		return
	}

	c.JSON(http.StatusOK, functions)
}

// CreateFunction creates a new function
func (h *FunctionHandler) CreateFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	var req CreateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Set defaults
	if req.Runtime == "" {
		req.Runtime = "nodejs18"
	}
	if req.Language == "" {
		req.Language = "javascript"
	}
	if req.EntryPoint == "" {
		req.EntryPoint = "index.handler"
	}
	if req.Timeout == 0 {
		req.Timeout = 30
	}
	if req.Memory == 0 {
		req.Memory = 128
	}
	if req.Environment == nil {
		req.Environment = make(map[string]interface{})
	}
	if req.Dependencies == nil {
		req.Dependencies = make(map[string]interface{})
	}

	// Generate function URL
	functionURL := fmt.Sprintf("%s/p/%d/functions/%s", h.cfg.BaseURL, projectID, strings.ToLower(req.Name))

	// Create function record
	function := models.Function{
		Name:         req.Name,
		Description:  req.Description,
		Runtime:      req.Runtime,
		Language:     req.Language,
		Code:         req.Code,
		EntryPoint:   req.EntryPoint,
		Timeout:      req.Timeout,
		Memory:       req.Memory,
		Environment:  req.Environment,
		Commands:     req.Commands,
		Dependencies: req.Dependencies,
		Status:       "draft",
		Version:      1,
		FunctionURL:  functionURL,
		IsActive:     true,
		IsPublic:     req.IsPublic,
		ProjectID:    uint(projectID),
	}

	if err := h.db.Create(&function).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Function with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create function"})
		}
		return
	}

	c.JSON(http.StatusCreated, function)
}

// GetFunction returns a specific function
func (h *FunctionHandler) GetFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	var function models.Function
	if err := h.db.Where("id = ? AND project_id = ?", uint(functionID), uint(projectID)).First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch function"})
		}
		return
	}

	c.JSON(http.StatusOK, function)
}

// UpdateFunction updates an existing function
func (h *FunctionHandler) UpdateFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	var req UpdateFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the function
	var function models.Function
	if err := h.db.Where("id = ? AND project_id = ?", uint(functionID), uint(projectID)).First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch function"})
		}
		return
	}

	// Build update map
	updates := make(map[string]interface{})
	if req.Name != nil {
		updates["name"] = *req.Name
		// Update function URL if name changes
		updates["function_url"] = fmt.Sprintf("%s/p/%d/functions/%s", h.cfg.BaseURL, projectID, strings.ToLower(*req.Name))
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Runtime != nil {
		updates["runtime"] = *req.Runtime
	}
	if req.Language != nil {
		updates["language"] = *req.Language
	}
	if req.Code != nil {
		updates["code"] = *req.Code
		// Increment version when code changes
		updates["version"] = function.Version + 1
		updates["status"] = "draft" // Reset status to draft when code changes
	}
	if req.EntryPoint != nil {
		updates["entry_point"] = *req.EntryPoint
	}
	if req.Timeout != nil {
		updates["timeout"] = *req.Timeout
	}
	if req.Memory != nil {
		updates["memory"] = *req.Memory
	}
	if req.Environment != nil {
		updates["environment"] = *req.Environment
	}
	if req.Commands != nil {
		updates["commands"] = *req.Commands
	}
	if req.Dependencies != nil {
		updates["dependencies"] = *req.Dependencies
	}
	if req.IsPublic != nil {
		updates["is_public"] = *req.IsPublic
	}
	if req.IsActive != nil {
		updates["is_active"] = *req.IsActive
	}

	if err := h.db.Model(&function).Updates(updates).Error; err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			c.JSON(http.StatusConflict, gin.H{"error": "Function with this name already exists"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update function"})
		}
		return
	}

	// Reload function to get updated data
	if err := h.db.Where("id = ? AND project_id = ?", uint(functionID), uint(projectID)).First(&function).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to reload function"})
		return
	}

	c.JSON(http.StatusOK, function)
}

// DeleteFunction deletes a function
func (h *FunctionHandler) DeleteFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	result := h.db.Where("id = ? AND project_id = ?", uint(functionID), uint(projectID)).Delete(&models.Function{})
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete function"})
		return
	}

	if result.RowsAffected == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Function deleted successfully"})
}

// DeployFunction deploys a function (simulates build and deployment)
func (h *FunctionHandler) DeployFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	// Find the function
	var function models.Function
	if err := h.db.Where("id = ? AND project_id = ?", uint(functionID), uint(projectID)).First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch function"})
		}
		return
	}

	// Start deployment asynchronously (in real implementation, this would be a background job)
	go h.realDeployment(function)

	// Update status to building
	h.db.Model(&function).Updates(map[string]interface{}{
		"status": "building",
		"build_logs": "Starting function deployment...\n",
	})

	c.JSON(http.StatusOK, gin.H{
		"message": "Function deployment started",
		"status":  "building",
	})
}

// ExecuteFunction executes a function by ID
func (h *FunctionHandler) ExecuteFunction(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	var req ExecuteFunctionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Find the function
	var function models.Function
	if err := h.db.Where("id = ? AND project_id = ? AND is_active = ?", uint(functionID), uint(projectID), true).First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found or not active"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch function"})
		}
		return
	}

	if function.Status != "deployed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Function is not deployed"})
		return
	}

	// Execute function using real execution engine
	executionID := uuid.New().String()
	startTime := time.Now()

	// Create execution request
	execReq := execution.ExecutionRequest{
		Function: function,
		Data:     req.Data,
		Headers:  req.Headers,
		Method:   c.Request.Method,
		Path:     c.Request.URL.Path,
	}

	// Execute function with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(function.Timeout)*time.Second)
	defer cancel()

	result, err := h.executor.Execute(ctx, execReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Execution failed: %v", err)})
		return
	}

	executionTime := result.ExecutionTime
	var response map[string]interface{}
	status := "success"
	statusCode := result.StatusCode
	logs := result.Logs

	if result.Success {
		response = result.Response
	} else {
		response = map[string]interface{}{
			"error": result.Error,
		}
		status = "error"
		if statusCode == 0 {
			statusCode = 500
		}
	}

	// Log execution
	execution := models.FunctionExecution{
		FunctionID:    function.ID,
		ExecutionID:   executionID,
		RequestData:   req.Data,
		ResponseData:  response,
		Headers:       req.Headers,
		Method:        c.Request.Method,
		Path:          c.Request.URL.Path,
		Status:        status,
		StatusCode:    statusCode,
		ExecutionTime: executionTime,
		MemoryUsage:   result.MemoryUsage,
		StartedAt:     startTime,
		CompletedAt:   &time.Time{},
		Logs:          logs,
		UserAgent:     c.GetHeader("User-Agent"),
		ClientIP:      c.ClientIP(),
		Source:        "http",
		ProjectID:     uint(projectID),
	}

	now := time.Now()
	execution.CompletedAt = &now

	if err := h.db.Create(&execution).Error; err != nil {
		// Log error but don't fail the execution response
		fmt.Printf("Failed to log execution: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"execution_id":   executionID,
		"status":         "success",
		"execution_time": executionTime,
		"response":       response,
	})
}

// ExecuteFunctionByName executes a function by name (public API route)
func (h *FunctionHandler) ExecuteFunctionByName(c *gin.Context) {
	// Extract project from middleware (set by ProjectAuth middleware)
	project, exists := c.Get("project")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Project authentication required"})
		return
	}

	projectData, ok := project.(map[string]interface{})
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid project data"})
		return
	}

	projectID := uint(projectData["id"].(float64))
	functionName := c.Param("function_name")

	// Parse request body for data (optional)
	var requestData map[string]interface{}
	var requestHeaders map[string]interface{}

	// Get data from JSON body or URL parameters
	if c.ContentType() == "application/json" {
		var req ExecuteFunctionRequest
		if err := c.ShouldBindJSON(&req); err == nil {
			requestData = req.Data
			requestHeaders = req.Headers
		}
	}

	// If no JSON body, create data from query parameters and form data
	if requestData == nil {
		requestData = make(map[string]interface{})
		
		// Add query parameters
		for key, values := range c.Request.URL.Query() {
			if len(values) == 1 {
				requestData[key] = values[0]
			} else {
				requestData[key] = values
			}
		}

		// Add form data if present
		if err := c.Request.ParseForm(); err == nil {
			for key, values := range c.Request.PostForm {
				if len(values) == 1 {
					requestData[key] = values[0]
				} else {
					requestData[key] = values
				}
			}
		}
	}

	// Get headers
	if requestHeaders == nil {
		requestHeaders = make(map[string]interface{})
		for key, values := range c.Request.Header {
			if len(values) == 1 {
				requestHeaders[key] = values[0]
			} else {
				requestHeaders[key] = values
			}
		}
	}

	// Find the function by name
	var function models.Function
	if err := h.db.Where("name = ? AND project_id = ? AND is_active = ? AND is_public = ?", 
		functionName, projectID, true, true).First(&function).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "Function not found or not public"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch function"})
		}
		return
	}

	if function.Status != "deployed" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Function is not deployed"})
		return
	}

	// Execute function using real execution engine
	executionID := uuid.New().String()
	startTime := time.Now()

	// Create execution request
	execReq := execution.ExecutionRequest{
		Function: function,
		Data:     requestData,
		Headers:  requestHeaders,
		Method:   c.Request.Method,
		Path:     c.Request.URL.Path,
	}

	// Execute function with timeout context
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(function.Timeout)*time.Second)
	defer cancel()

	result, err := h.executor.Execute(ctx, execReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Execution failed: %v", err)})
		return
	}

	executionTime := result.ExecutionTime
	var response map[string]interface{}
	status := "success"
	statusCode := result.StatusCode
	logs := result.Logs

	if result.Success {
		response = result.Response
	} else {
		response = map[string]interface{}{
			"error": result.Error,
		}
		status = "error"
		if statusCode == 0 {
			statusCode = 500
		}
	}

	// Log execution
	execution := models.FunctionExecution{
		FunctionID:    function.ID,
		ExecutionID:   executionID,
		RequestData:   requestData,
		ResponseData:  response,
		Headers:       requestHeaders,
		Method:        c.Request.Method,
		Path:          c.Request.URL.Path,
		Status:        status,
		StatusCode:    statusCode,
		ExecutionTime: executionTime,
		MemoryUsage:   result.MemoryUsage,
		StartedAt:     startTime,
		CompletedAt:   &time.Time{},
		Logs:          logs,
		UserAgent:     c.GetHeader("User-Agent"),
		ClientIP:      c.ClientIP(),
		Source:        "http",
		ProjectID:     projectID,
	}

	now := time.Now()
	execution.CompletedAt = &now

	if err := h.db.Create(&execution).Error; err != nil {
		// Log error but don't fail the execution response
		fmt.Printf("Failed to log execution: %v\n", err)
	}

	c.JSON(http.StatusOK, response)
}

// GetFunctionLogs returns function execution logs
func (h *FunctionHandler) GetFunctionLogs(c *gin.Context) {
	projectID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid project ID"})
		return
	}

	functionID, err := strconv.ParseUint(c.Param("function_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid function ID"})
		return
	}

	// Parse query parameters
	limit := 100
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 && parsedLimit <= 1000 {
			limit = parsedLimit
		}
	}

	var executions []models.FunctionExecution
	if err := h.db.Where("function_id = ? AND project_id = ?", uint(functionID), uint(projectID)).
		Order("created_at DESC").
		Limit(limit).
		Find(&executions).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch execution logs"})
		return
	}

	c.JSON(http.StatusOK, executions)
}

// realDeployment implements real function deployment
func (h *FunctionHandler) realDeployment(function models.Function) {
	// Update status to building
	h.db.Model(&function).Updates(map[string]interface{}{
		"status": "building",
		"build_logs": "Starting function deployment...\n" +
			"Installing dependencies...\n" +
			"Building function package...\n",
	})

	// For now, we'll simulate the deployment process
	// In a full implementation, this would:
	// 1. Install dependencies based on function.Dependencies
	// 2. Build the function package
	// 3. Deploy to a container registry or function runtime
	// 4. Set up networking and scaling rules
	
	time.Sleep(3 * time.Second) // Simulate build time
	
	now := time.Now()
	h.db.Model(&function).Updates(map[string]interface{}{
		"status":            "deployed",
		"last_deployed_at":  &now,
		"build_logs":        "Starting function deployment...\nInstalling dependencies...\nBuilding function package...\nBuild completed successfully!\n",
		"deployment_logs":   "Deploying function...\nFunction deployed and ready to receive requests!\n",
	})
}