package services

import (
	"bufio"
	"context"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"
	"github.com/cloudbox/backend/internal/models"
)

// RemoteTerminalService handles SSH-based remote script execution with interactive support
type RemoteTerminalService struct {
	client *ssh.Client
	mutex  sync.RWMutex
}

// TerminalSession represents an active terminal session
type TerminalSession struct {
	ID              string
	SSH             *ssh.Client
	Session         *ssh.Session
	StdinPipe       io.WriteCloser
	StdoutReader    io.Reader
	StderrReader    io.Reader
	Context         context.Context
	Cancel          context.CancelFunc
	Environment     map[string]string
	DeploymentID    uint
	OutputCallback  func(string, string) // output, logType (stdout/stderr/info/error)
	PromptCallback  func(string) string  // Handle interactive prompts
}

// NewRemoteTerminalService creates a new remote terminal service
func NewRemoteTerminalService() *RemoteTerminalService {
	return &RemoteTerminalService{}
}

// getDeploymentPath calculates the deployment path for a deployment
func (rts *RemoteTerminalService) getDeploymentPath(deployment models.Deployment, webServer models.WebServer) string {
	// Priority: 1. Deployment's DeployPath, 2. WebServer's DeployPath, 3. Default ~/deploys
	if deployment.DeployPath != "" {
		return deployment.DeployPath
	}
	if webServer.DeployPath != "" {
		return fmt.Sprintf("%s/%s", webServer.DeployPath, deployment.Name)
	}
	// Use ~/deploys as default (user's home directory)
	return fmt.Sprintf("/home/%s/deploys/%s", webServer.Username, deployment.Name)
}

// CreateSession establishes SSH connection and creates interactive session
func (rts *RemoteTerminalService) CreateSession(webServer models.WebServer, deployment models.Deployment) (*TerminalSession, error) {
	// Create SSH client configuration
	config := &ssh.ClientConfig{
		User:            webServer.Username,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	// Debug: Always log SSH key info
	log.Printf("[DEBUG TERMINAL] WebServer %s: SSH Key ID=%d, SSH Key loaded: ID=%d, Name=%s, PrivateKey length=%d", 
		webServer.Name, webServer.SSHKeyID, webServer.SSHKey.ID, webServer.SSHKey.Name, len(webServer.SSHKey.PrivateKey))
	
	// Add SSH key authentication
	if webServer.SSHKey.PrivateKey != "" {
		// Parse private key from SSH key
		privateKey, err := rts.parsePrivateKey(webServer.SSHKey.PrivateKey)
		if err != nil {
			return nil, fmt.Errorf("failed to parse SSH private key for server %s (key: %s): %w", 
				webServer.Name, webServer.SSHKey.Name, err)
		}
		config.Auth = append(config.Auth, ssh.PublicKeys(privateKey))
	} else {
		return nil, fmt.Errorf("no SSH private key data for server %s (SSH Key ID: %d, Key Name: %s)", 
			webServer.Name, webServer.SSHKeyID, webServer.SSHKey.Name)
	}

	// Establish SSH connection
	client, err := ssh.Dial("tcp", fmt.Sprintf("%s:%d", webServer.Hostname, webServer.Port), config)
	if err != nil {
		return nil, fmt.Errorf("failed to establish SSH connection: %w", err)
	}

	// Create SSH session
	session, err := client.NewSession()
	if err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create SSH session: %w", err)
	}

	// Set up pseudo-terminal for interactive support
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,     // Enable echoing
		ssh.TTY_OP_ISPEED: 14400, // Input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // Output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 24, modes); err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to request pseudo terminal: %w", err)
	}

	// Set up pipes
	stdinPipe, err := session.StdinPipe()
	if err != nil {
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdoutReader, err := session.StdoutPipe()
	if err != nil {
		stdinPipe.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderrReader, err := session.StderrPipe()
	if err != nil {
		stdinPipe.Close()
		session.Close()
		client.Close()
		return nil, fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	// Create context for session management
	ctx, cancel := context.WithCancel(context.Background())

	// Generate CloudBox environment variables with calculated deployment path
	deploymentPath := rts.getDeploymentPath(deployment, webServer)
	environment := rts.generateCloudBoxEnvironment(deployment, webServer, deploymentPath)

	return &TerminalSession{
		ID:           fmt.Sprintf("term_%d_%d", deployment.ID, time.Now().Unix()),
		SSH:          client,
		Session:      session,
		StdinPipe:    stdinPipe,
		StdoutReader: stdoutReader,
		StderrReader: stderrReader,
		Context:      ctx,
		Cancel:       cancel,
		Environment:  environment,
		DeploymentID: deployment.ID,
	}, nil
}

// ExecuteCIPScript executes a CloudBox Install Protocol script with full environment injection
func (rts *RemoteTerminalService) ExecuteCIPScript(session *TerminalSession, scriptType string, appPath string) error {
	// Validate CIP compliance first
	if err := rts.validateCIPCompliance(session, appPath); err != nil {
		return fmt.Errorf("CIP validation failed: %w", err)
	}

	// Get script path from cloudbox.json
	scriptPath, err := rts.getCIPScriptPath(session, appPath, scriptType)
	if err != nil {
		return fmt.Errorf("failed to get script path: %w", err)
	}

	// Inject CloudBox environment
	if err := rts.injectEnvironment(session, appPath); err != nil {
		return fmt.Errorf("failed to inject environment: %w", err)
	}

	// Start output monitoring
	go rts.monitorOutput(session)

	// Execute the script
	command := fmt.Sprintf("cd %s && chmod +x %s && %s", appPath, scriptPath, scriptPath)
	
	session.OutputCallback("🚀 [CIP] Executing CloudBox Install Protocol script", "info")
	session.OutputCallback(fmt.Sprintf("📁 Path: %s", appPath), "info")
	session.OutputCallback(fmt.Sprintf("📜 Script: %s", scriptPath), "info")
	session.OutputCallback(fmt.Sprintf("⚡ Command: %s", command), "info")

	if err := session.Session.Start(command); err != nil {
		return fmt.Errorf("failed to start script execution: %w", err)
	}

	// Wait for completion or cancellation
	done := make(chan error, 1)
	go func() {
		done <- session.Session.Wait()
	}()

	select {
	case err := <-done:
		if err != nil {
			session.OutputCallback(fmt.Sprintf("❌ [CIP] Script execution failed: %v", err), "error")
			return fmt.Errorf("script execution failed: %w", err)
		}
		session.OutputCallback("✅ [CIP] Script execution completed successfully", "info")
		return nil
	case <-session.Context.Done():
		session.OutputCallback("⏹️  [CIP] Script execution cancelled", "info")
		return fmt.Errorf("script execution cancelled")
	}
}

// validateCIPCompliance checks if the app follows CloudBox Install Protocol
func (rts *RemoteTerminalService) validateCIPCompliance(session *TerminalSession, appPath string) error {
	session.OutputCallback("🔍 [CIP] Validating CloudBox Install Protocol compliance", "info")

	// Check for cloudbox.json
	checkCmd := fmt.Sprintf("cd %s && test -f cloudbox.json", appPath)
	if err := rts.runSimpleCommand(session, checkCmd); err != nil {
		return fmt.Errorf("cloudbox.json not found - app is not CIP compliant")
	}

	// Validate required scripts exist
	validateCmd := fmt.Sprintf(`cd %s && 
		echo "🔍 Current directory: $(pwd)"
		echo "📁 Directory contents:"
		ls -la
		echo ""
		
		# Try to install jq if not available, but don't fail if we can't
		if ! command -v jq >/dev/null 2>&1; then
			echo "🔧 jq not found, trying to install..."
			if command -v apt-get >/dev/null 2>&1; then
				sudo apt-get update && sudo apt-get install -y jq 2>/dev/null || echo "⚠️  Could not install jq via apt-get"
			elif command -v yum >/dev/null 2>&1; then
				sudo yum install -y jq 2>/dev/null || echo "⚠️  Could not install jq via yum"
			elif command -v apk >/dev/null 2>&1; then
				sudo apk add --no-cache jq 2>/dev/null || echo "⚠️  Could not install jq via apk"
			else
				echo "⚠️  Package manager not found, will use alternative JSON parsing"
			fi
		fi
		
		echo "📝 Parsing cloudbox.json..."
		if command -v jq >/dev/null 2>&1; then
			echo "✅ Using jq for JSON parsing"
			cat cloudbox.json | jq '.scripts' || (echo "Failed to parse cloudbox.json with jq" && exit 1)
		else
			echo "⚠️  Using alternative JSON parsing (basic grep/sed)"
			echo "Scripts section from cloudbox.json:"
			grep -A 10 '"scripts"' cloudbox.json | head -15
		fi
		echo ""
		
		for script in install start stop status; do 
			echo "🔍 Checking script: $script"
			
			# Try jq first, fall back to grep/sed if jq is not available
			if command -v jq >/dev/null 2>&1; then
				script_path=$(jq -r ".scripts.$script // empty" cloudbox.json 2>/dev/null)
			else
				# Alternative parsing using grep and sed
				script_path=$(grep -A 20 '"scripts"' cloudbox.json | grep "\"$script\"" | sed 's/.*: *"\([^"]*\)".*/\1/' | head -1)
			fi
			
			if [[ -z "$script_path" || "$script_path" == "null" ]]; then
				echo "❌ Required script '$script' not defined in cloudbox.json"
				exit 1
			fi
			echo "   Script path: $script_path"
			if [[ ! -f "$script_path" ]]; then
				echo "❌ Required script file '$script_path' not found"
				echo "   Directory listing where script should be:"
				ls -la $(dirname "$script_path") 2>/dev/null || echo "   Directory $(dirname "$script_path") does not exist"
				exit 1
			fi
			if [[ ! -x "$script_path" ]]; then
				echo "⚠️  Making script executable: $script_path"
				chmod +x "$script_path"
			fi
			echo "✅ Script '$script' found and ready at $script_path"
		done`, appPath)

	// Use command with output instead of simple command to see debug info
	output, err := rts.runCommandWithOutput(session, validateCmd)
	// Always show the output, regardless of error status
	if output != "" {
		session.OutputCallback(fmt.Sprintf("🔍 [CIP] Validation debug output:\n%s", output), "info")
	}
	if err != nil {
		session.OutputCallback(fmt.Sprintf("❌ [CIP] Validation failed with error: %s", err.Error()), "error")
		return fmt.Errorf("CIP script validation failed: %w", err)
	}
	session.OutputCallback("✅ [CIP] All validation checks passed", "info")

	session.OutputCallback("✅ [CIP] CloudBox Install Protocol validation passed", "info")
	return nil
}

// getCIPScriptPath extracts script path from cloudbox.json
func (rts *RemoteTerminalService) getCIPScriptPath(session *TerminalSession, appPath, scriptType string) (string, error) {
	// Try jq first, fall back to grep/sed if not available
	command := fmt.Sprintf(`cd %s && 
		if command -v jq >/dev/null 2>&1; then
			jq -r ".scripts.%s" cloudbox.json
		else
			# Alternative parsing using grep and sed
			grep -A 20 '"scripts"' cloudbox.json | grep "\"%s\"" | sed 's/.*: *"\([^"]*\)".*/\1/' | head -1
		fi`, appPath, scriptType, scriptType)
	
	output, err := rts.runCommandWithOutput(session, command)
	if err != nil {
		return "", fmt.Errorf("failed to get script path for %s: %w", scriptType, err)
	}

	scriptPath := strings.TrimSpace(output)
	if scriptPath == "" || scriptPath == "null" {
		return "", fmt.Errorf("script '%s' not defined in cloudbox.json", scriptType)
	}

	return scriptPath, nil
}

// injectEnvironment creates .env files and exports CloudBox environment variables
func (rts *RemoteTerminalService) injectEnvironment(session *TerminalSession, appPath string) error {
	session.OutputCallback("🔧 [CIP] Injecting CloudBox environment variables", "info")

	// Create environment script
	envScript := ""
	for key, value := range session.Environment {
		envScript += fmt.Sprintf("export %s=\"%s\"\n", key, value)
		session.OutputCallback(fmt.Sprintf("📋 %s=%s", key, value), "info")
	}

	// Write environment to temporary script
	writeEnvCmd := fmt.Sprintf(`cd %s && cat > .cloudbox-env.sh << 'EOF'
#!/bin/bash
# CloudBox Install Protocol Environment Variables
# Auto-generated by CloudBox v1.0

%s
EOF`, appPath, envScript)

	if err := rts.runSimpleCommand(session, writeEnvCmd); err != nil {
		return fmt.Errorf("failed to write environment script: %w", err)
	}

	// Source environment in current shell
	sourceCmd := fmt.Sprintf("cd %s && chmod +x .cloudbox-env.sh && source .cloudbox-env.sh", appPath)
	if err := rts.runSimpleCommand(session, sourceCmd); err != nil {
		return fmt.Errorf("failed to source environment: %w", err)
	}

	session.OutputCallback("✅ [CIP] Environment injection completed", "info")
	return nil
}

// generateCloudBoxEnvironment creates environment variables for CIP scripts
func (rts *RemoteTerminalService) generateCloudBoxEnvironment(deployment models.Deployment, webServer models.WebServer, deploymentPath string) map[string]string {
	env := map[string]string{
		"CLOUDBOX_API_URL":         fmt.Sprintf("https://cloudbox.domain/api/projects/%d", deployment.ProjectID),
		"CLOUDBOX_PROJECT_ID":      fmt.Sprintf("%d", deployment.ProjectID),
		"CLOUDBOX_PROJECT_SLUG":    deployment.Name,
		"CLOUDBOX_DEPLOYMENT_ID":   fmt.Sprintf("%d", deployment.ID),
		"CLOUDBOX_DEPLOYMENT_PATH": deploymentPath,
		"CLOUDBOX_ENVIRONMENT":     "production",
		"CLOUDBOX_VERSION":         "1.0",
		"CLOUDBOX_DOCKER_ENABLED":  "true",
	}

	// Add port configuration
	if deployment.PortConfiguration != nil {
		for portName, portConfig := range deployment.PortConfiguration {
			envVarName := fmt.Sprintf("CLOUDBOX_%s_PORT", strings.ToUpper(portName))
			env[envVarName] = fmt.Sprintf("%d", portConfig)
		}
		
		// Set main web port if available
		if webPort, exists := deployment.PortConfiguration["web"]; exists {
			env["CLOUDBOX_WEB_PORT"] = fmt.Sprintf("%d", webPort)
		}
	}

	return env
}

// monitorOutput handles real-time output streaming from SSH session
func (rts *RemoteTerminalService) monitorOutput(session *TerminalSession) {
	if session.OutputCallback == nil {
		return
	}

	// Monitor stdout
	go func() {
		scanner := bufio.NewScanner(session.StdoutReader)
		for scanner.Scan() {
			line := scanner.Text()
			session.OutputCallback(line, "stdout")
		}
	}()

	// Monitor stderr
	go func() {
		scanner := bufio.NewScanner(session.StderrReader)
		for scanner.Scan() {
			line := scanner.Text()
			session.OutputCallback(line, "stderr")
		}
	}()
}

// runSimpleCommand executes a command and waits for completion
func (rts *RemoteTerminalService) runSimpleCommand(session *TerminalSession, command string) error {
	tempSession, err := session.SSH.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create temporary session: %w", err)
	}
	defer tempSession.Close()

	return tempSession.Run(command)
}

// runCommandWithOutput executes a command and returns its output (both stdout and stderr)
func (rts *RemoteTerminalService) runCommandWithOutput(session *TerminalSession, command string) (string, error) {
	tempSession, err := session.SSH.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create temporary session: %w", err)
	}
	defer tempSession.Close()

	// Capture both stdout and stderr
	output, err := tempSession.CombinedOutput(command)
	if err != nil {
		// Return both the error and any output that was generated
		return string(output), fmt.Errorf("command execution failed: %w", err)
	}

	return string(output), nil
}

// HandleInteractivePrompt processes interactive prompts during script execution
func (rts *RemoteTerminalService) HandleInteractivePrompt(session *TerminalSession, prompt string) {
	if session.PromptCallback != nil {
		response := session.PromptCallback(prompt)
		if response != "" {
			session.StdinPipe.Write([]byte(response + "\n"))
		}
	}
}

// CloseSession properly closes SSH session and connections
func (rts *RemoteTerminalService) CloseSession(session *TerminalSession) error {
	session.OutputCallback("🔌 [CIP] Closing terminal session", "info")
	
	session.Cancel()
	
	if session.StdinPipe != nil {
		session.StdinPipe.Close()
	}
	
	if session.Session != nil {
		session.Session.Close()
	}
	
	if session.SSH != nil {
		session.SSH.Close()
	}

	return nil
}

// GetCIPManifest reads and parses cloudbox.json from remote server
func (rts *RemoteTerminalService) GetCIPManifest(session *TerminalSession, appPath string) (map[string]interface{}, error) {
	command := fmt.Sprintf("cd %s && cat cloudbox.json", appPath)
	
	_, err := rts.runCommandWithOutput(session, command)
	if err != nil {
		return nil, fmt.Errorf("failed to read cloudbox.json: %w", err)
	}

	// Parse JSON (simplified - in production use proper JSON parsing)
	manifest := make(map[string]interface{})
	// TODO: Add proper JSON parsing
	
	return manifest, nil
}

// parsePrivateKey parses SSH private key from various formats
func (rts *RemoteTerminalService) parsePrivateKey(privateKeyData string) (ssh.Signer, error) {
	// Try to parse as PEM first
	block, _ := pem.Decode([]byte(privateKeyData))
	if block != nil {
		switch block.Type {
		case "RSA PRIVATE KEY":
			return ssh.ParsePrivateKey([]byte(privateKeyData))
		case "OPENSSH PRIVATE KEY":
			return ssh.ParsePrivateKey([]byte(privateKeyData))
		case "EC PRIVATE KEY":
			return ssh.ParsePrivateKey([]byte(privateKeyData))
		default:
			return ssh.ParsePrivateKey([]byte(privateKeyData))
		}
	}

	// If not PEM format, try to parse directly
	return ssh.ParsePrivateKey([]byte(privateKeyData))
}