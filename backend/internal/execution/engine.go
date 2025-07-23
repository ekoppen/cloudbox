package execution

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/models"
	"github.com/google/uuid"
)

// ExecutionEngine handles function execution
type ExecutionEngine struct {
	workDir     string
	timeout     time.Duration
	maxMemory   int64 // in bytes
	enableDocker bool
}

// NewExecutionEngine creates a new execution engine
func NewExecutionEngine(workDir string, timeout time.Duration, maxMemory int64) *ExecutionEngine {
	return &ExecutionEngine{
		workDir:      workDir,
		timeout:      timeout,
		maxMemory:    maxMemory,
		enableDocker: checkDockerAvailable(),
	}
}

// ExecutionRequest represents a function execution request
type ExecutionRequest struct {
	Function models.Function
	Data     map[string]interface{}
	Headers  map[string]interface{}
	Method   string
	Path     string
}

// ExecutionResult represents the result of function execution
type ExecutionResult struct {
	Success       bool                   `json:"success"`
	Response      map[string]interface{} `json:"response,omitempty"`
	Error         string                 `json:"error,omitempty"`
	ExecutionTime int64                  `json:"execution_time"` // milliseconds
	MemoryUsage   int64                  `json:"memory_usage"`   // bytes
	Logs          string                 `json:"logs"`
	StatusCode    int                    `json:"status_code"`
}

// Execute runs a function with the given request
func (e *ExecutionEngine) Execute(ctx context.Context, req ExecutionRequest) (*ExecutionResult, error) {
	startTime := time.Now()
	
	// Create execution context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()
	
	// Create temporary workspace
	executionID := uuid.New().String()
	workspaceDir := filepath.Join(e.workDir, executionID)
	
	if err := os.MkdirAll(workspaceDir, 0755); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create workspace: %v", err),
			StatusCode: 500,
		}, err
	}
	defer os.RemoveAll(workspaceDir) // Cleanup
	
	var result *ExecutionResult
	var err error
	
	if e.enableDocker {
		result, err = e.executeInDocker(execCtx, req, workspaceDir)
	} else {
		result, err = e.executeNative(execCtx, req, workspaceDir)
	}
	
	if result != nil {
		result.ExecutionTime = time.Since(startTime).Milliseconds()
	}
	
	return result, err
}

// executeInDocker runs function in Docker container for isolation
func (e *ExecutionEngine) executeInDocker(ctx context.Context, req ExecutionRequest, workspaceDir string) (*ExecutionResult, error) {
	// Prepare execution environment based on runtime
	dockerImage, err := e.getDockerImage(req.Function.Runtime)
	if err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Unsupported runtime: %s", req.Function.Runtime),
			StatusCode: 400,
		}, err
	}
	
	// Create function file
	if err := e.createFunctionFile(workspaceDir, req.Function); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create function file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Create input data file
	inputFile := filepath.Join(workspaceDir, "input.json")
	inputData := map[string]interface{}{
		"data":    req.Data,
		"headers": req.Headers,
		"method":  req.Method,
		"path":    req.Path,
	}
	
	inputBytes, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputBytes, 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create input file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Prepare Docker command
	cmd := exec.CommandContext(ctx, "docker", "run", "--rm",
		"-v", fmt.Sprintf("%s:/workspace", workspaceDir),
		"-w", "/workspace",
		"--memory", fmt.Sprintf("%dm", e.maxMemory/(1024*1024)), // Convert to MB
		"--network", "none", // No network access for security
		"--user", "1000:1000", // Non-root user
		dockerImage,
	)
	
	// Add runtime-specific execution command
	execCmd := e.getRuntimeCommand(req.Function)
	cmd.Args = append(cmd.Args, execCmd...)
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	// Execute
	err = cmd.Run()
	
	// Parse results
	result := &ExecutionResult{
		Logs: stdout.String() + stderr.String(),
	}
	
	if err != nil {
		result.Success = false
		result.Error = fmt.Sprintf("Execution failed: %v", err)
		result.StatusCode = 500
		return result, nil
	}
	
	// Try to parse JSON output
	outputFile := filepath.Join(workspaceDir, "output.json")
	if outputBytes, err := os.ReadFile(outputFile); err == nil {
		var output map[string]interface{}
		if json.Unmarshal(outputBytes, &output) == nil {
			result.Success = true
			result.Response = output
			result.StatusCode = 200
		}
	}
	
	if !result.Success {
		// Fallback to stdout as response
		result.Success = true
		result.Response = map[string]interface{}{
			"output": stdout.String(),
		}
		result.StatusCode = 200
	}
	
	return result, nil
}

// executeNative runs function directly on host (fallback)
func (e *ExecutionEngine) executeNative(ctx context.Context, req ExecutionRequest, workspaceDir string) (*ExecutionResult, error) {
	// This is a simplified native execution
	// In production, you'd want better isolation
	
	switch req.Function.Language {
	case "javascript":
		return e.executeJavaScript(ctx, req, workspaceDir)
	case "python":
		return e.executePython(ctx, req, workspaceDir)
	case "go":
		return e.executeGo(ctx, req, workspaceDir)
	default:
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Unsupported language: %s", req.Function.Language),
			StatusCode: 400,
		}, nil
	}
}

// executeJavaScript runs JavaScript function using Node.js
func (e *ExecutionEngine) executeJavaScript(ctx context.Context, req ExecutionRequest, workspaceDir string) (*ExecutionResult, error) {
	// Create JavaScript file
	jsFile := filepath.Join(workspaceDir, "function.js")
	
	// Wrapper code to handle CloudBox function execution
	wrapper := fmt.Sprintf(`
const fs = require('fs');

// User function code
%s

// CloudBox execution wrapper
async function executeFunction() {
    try {
        const input = JSON.parse(fs.readFileSync('input.json', 'utf8'));
        
        // Call the user's function
        let result;
        if (typeof %s === 'function') {
            result = await %s(input.data, input);
        } else if (typeof handler === 'function') {
            result = await handler(input.data, input);
        } else {
            throw new Error('No handler function found');
        }
        
        // Write result
        fs.writeFileSync('output.json', JSON.stringify({
            success: true,
            data: result
        }));
        
        console.log('Function executed successfully');
    } catch (error) {
        fs.writeFileSync('output.json', JSON.stringify({
            success: false,
            error: error.message
        }));
        console.error('Function execution failed:', error.message);
        process.exit(1);
    }
}

executeFunction();
`, req.Function.Code, req.Function.EntryPoint, req.Function.EntryPoint)
	
	if err := os.WriteFile(jsFile, []byte(wrapper), 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create JavaScript file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Create input file
	inputFile := filepath.Join(workspaceDir, "input.json")
	inputData := map[string]interface{}{
		"data":    req.Data,
		"headers": req.Headers,
		"method":  req.Method,
		"path":    req.Path,
	}
	
	inputBytes, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputBytes, 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create input file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Execute with Node.js
	cmd := exec.CommandContext(ctx, "node", jsFile)
	cmd.Dir = workspaceDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	result := &ExecutionResult{
		Logs: stdout.String() + stderr.String(),
	}
	
	// Parse output
	outputFile := filepath.Join(workspaceDir, "output.json")
	if outputBytes, err := os.ReadFile(outputFile); err == nil {
		var output map[string]interface{}
		if json.Unmarshal(outputBytes, &output) == nil {
			if success, ok := output["success"].(bool); ok && success {
				result.Success = true
				result.Response = output["data"].(map[string]interface{})
				result.StatusCode = 200
			} else {
				result.Success = false
				result.Error = output["error"].(string)
				result.StatusCode = 500
			}
		}
	}
	
	if !result.Success && result.Error == "" {
		result.Error = "Function execution failed"
		result.StatusCode = 500
	}
	
	return result, nil
}

// executePython runs Python function
func (e *ExecutionEngine) executePython(ctx context.Context, req ExecutionRequest, workspaceDir string) (*ExecutionResult, error) {
	// Create Python file
	pyFile := filepath.Join(workspaceDir, "function.py")
	
	wrapper := fmt.Sprintf(`
import json
import sys
import traceback

# User function code
%s

def execute_function():
    try:
        with open('input.json', 'r') as f:
            input_data = json.load(f)
        
        # Call the user's function
        if '%s' in globals() and callable(globals()['%s']):
            result = globals()['%s'](input_data['data'], input_data)
        elif 'handler' in globals() and callable(globals()['handler']):
            result = globals()['handler'](input_data['data'], input_data)
        else:
            raise Exception('No handler function found')
        
        # Write result
        with open('output.json', 'w') as f:
            json.dump({
                'success': True,
                'data': result
            }, f)
        
        print('Function executed successfully')
    except Exception as e:
        with open('output.json', 'w') as f:
            json.dump({
                'success': False,
                'error': str(e)
            }, f)
        print(f'Function execution failed: {str(e)}')
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    execute_function()
`, req.Function.Code, req.Function.EntryPoint, req.Function.EntryPoint, req.Function.EntryPoint)
	
	if err := os.WriteFile(pyFile, []byte(wrapper), 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create Python file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Create input file
	inputFile := filepath.Join(workspaceDir, "input.json")
	inputData := map[string]interface{}{
		"data":    req.Data,
		"headers": req.Headers,
		"method":  req.Method,
		"path":    req.Path,
	}
	
	inputBytes, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputBytes, 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create input file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Execute with Python
	cmd := exec.CommandContext(ctx, "python3", pyFile)
	cmd.Dir = workspaceDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	result := &ExecutionResult{
		Logs: stdout.String() + stderr.String(),
	}
	
	// Parse output
	outputFile := filepath.Join(workspaceDir, "output.json")
	if outputBytes, err := os.ReadFile(outputFile); err == nil {
		var output map[string]interface{}
		if json.Unmarshal(outputBytes, &output) == nil {
			if success, ok := output["success"].(bool); ok && success {
				result.Success = true
				result.Response = output["data"].(map[string]interface{})
				result.StatusCode = 200
			} else {
				result.Success = false
				result.Error = output["error"].(string)
				result.StatusCode = 500
			}
		}
	}
	
	if !result.Success && result.Error == "" {
		result.Error = "Function execution failed"
		result.StatusCode = 500
	}
	
	return result, nil
}

// executeGo runs Go function
func (e *ExecutionEngine) executeGo(ctx context.Context, req ExecutionRequest, workspaceDir string) (*ExecutionResult, error) {
	// For Go, we need to compile and then execute
	goFile := filepath.Join(workspaceDir, "main.go")
	
	wrapper := fmt.Sprintf(`
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

%s

func main() {
	inputFile, err := os.Open("input.json")
	if err != nil {
		fmt.Printf("Failed to open input file: %%v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	var inputData map[string]interface{}
	if err := json.NewDecoder(inputFile).Decode(&inputData); err != nil {
		fmt.Printf("Failed to decode input: %%v\n", err)
		os.Exit(1)
	}

	// Call handler function (simplified)
	result := map[string]interface{}{
		"message": "Hello from Go function!",
		"input":   inputData,
	}

	outputData := map[string]interface{}{
		"success": true,
		"data":    result,
	}

	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Printf("Failed to create output file: %%v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	if err := json.NewEncoder(outputFile).Encode(outputData); err != nil {
		fmt.Printf("Failed to encode output: %%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Function executed successfully")
}
`, req.Function.Code)
	
	if err := os.WriteFile(goFile, []byte(wrapper), 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create Go file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Create input file
	inputFile := filepath.Join(workspaceDir, "input.json")
	inputData := map[string]interface{}{
		"data":    req.Data,
		"headers": req.Headers,
		"method":  req.Method,
		"path":    req.Path,
	}
	
	inputBytes, _ := json.Marshal(inputData)
	if err := os.WriteFile(inputFile, inputBytes, 0644); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Failed to create input file: %v", err),
			StatusCode: 500,
		}, err
	}
	
	// Compile Go program
	binaryPath := filepath.Join(workspaceDir, "function")
	compileCmd := exec.CommandContext(ctx, "go", "build", "-o", binaryPath, goFile)
	compileCmd.Dir = workspaceDir
	
	var compileStderr bytes.Buffer
	compileCmd.Stderr = &compileStderr
	
	if err := compileCmd.Run(); err != nil {
		return &ExecutionResult{
			Success:    false,
			Error:      fmt.Sprintf("Go compilation failed: %v - %s", err, compileStderr.String()),
			StatusCode: 500,
			Logs:       compileStderr.String(),
		}, err
	}
	
	// Execute compiled binary
	cmd := exec.CommandContext(ctx, binaryPath)
	cmd.Dir = workspaceDir
	
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	
	err := cmd.Run()
	
	result := &ExecutionResult{
		Logs: stdout.String() + stderr.String(),
	}
	
	// Parse output
	outputFile := filepath.Join(workspaceDir, "output.json")
	if outputBytes, err := os.ReadFile(outputFile); err == nil {
		var output map[string]interface{}
		if json.Unmarshal(outputBytes, &output) == nil {
			if success, ok := output["success"].(bool); ok && success {
				result.Success = true
				result.Response = output["data"].(map[string]interface{})
				result.StatusCode = 200
			} else {
				result.Success = false
				result.Error = output["error"].(string)
				result.StatusCode = 500
			}
		}
	}
	
	if !result.Success && result.Error == "" {
		result.Error = "Function execution failed"
		result.StatusCode = 500
	}
	
	return result, nil
}

// Helper functions

func checkDockerAvailable() bool {
	cmd := exec.Command("docker", "--version")
	return cmd.Run() == nil
}

func (e *ExecutionEngine) getDockerImage(runtime string) (string, error) {
	switch runtime {
	case "nodejs18", "nodejs16", "nodejs14":
		return "node:18-alpine", nil
	case "python3.9", "python3.8", "python3.7":
		return "python:3.9-alpine", nil
	case "go1.19", "go1.18":
		return "golang:1.19-alpine", nil
	default:
		return "", fmt.Errorf("unsupported runtime: %s", runtime)
	}
}

func (e *ExecutionEngine) getRuntimeCommand(function models.Function) []string {
	switch function.Language {
	case "javascript":
		return []string{"node", "function.js"}
	case "python":
		return []string{"python3", "function.py"}
	case "go":
		return []string{"sh", "-c", "go build -o function main.go && ./function"}
	default:
		return []string{"echo", "Unsupported runtime"}
	}
}

func (e *ExecutionEngine) createFunctionFile(workspaceDir string, function models.Function) error {
	var filename, content string
	
	switch function.Language {
	case "javascript":
		filename = "function.js"
		content = fmt.Sprintf(`
const fs = require('fs');

// User function code
%s

// CloudBox execution wrapper
async function executeFunction() {
    try {
        const input = JSON.parse(fs.readFileSync('input.json', 'utf8'));
        
        let result;
        if (typeof %s === 'function') {
            result = await %s(input.data, input);
        } else if (typeof handler === 'function') {
            result = await handler(input.data, input);
        } else {
            throw new Error('No handler function found');
        }
        
        fs.writeFileSync('output.json', JSON.stringify({
            success: true,
            data: result
        }));
        
        console.log('Function executed successfully');
    } catch (error) {
        fs.writeFileSync('output.json', JSON.stringify({
            success: false,
            error: error.message
        }));
        console.error('Function execution failed:', error.message);
        process.exit(1);
    }
}

executeFunction();
`, function.Code, function.EntryPoint, function.EntryPoint)
		
	case "python":
		filename = "function.py"
		content = fmt.Sprintf(`
import json
import sys
import traceback

# User function code
%s

def execute_function():
    try:
        with open('input.json', 'r') as f:
            input_data = json.load(f)
        
        if '%s' in globals() and callable(globals()['%s']):
            result = globals()['%s'](input_data['data'], input_data)
        elif 'handler' in globals() and callable(globals()['handler']):
            result = globals()['handler'](input_data['data'], input_data)
        else:
            raise Exception('No handler function found')
        
        with open('output.json', 'w') as f:
            json.dump({
                'success': True,
                'data': result
            }, f)
        
        print('Function executed successfully')
    except Exception as e:
        with open('output.json', 'w') as f:
            json.dump({
                'success': False,
                'error': str(e)
            }, f)
        print(f'Function execution failed: {str(e)}')
        traceback.print_exc()
        sys.exit(1)

if __name__ == '__main__':
    execute_function()
`, function.Code, function.EntryPoint, function.EntryPoint, function.EntryPoint)
		
	case "go":
		filename = "main.go"
		content = fmt.Sprintf(`
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

%s

func main() {
	inputFile, err := os.Open("input.json")
	if err != nil {
		fmt.Printf("Failed to open input file: %%v\n", err)
		os.Exit(1)
	}
	defer inputFile.Close()

	var inputData map[string]interface{}
	if err := json.NewDecoder(inputFile).Decode(&inputData); err != nil {
		fmt.Printf("Failed to decode input: %%v\n", err)
		os.Exit(1)
	}

	result := map[string]interface{}{
		"message": "Hello from Go function!",
		"input":   inputData,
	}

	outputData := map[string]interface{}{
		"success": true,
		"data":    result,
	}

	outputFile, err := os.Create("output.json")
	if err != nil {
		fmt.Printf("Failed to create output file: %%v\n", err)
		os.Exit(1)
	}
	defer outputFile.Close()

	if err := json.NewEncoder(outputFile).Encode(outputData); err != nil {
		fmt.Printf("Failed to encode output: %%v\n", err)
		os.Exit(1)
	}

	fmt.Println("Function executed successfully")
}
`, function.Code)
	}
	
	filePath := filepath.Join(workspaceDir, filename)
	return os.WriteFile(filePath, []byte(content), 0644)
}