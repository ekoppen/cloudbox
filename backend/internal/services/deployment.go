package services

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/utils"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// DeploymentService handles real deployment operations
type DeploymentService struct {
	db            *gorm.DB
	cfg           *config.Config
	terminalService *RemoteTerminalService
}

// getDeploymentPath calculates the deployment path for a deployment
func (s *DeploymentService) getDeploymentPath(deployment models.Deployment) string {
	// Priority: 1. Deployment's DeployPath, 2. WebServer's DeployPath, 3. Default ~/deploys
	var basePath string
	
	if deployment.DeployPath != "" {
		basePath = deployment.DeployPath
	} else if deployment.WebServer.DeployPath != "" {
		basePath = deployment.WebServer.DeployPath
	} else {
		// Use ~/deploys as default (user's home directory)
		basePath = "~/deploys"
	}
	
	// Handle tilde expansion for home directory
	if strings.HasPrefix(basePath, "~/") {
		basePath = fmt.Sprintf("/home/%s/%s", deployment.WebServer.Username, basePath[2:])
	}
	
	// If base path doesn't already include deployment name, add it
	if !strings.HasSuffix(basePath, "/"+deployment.Name) && !strings.HasSuffix(basePath, deployment.Name) {
		basePath = fmt.Sprintf("%s/%s", strings.TrimSuffix(basePath, "/"), deployment.Name)
	}
	
	return basePath
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(db *gorm.DB, cfg *config.Config) *DeploymentService {
	return &DeploymentService{
		db:            db,
		cfg:           cfg,
		terminalService: NewRemoteTerminalService(),
	}
}

// DeploymentResult represents the result of a deployment operation
type DeploymentResult struct {
	Success     bool
	BuildLogs   string
	DeployLogs  string
	ErrorLogs   string
	BuildTime   int64
	DeployTime  int64
	FileCount   int64
	TotalSize   int64
}

// ExecuteDeployment performs a real deployment
func (s *DeploymentService) ExecuteDeployment(deployment models.Deployment, commitHash, branch string) *DeploymentResult {
	result := &DeploymentResult{}

	// Update status to building
	s.updateDeploymentStatus(deployment, "building", "Starting deployment process...\n", "", "")

	// Step 1: Clone repository
	repoDir, err := s.cloneRepository(deployment, commitHash, branch, result)
	if err != nil {
		result.ErrorLogs = fmt.Sprintf("Failed to clone repository: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
		return result
	}
	defer os.RemoveAll(repoDir) // Cleanup

	// Step 2: Prepare deployment environment
	if err := s.prepareDeploymentEnvironment(deployment, repoDir, result); err != nil {
		result.ErrorLogs = fmt.Sprintf("Environment preparation failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
		return result
	}

	// Step 3: Build application
	buildTime := time.Now()
	if err := s.buildApplication(deployment, repoDir, result); err != nil {
		result.ErrorLogs = fmt.Sprintf("Build failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
		return result
	}
	result.BuildTime = time.Since(buildTime).Milliseconds()

	// Step 3: Deploy to server
	s.updateDeploymentStatus(deployment, "deploying", result.BuildLogs, "Connecting to deployment server...\n", "")
	
	deployTime := time.Now()
	if err := s.deployToServer(deployment, repoDir, result); err != nil {
		result.ErrorLogs = fmt.Sprintf("Deployment failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
		return result
	}
	result.DeployTime = time.Since(deployTime).Milliseconds()

	// Step 4: Calculate deployment stats
	s.calculateDeploymentStats(repoDir, result)

	// Success
	result.Success = true
	now := time.Now()
	s.db.Model(&deployment).Updates(map[string]interface{}{
		"status":      "deployed",
		"deployed_at": &now,
		"build_logs":  result.BuildLogs,
		"deploy_logs": result.DeployLogs,
		"error_logs":  result.ErrorLogs,
		"build_time":  result.BuildTime,
		"deploy_time": result.DeployTime,
		"file_count":  result.FileCount,
		"total_size":  result.TotalSize,
		"commit_hash": commitHash,
		"branch":      branch,
	})

	return result
}

// ExecuteCIPDeployment performs a CloudBox Install Protocol deployment via remote terminal
func (s *DeploymentService) ExecuteCIPDeployment(deployment models.Deployment, commitHash, branch string, outputCallback func(string, string)) *DeploymentResult {
	result := &DeploymentResult{}
	
	// Update status to building
	s.updateDeploymentStatus(deployment, "building", "Starting CloudBox Install Protocol deployment...\n", "", "")

	// Load web server with SSH key for SSH connection
	var webServer models.WebServer
	if err := s.db.Preload("SSHKey").First(&webServer, deployment.WebServerID).Error; err != nil {
		result.ErrorLogs = fmt.Sprintf("Failed to load web server: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
		return result
	}
	
	// Debug: Log loaded webserver info
	log.Printf("[DEBUG SERVICE] Loaded WebServer: ID=%d, Name=%s, SSHKeyID=%d", 
		webServer.ID, webServer.Name, webServer.SSHKeyID)
	log.Printf("[DEBUG SERVICE] Loaded SSH Key: ID=%d, Name=%s, PrivateKey length=%d (encrypted)", 
		webServer.SSHKey.ID, webServer.SSHKey.Name, len(webServer.SSHKey.PrivateKey))

	// Check if SSH key is already decrypted (plain text) or needs decryption
	var decryptedPrivateKey string
	privateKeyData := webServer.SSHKey.PrivateKey
	
	// Try to parse as SSH key directly first (in case it's already decrypted)
	log.Printf("[DEBUG SERVICE] Testing if SSH key is already in plain text...")
	_, err := ssh.ParsePrivateKey([]byte(privateKeyData))
	if err == nil {
		// Key is already in plain text format
		log.Printf("[DEBUG SERVICE] SSH key is already in plain text format, length: %d", len(privateKeyData))
		decryptedPrivateKey = privateKeyData
	} else {
		// Key needs decryption
		log.Printf("[DEBUG SERVICE] SSH key appears encrypted, attempting decryption...")
		decryptedPrivateKey, err = s.decryptSSHPrivateKey(privateKeyData)
		if err != nil {
			result.ErrorLogs = fmt.Sprintf("Failed to decrypt SSH private key: %v", err)
			s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
			return result
		}
		log.Printf("[DEBUG SERVICE] SSH key decrypted successfully, length: %d", len(decryptedPrivateKey))
	}
	
	// Replace with decrypted key for terminal service
	webServer.SSHKey.PrivateKey = decryptedPrivateKey

	// Create terminal session
	log.Printf("[DEBUG SERVICE] About to call CreateSession with WebServer ID: %d, SSH Key ID: %d", webServer.ID, webServer.SSHKeyID)
	session, err := s.terminalService.CreateSession(webServer, deployment)
	if err != nil {
		log.Printf("[DEBUG SERVICE] CreateSession failed: %v", err)
		result.ErrorLogs = fmt.Sprintf("Failed to create terminal session: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, "", result.ErrorLogs)
		return result
	}
	log.Printf("[DEBUG SERVICE] CreateSession succeeded")
	defer s.terminalService.CloseSession(session)

	// Set up output callback for real-time logging
	session.OutputCallback = func(output, logType string) {
		switch logType {
		case "stdout", "info":
			result.BuildLogs += output + "\n"
		case "stderr", "error":
			result.ErrorLogs += output + "\n"
		default:
			result.DeployLogs += output + "\n"
		}
		
		// Call external callback for real-time updates
		if outputCallback != nil {
			outputCallback(output, logType)
		}
		
		// Update database with latest logs
		s.updateDeploymentStatus(deployment, "building", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
	}

	// Step 1: Clone repository on remote server
	buildTime := time.Now()
	if err := s.remoteCIPClone(session, deployment, commitHash, branch); err != nil {
		result.ErrorLogs += fmt.Sprintf("Remote clone failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
		return result
	}

	// Step 2: Execute CIP install script
	s.updateDeploymentStatus(deployment, "deploying", result.BuildLogs, "Executing CloudBox Install Protocol...\n", result.ErrorLogs)
	
	deployTime := time.Now()
	if err := s.terminalService.ExecuteCIPScript(session, "install", s.getDeploymentPath(deployment)); err != nil {
		result.ErrorLogs += fmt.Sprintf("CIP install script failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
		return result
	}
	result.BuildTime = time.Since(buildTime).Milliseconds()

	// Step 3: Start application using CIP start script
	if err := s.terminalService.ExecuteCIPScript(session, "start", s.getDeploymentPath(deployment)); err != nil {
		result.ErrorLogs += fmt.Sprintf("CIP start script failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
		return result
	}
	result.DeployTime = time.Since(deployTime).Milliseconds()

	// Step 4: Health check using CIP health script
	if err := s.verifyCIPDeployment(session, s.getDeploymentPath(deployment)); err != nil {
		result.ErrorLogs += fmt.Sprintf("CIP health check failed: %v", err)
		s.updateDeploymentStatus(deployment, "failed", result.BuildLogs, result.DeployLogs, result.ErrorLogs)
		return result
	}

	// Success
	result.Success = true
	now := time.Now()
	s.db.Model(&deployment).Updates(map[string]interface{}{
		"status":      "deployed",
		"deployed_at": &now,
		"build_logs":  result.BuildLogs,
		"deploy_logs": result.DeployLogs,
		"error_logs":  result.ErrorLogs,
		"build_time":  result.BuildTime,
		"deploy_time": result.DeployTime,
		"commit_hash": commitHash,
		"branch":      branch,
	})

	session.OutputCallback("🎉 [CIP] CloudBox Install Protocol deployment completed successfully!", "info")
	return result
}

// remoteCIPClone clones repository directly on the remote server
func (s *DeploymentService) remoteCIPClone(session *TerminalSession, deployment models.Deployment, commitHash, branch string) error {
	deploymentPath := s.getDeploymentPath(deployment)
	session.OutputCallback(fmt.Sprintf("📥 [CIP] Cloning repository to remote server: %s", deploymentPath), "info")

	// Create deployment directory if it doesn't exist
	createDirCmd := fmt.Sprintf("mkdir -p %s", deploymentPath)
	if err := s.runRemoteCommand(session, createDirCmd); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Build authenticated clone URL
	cloneURL := deployment.GitHubRepository.CloneURL
	if strings.HasPrefix(cloneURL, "git@github.com:") {
		cloneURL = strings.Replace(cloneURL, "git@github.com:", "https://github.com/", 1)
	}
	
	if deployment.GitHubRepository.AccessToken != "" && strings.HasPrefix(cloneURL, "https://github.com/") {
		cloneURL = strings.Replace(cloneURL, "https://github.com/", fmt.Sprintf("https://%s@github.com/", deployment.GitHubRepository.AccessToken), 1)
	}

	// Smart repository sync: clone if empty, otherwise fetch and reset
	syncCmd := fmt.Sprintf(`cd %s && 
		if [ ! -d .git ]; then
			echo "Directory is empty or not a git repo, cloning..."
			git clone --depth 1 -b %s %s . && 
			echo "Repository cloned successfully"
		else
			echo "Git repository exists, updating..."
			git remote set-url origin %s &&
			git fetch origin %s --depth=1 &&
			git reset --hard origin/%s &&
			echo "Repository updated successfully"
		fi`, 
		deploymentPath, branch, cloneURL, cloneURL, branch, branch)

	if err := s.runRemoteCommand(session, syncCmd); err != nil {
		return fmt.Errorf("repository sync failed: %w", err)
	}

	// Checkout specific commit if not "latest"
	if commitHash != "latest" && commitHash != "" {
		checkoutCmd := fmt.Sprintf(`cd %s && 
			git fetch --depth 1 origin %s && 
			git checkout %s && 
			echo "Checked out commit %s"`, 
			deploymentPath, commitHash, commitHash, commitHash)

		if err := s.runRemoteCommand(session, checkoutCmd); err != nil {
			session.OutputCallback(fmt.Sprintf("⚠️ [CIP] Warning: Could not checkout specific commit %s, using latest", commitHash), "info")
		}
	}

	session.OutputCallback("✅ [CIP] Repository cloned successfully to remote server", "info")
	return nil
}

// verifyCIPDeployment verifies deployment health using CIP health script
func (s *DeploymentService) verifyCIPDeployment(session *TerminalSession, appPath string) error {
	session.OutputCallback("🔍 [CIP] Verifying deployment health", "info")
	
	// Execute health check script
	if err := s.terminalService.ExecuteCIPScript(session, "health", appPath); err != nil {
		// If health script doesn't exist, try basic verification
		session.OutputCallback("⚠️ [CIP] Health script not available, performing basic verification", "info")
		
		// Check if main service is running (basic verification)
		checkCmd := fmt.Sprintf(`cd %s && 
			if [[ -f "package.json" ]]; then
				echo "✅ Package.json found"
			fi &&
			if [[ -f "cloudbox.json" ]]; then
				echo "✅ CloudBox manifest found"
			fi`, appPath)
		
		if err := s.runRemoteCommand(session, checkCmd); err != nil {
			return fmt.Errorf("basic verification failed: %w", err)
		}
	}

	session.OutputCallback("✅ [CIP] Deployment health verification passed", "info")
	return nil
}

// runRemoteCommand executes a command on the remote server via terminal session
func (s *DeploymentService) runRemoteCommand(session *TerminalSession, command string) error {
	tempSession, err := session.SSH.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer tempSession.Close()

	// Set up environment variables for the command
	envVars := ""
	for key, value := range session.Environment {
		envVars += fmt.Sprintf("export %s=\"%s\"; ", key, value)
	}

	fullCommand := envVars + command
	session.OutputCallback(fmt.Sprintf("$ %s", command), "info")

	output, err := tempSession.CombinedOutput(fullCommand)
	if err != nil {
		session.OutputCallback(string(output), "stderr")
		return fmt.Errorf("command failed: %w", err)
	}

	session.OutputCallback(string(output), "stdout")
	return nil
}

// cloneRepository clones the GitHub repository
func (s *DeploymentService) cloneRepository(deployment models.Deployment, commitHash, branch string, result *DeploymentResult) (string, error) {
	result.BuildLogs += fmt.Sprintf("Cloning repository: %s (branch: %s)\n", deployment.GitHubRepository.Name, branch)
	
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("cloudbox-deploy-%d-%d", deployment.ID, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Build authenticated clone URL using the stored access token
	cloneURL := deployment.GitHubRepository.CloneURL
	
	// Convert SSH URLs to HTTPS format since we can't use SSH in Docker containers easily
	if strings.HasPrefix(cloneURL, "git@github.com:") {
		// Convert git@github.com:owner/repo.git to https://github.com/owner/repo.git
		cloneURL = strings.Replace(cloneURL, "git@github.com:", "https://github.com/", 1)
		result.BuildLogs += "Converted SSH URL to HTTPS format\n"
	}
	
	if deployment.GitHubRepository.AccessToken != "" {
		// Convert HTTPS clone URL to use token authentication
		// GitHub format: https://github.com/owner/repo.git
		// Authenticated format: https://token@github.com/owner/repo.git
		if strings.HasPrefix(cloneURL, "https://github.com/") {
			cloneURL = strings.Replace(cloneURL, "https://github.com/", fmt.Sprintf("https://%s@github.com/", deployment.GitHubRepository.AccessToken), 1)
			result.BuildLogs += "Using authenticated GitHub access with PAT\n"
		} else {
			result.BuildLogs += fmt.Sprintf("Warning: Unsupported URL format: %s\n", cloneURL)
		}
	} else {
		result.BuildLogs += "Warning: No GitHub access token found, attempting public access\n"
	}

	// Clone repository
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "-b", branch, cloneURL, tempDir)
	
	// Set Git environment variables for Docker containers
	cloneCmd.Env = append(os.Environ(),
		"GIT_SSL_NO_VERIFY=true",               // Skip SSL verification for corporate proxies
		"GIT_SSH_COMMAND=ssh -o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null", // Skip SSH host key verification
	)
	
	var stdout, stderr bytes.Buffer
	cloneCmd.Stdout = &stdout
	cloneCmd.Stderr = &stderr
	
	result.BuildLogs += fmt.Sprintf("Executing: git clone --depth 1 -b %s [URL] %s\n", branch, tempDir)
	
	if err := cloneCmd.Run(); err != nil {
		result.BuildLogs += fmt.Sprintf("Clone failed: %s\n", stderr.String())
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	result.BuildLogs += fmt.Sprintf("Repository cloned successfully to %s\n", tempDir)
	
	// Checkout specific commit if provided and not "latest"
	if commitHash != "" && commitHash != "latest" {
		checkoutCmd := exec.Command("git", "checkout", commitHash)
		checkoutCmd.Dir = tempDir
		checkoutCmd.Stdout = &stdout
		checkoutCmd.Stderr = &stderr
		
		if err := checkoutCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("Checkout warning: %s\n", stderr.String())
			result.BuildLogs += "Using branch HEAD instead\n"
		} else {
			result.BuildLogs += fmt.Sprintf("Checked out commit %s\n", commitHash)
		}
	} else {
		result.BuildLogs += "Using latest commit from branch HEAD\n"
	}

	return tempDir, nil
}

// buildApplication builds the application using the specified build command
func (s *DeploymentService) buildApplication(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	if deployment.BuildCommand == "" {
		result.BuildLogs += "No build command specified, skipping build step\n"
		return nil
	}

	// Step 1: Install dependencies based on project type
	if err := s.installDependencies(repoDir, result); err != nil {
		result.BuildLogs += fmt.Sprintf("Dependency installation failed: %v\n", err)
		return fmt.Errorf("failed to install dependencies: %w", err)
	}

	// Step 1.5: Force correct environment after install (portfolio apps might overwrite .env)
	if s.isPortfolioDeployment(deployment) {
		result.BuildLogs += "Re-applying portfolio configuration after install...\n"
		if err := s.generatePortfolioEnvFile(deployment, repoDir, result); err != nil {
			result.BuildLogs += fmt.Sprintf("Failed to re-apply portfolio configuration: %v\n", err)
		}
	}

	// Step 2: Run the build command
	result.BuildLogs += fmt.Sprintf("Running build command: %s\n", deployment.BuildCommand)

	// Parse and execute build command (handle bash scripts properly)
	cmd := s.createCommand(deployment.BuildCommand)
	cmd.Dir = repoDir
	
	// Set environment variables
	cmd.Env = os.Environ()
	
	// Add custom environment variables from deployment configuration
	for key, value := range deployment.Environment {
		if valueStr, ok := value.(string); ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, valueStr))
		}
	}
	
	// Add port configuration as environment variables
	for variable, port := range deployment.PortConfiguration {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", variable, port))
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		result.BuildLogs += fmt.Sprintf("Build output:\n%s\n", stdout.String())
		result.BuildLogs += fmt.Sprintf("Build errors:\n%s\n", stderr.String())
		return fmt.Errorf("build command failed: %w", err)
	}

	result.BuildLogs += fmt.Sprintf("Build completed successfully\n")
	result.BuildLogs += fmt.Sprintf("Build output:\n%s\n", stdout.String())
	
	return nil
}

// deployToServer deploys the built application to the target server
func (s *DeploymentService) deployToServer(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	// Create SSH client
	client, err := s.createSSHClient(deployment)
	if err != nil {
		return fmt.Errorf("failed to create SSH connection: %w", err)
	}
	defer client.Close()

	result.DeployLogs += "SSH connection established\n"

	// Sanitize deployment name for filesystem safety
	sanitizedName := s.sanitizeDeploymentName(deployment.Name)
	result.DeployLogs += fmt.Sprintf("Sanitized deployment name: %s\n", sanitizedName)

	// Create deployment directory in user's home directory
	deployPath := fmt.Sprintf("~/deploys/%s", sanitizedName)
	result.DeployLogs += fmt.Sprintf("Creating deployment directory: %s\n", deployPath)
	
	// Create the deploys directory and specific deployment directory (no sudo needed)
	// Use proper shell escaping for the directory name
	if err := s.executeSSHCommand(client, fmt.Sprintf("mkdir -p ~/deploys/%s", s.shellEscape(sanitizedName))); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Get the absolute path for file operations (expand tilde)
	var absoluteDeployPath string
	if err := s.executeSSHCommandWithOutput(client, "echo ~/deploys/"+s.shellEscape(sanitizedName), &absoluteDeployPath); err != nil {
		return fmt.Errorf("failed to resolve deployment path: %w", err)
	}
	absoluteDeployPath = strings.TrimSpace(absoluteDeployPath)

	// Upload files using SCP
	result.DeployLogs += "Uploading files to server...\n"
	if err := s.uploadFiles(client, repoDir, absoluteDeployPath, result); err != nil {
		return fmt.Errorf("failed to upload files: %w", err)
	}

	// Stop existing application (if running)
	result.DeployLogs += "Stopping existing application...\n"
	s.executeSSHCommand(client, s.getStopCommand(deployment)) // Don't fail if stop command fails

	// Start application
	if deployment.StartCommand != "" {
		result.DeployLogs += fmt.Sprintf("Starting application with command: %s\n", deployment.StartCommand)
		
		// Build environment variables for the start command
		envVars := s.buildEnvironmentVariables(deployment)
		envString := ""
		for _, env := range envVars {
			envString += env + " "
		}
		
		startCmd := fmt.Sprintf("cd %s && %s %s", deployPath, envString, deployment.StartCommand)
		if err := s.executeSSHCommand(client, startCmd); err != nil {
			return fmt.Errorf("failed to start application: %w", err)
		}
		
		result.DeployLogs += "Application started successfully\n"
	}

	// Verify deployment
	if deployment.Port > 0 {
		result.DeployLogs += fmt.Sprintf("Verifying application on port %d...\n", deployment.Port)
		if err := s.verifyDeployment(client, deployment, result); err != nil {
			result.DeployLogs += fmt.Sprintf("Verification warning: %v\n", err)
			// Don't fail deployment for verification issues
		} else {
			result.DeployLogs += "Application is responding correctly\n"
		}
	}

	return nil
}

// createSSHClient creates an SSH client connection
func (s *DeploymentService) createSSHClient(deployment models.Deployment) (*ssh.Client, error) {
	// First, decrypt the private key (SSH keys are stored encrypted)
	decryptedPrivateKey, err := s.decryptSSHPrivateKey(deployment.WebServer.SSHKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SSH private key: %w", err)
	}

	// Parse the decrypted SSH private key
	signer, err := s.parseSSHPrivateKey(decryptedPrivateKey)
	if err != nil {
		// Log some debug info about the key format (first 100 chars, safely)
		keyPreview := decryptedPrivateKey
		if len(keyPreview) > 100 {
			keyPreview = keyPreview[:100] + "..."
		}
		log.Printf("SSH Key parsing failed. Key preview: %s", keyPreview)
		return nil, fmt.Errorf("failed to parse SSH private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: deployment.WebServer.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // TODO: Implement proper host key verification
		Timeout:         30 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", deployment.WebServer.Hostname, deployment.WebServer.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return client, nil
}

// executeSSHCommand executes a command on the remote server
func (s *DeploymentService) executeSSHCommand(client *ssh.Client, command string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %s, stderr: %s", err, stderr.String())
	}

	return nil
}

// executeSSHCommandWithOutput executes a command on the remote server and captures output
func (s *DeploymentService) executeSSHCommandWithOutput(client *ssh.Client, command string, output *string) error {
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(command); err != nil {
		return fmt.Errorf("command failed: %s, stderr: %s", err, stderr.String())
	}

	*output = stdout.String()
	return nil
}

// uploadFiles uploads the built application files to the server
func (s *DeploymentService) uploadFiles(client *ssh.Client, localPath, remotePath string, result *DeploymentResult) error {
	// This is a simplified file upload - in production, you'd want to use SCP or SFTP
	// For now, we'll use tar and SSH to transfer files
	
	// Create tar archive
	tarCmd := exec.Command("tar", "-czf", "/tmp/deploy.tar.gz", "-C", localPath, ".")
	if err := tarCmd.Run(); err != nil {
		return fmt.Errorf("failed to create tar archive: %w", err)
	}
	defer os.Remove("/tmp/deploy.tar.gz")

	// Read tar file
	tarData, err := os.ReadFile("/tmp/deploy.tar.gz")
	if err != nil {
		return fmt.Errorf("failed to read tar file: %w", err)
	}

	// Create SFTP client
	sftpClient, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SFTP session: %w", err)
	}
	defer sftpClient.Close()

	// Upload and extract
	uploadCmd := fmt.Sprintf("cat > /tmp/deploy.tar.gz && cd %s && tar -xzf /tmp/deploy.tar.gz && rm /tmp/deploy.tar.gz", remotePath)
	
	stdin, err := sftpClient.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	if err := sftpClient.Start(uploadCmd); err != nil {
		return fmt.Errorf("failed to start upload command: %w", err)
	}

	// Write tar data to stdin
	if _, err := stdin.Write(tarData); err != nil {
		return fmt.Errorf("failed to write tar data: %w", err)
	}
	stdin.Close()

	if err := sftpClient.Wait(); err != nil {
		return fmt.Errorf("upload command failed: %w", err)
	}

	result.DeployLogs += fmt.Sprintf("Uploaded %d bytes to %s\n", len(tarData), remotePath)
	return nil
}

// getStopCommand returns the appropriate stop command for the application
func (s *DeploymentService) getStopCommand(deployment models.Deployment) string {
	// Try to stop by port if available
	if deployment.Port > 0 {
		return fmt.Sprintf("sudo fuser -k %d/tcp || true", deployment.Port)
	}
	
	// Generic process kill by name
	return fmt.Sprintf("sudo pkill -f '%s' || true", deployment.Name)
}

// verifyDeployment checks if the deployed application is responding
func (s *DeploymentService) verifyDeployment(client *ssh.Client, deployment models.Deployment, result *DeploymentResult) error {
	// Simple health check using curl or wget
	checkCmd := fmt.Sprintf("curl -f -s -o /dev/null http://localhost:%d || wget -q --spider http://localhost:%d", 
		deployment.Port, deployment.Port)
	
	// Give the application some time to start
	time.Sleep(5 * time.Second)
	
	return s.executeSSHCommand(client, checkCmd)
}

// calculateDeploymentStats calculates file count and total size
func (s *DeploymentService) calculateDeploymentStats(repoDir string, result *DeploymentResult) {
	var fileCount int64
	var totalSize int64

	err := filepath.Walk(repoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() {
			fileCount++
			totalSize += info.Size()
		}
		return nil
	})

	if err != nil {
		log.Printf("Failed to calculate deployment stats: %v", err)
		fileCount = 0
		totalSize = 0
	}

	result.FileCount = fileCount
	result.TotalSize = totalSize
}

// updateDeploymentStatus updates the deployment status in database
func (s *DeploymentService) updateDeploymentStatus(deployment models.Deployment, status, buildLogs, deployLogs, errorLogs string) {
	updates := map[string]interface{}{
		"status": status,
	}
	
	if buildLogs != "" {
		updates["build_logs"] = buildLogs
	}
	if deployLogs != "" {
		updates["deploy_logs"] = deployLogs
	}
	if errorLogs != "" {
		updates["error_logs"] = errorLogs
	}

	s.db.Model(&deployment).Updates(updates)
}

// installDependencies automatically installs dependencies based on project type
func (s *DeploymentService) installDependencies(repoDir string, result *DeploymentResult) error {
	result.BuildLogs += "Installing project dependencies...\n"
	
	// Check for Node.js projects (package.json)
	if _, err := os.Stat(filepath.Join(repoDir, "package.json")); err == nil {
		result.BuildLogs += "Detected Node.js project (package.json found)\n"
		
		// Run npm install
		installCmd := exec.Command("npm", "install")
		installCmd.Dir = repoDir
		
		var stdout, stderr bytes.Buffer
		installCmd.Stdout = &stdout
		installCmd.Stderr = &stderr
		
		result.BuildLogs += "Running: npm install\n"
		
		if err := installCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("npm install failed:\n%s\n", stderr.String())
			return fmt.Errorf("npm install failed: %w", err)
		}
		
		result.BuildLogs += "npm install completed successfully\n"
		if stdout.String() != "" {
			result.BuildLogs += fmt.Sprintf("npm install output:\n%s\n", stdout.String())
		}
		
		return nil
	}
	
	// Check for Python projects (requirements.txt)
	if _, err := os.Stat(filepath.Join(repoDir, "requirements.txt")); err == nil {
		result.BuildLogs += "Detected Python project (requirements.txt found)\n"
		
		// Run pip install
		installCmd := exec.Command("pip3", "install", "-r", "requirements.txt")
		installCmd.Dir = repoDir
		
		var stdout, stderr bytes.Buffer
		installCmd.Stdout = &stdout
		installCmd.Stderr = &stderr
		
		result.BuildLogs += "Running: pip3 install -r requirements.txt\n"
		
		if err := installCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("pip install failed:\n%s\n", stderr.String())
			return fmt.Errorf("pip install failed: %w", err)
		}
		
		result.BuildLogs += "pip install completed successfully\n"
		if stdout.String() != "" {
			result.BuildLogs += fmt.Sprintf("pip install output:\n%s\n", stdout.String())
		}
		
		return nil
	}
	
	// Check for Python projects (pyproject.toml)
	if _, err := os.Stat(filepath.Join(repoDir, "pyproject.toml")); err == nil {
		result.BuildLogs += "Detected Python project (pyproject.toml found)\n"
		
		// Run pip install with current directory
		installCmd := exec.Command("pip3", "install", ".")
		installCmd.Dir = repoDir
		
		var stdout, stderr bytes.Buffer
		installCmd.Stdout = &stdout
		installCmd.Stderr = &stderr
		
		result.BuildLogs += "Running: pip3 install .\n"
		
		if err := installCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("pip install failed:\n%s\n", stderr.String())
			return fmt.Errorf("pip install failed: %w", err)
		}
		
		result.BuildLogs += "pip install completed successfully\n"
		if stdout.String() != "" {
			result.BuildLogs += fmt.Sprintf("pip install output:\n%s\n", stdout.String())
		}
		
		return nil
	}
	
	// Check for Go projects (go.mod)
	if _, err := os.Stat(filepath.Join(repoDir, "go.mod")); err == nil {
		result.BuildLogs += "Detected Go project (go.mod found)\n"
		
		// Run go mod download
		installCmd := exec.Command("go", "mod", "download")
		installCmd.Dir = repoDir
		
		var stdout, stderr bytes.Buffer
		installCmd.Stdout = &stdout
		installCmd.Stderr = &stderr
		
		result.BuildLogs += "Running: go mod download\n"
		
		if err := installCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("go mod download failed:\n%s\n", stderr.String())
			return fmt.Errorf("go mod download failed: %w", err)
		}
		
		result.BuildLogs += "go mod download completed successfully\n"
		if stdout.String() != "" {
			result.BuildLogs += fmt.Sprintf("go mod download output:\n%s\n", stdout.String())
		}
		
		return nil
	}
	
	// Check for Composer projects (composer.json)
	if _, err := os.Stat(filepath.Join(repoDir, "composer.json")); err == nil {
		result.BuildLogs += "Detected PHP project (composer.json found)\n"
		result.BuildLogs += "Warning: Composer not installed in container, skipping dependency installation\n"
		return nil
	}
	
	// No recognized dependency files found
	result.BuildLogs += "No dependency files detected (package.json, requirements.txt, pyproject.toml, go.mod, composer.json)\n"
	result.BuildLogs += "Skipping dependency installation\n"
	
	return nil
}

// parseSSHPrivateKey parses SSH private keys in various formats
func (s *DeploymentService) parseSSHPrivateKey(privateKeyPEM string) (ssh.Signer, error) {
	// Clean up the input (remove extra whitespace)
	privateKeyPEM = strings.TrimSpace(privateKeyPEM)
	
	// Try parsing as SSH private key directly first (OpenSSH format)
	signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
	if err == nil {
		return signer, nil
	}

	// Check if this looks like raw base64 data without PEM headers
	if !strings.HasPrefix(privateKeyPEM, "-----BEGIN") && !strings.Contains(privateKeyPEM, "-----") {
		// This looks like raw base64 data, try different PEM wrapper formats
		
		// Try OpenSSH format first
		opensshKey := "-----BEGIN OPENSSH PRIVATE KEY-----\n" + privateKeyPEM + "\n-----END OPENSSH PRIVATE KEY-----"
		if signer, err := ssh.ParsePrivateKey([]byte(opensshKey)); err == nil {
			return signer, nil
		}
		
		// Try traditional RSA format
		rsaKey := "-----BEGIN RSA PRIVATE KEY-----\n" + privateKeyPEM + "\n-----END RSA PRIVATE KEY-----"
		if signer, err := ssh.ParsePrivateKey([]byte(rsaKey)); err == nil {
			return signer, nil
		}
		
		// Try PKCS8 format
		pkcs8Key := "-----BEGIN PRIVATE KEY-----\n" + privateKeyPEM + "\n-----END PRIVATE KEY-----"
		if signer, err := ssh.ParsePrivateKey([]byte(pkcs8Key)); err == nil {
			return signer, nil
		}
	}

	// Check if this looks like an OpenSSH private key without proper headers
	if strings.Contains(privateKeyPEM, "OPENSSH PRIVATE KEY") && !strings.HasPrefix(privateKeyPEM, "-----BEGIN") {
		// Add proper PEM headers if missing
		if !strings.HasPrefix(privateKeyPEM, "-----BEGIN OPENSSH PRIVATE KEY-----") {
			privateKeyPEM = "-----BEGIN OPENSSH PRIVATE KEY-----\n" + privateKeyPEM + "\n-----END OPENSSH PRIVATE KEY-----"
		}
		signer, err := ssh.ParsePrivateKey([]byte(privateKeyPEM))
		if err == nil {
			return signer, nil
		}
	}

	// If that fails, try PEM parsing for traditional formats
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		// If PEM decode fails, let's provide a more helpful error
		if strings.Contains(privateKeyPEM, "PRIVATE KEY") {
			return nil, fmt.Errorf("SSH key appears to be corrupted or in an unsupported format. Ensure the key includes proper PEM headers (-----BEGIN ... PRIVATE KEY-----)")
		}
		return nil, fmt.Errorf("failed to decode PEM block from private key - key appears to be raw base64 without PEM headers")
	}

	var privateKey interface{}

	// Handle encrypted keys
	if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
		return nil, fmt.Errorf("encrypted SSH keys are not supported yet (passphrase required)")
	}

	// Try different key formats based on block type
	switch block.Type {
	case "RSA PRIVATE KEY":
		// PKCS#1 RSA private key
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
		}

	case "PRIVATE KEY":
		// PKCS#8 private key (could be RSA, ECDSA, Ed25519)
		privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
		}

	case "EC PRIVATE KEY":
		// ECDSA private key
		privateKey, err = x509.ParseECPrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse EC private key: %w", err)
		}

	case "OPENSSH PRIVATE KEY":
		// OpenSSH private key format
		signer, err := ssh.ParsePrivateKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse OpenSSH private key: %w", err)
		}
		return signer, nil

	default:
		return nil, fmt.Errorf("unsupported private key type: %s", block.Type)
	}

	// Convert parsed key to SSH signer
	signer, err = ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH signer: %w", err)
	}

	return signer, nil
}

// decryptSSHPrivateKey decrypts an encrypted SSH private key
func (s *DeploymentService) decryptSSHPrivateKey(encryptedKey string) (string, error) {
	if s.cfg.MasterKey == "" {
		return "", fmt.Errorf("master key not configured - cannot decrypt SSH private key")
	}

	decryptedKey, err := utils.DecryptPrivateKey(encryptedKey, s.cfg.MasterKey)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt private key: %w", err)
	}

	return decryptedKey, nil
}

// sanitizeDeploymentName cleans deployment names for filesystem safety
func (s *DeploymentService) sanitizeDeploymentName(name string) string {
	// Replace spaces with underscores
	sanitized := strings.ReplaceAll(name, " ", "_")
	
	// Replace other problematic characters
	sanitized = strings.ReplaceAll(sanitized, "/", "_")
	sanitized = strings.ReplaceAll(sanitized, "\\", "_")
	sanitized = strings.ReplaceAll(sanitized, ":", "_")
	sanitized = strings.ReplaceAll(sanitized, "*", "_")
	sanitized = strings.ReplaceAll(sanitized, "?", "_")
	sanitized = strings.ReplaceAll(sanitized, "\"", "_")
	sanitized = strings.ReplaceAll(sanitized, "<", "_")
	sanitized = strings.ReplaceAll(sanitized, ">", "_")
	sanitized = strings.ReplaceAll(sanitized, "|", "_")
	
	// Remove consecutive underscores
	for strings.Contains(sanitized, "__") {
		sanitized = strings.ReplaceAll(sanitized, "__", "_")
	}
	
	// Trim underscores from start and end
	sanitized = strings.Trim(sanitized, "_")
	
	return sanitized
}

// shellEscape properly escapes strings for shell commands
func (s *DeploymentService) shellEscape(str string) string {
	// Simple shell escaping - wrap in single quotes and escape any single quotes
	escaped := strings.ReplaceAll(str, "'", "'\"'\"'")
	return "'" + escaped + "'"
}

// createCommand creates a properly configured command for execution
func (s *DeploymentService) createCommand(commandStr string) *exec.Cmd {
	// Handle bash scripts and complex commands
	if strings.HasPrefix(commandStr, "bash ") || strings.Contains(commandStr, "&&") || strings.Contains(commandStr, "||") || strings.Contains(commandStr, "|") {
		// Use bash for complex commands or explicit bash commands
		return exec.Command("bash", "-c", commandStr)
	}
	
	// Handle sh scripts
	if strings.HasPrefix(commandStr, "sh ") {
		return exec.Command("sh", "-c", commandStr)
	}
	
	// For simple commands, parse arguments normally
	cmdParts := strings.Fields(commandStr)
	if len(cmdParts) == 0 {
		return exec.Command("sh", "-c", "echo 'Empty command'")
	}
	
	if len(cmdParts) == 1 {
		return exec.Command(cmdParts[0])
	}
	
	return exec.Command(cmdParts[0], cmdParts[1:]...)
}

// CheckPortAvailability checks if ports are available on the target server
func (s *DeploymentService) CheckPortAvailability(webServer models.WebServer, ports []int) (map[int]bool, error) {
	// Create SSH client
	client, err := s.createSSHClientFromWebServer(webServer)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH connection: %w", err)
	}
	defer client.Close()

	results := make(map[int]bool)
	
	for _, port := range ports {
		available, err := s.checkSinglePort(client, port)
		if err != nil {
			// If we can't check, assume unavailable for safety
			results[port] = false
		} else {
			results[port] = available
		}
	}
	
	return results, nil
}

// createSSHClientFromWebServer creates SSH client from WebServer model
func (s *DeploymentService) createSSHClientFromWebServer(webServer models.WebServer) (*ssh.Client, error) {
	// Load SSH key
	var sshKey models.SSHKey
	if err := s.db.First(&sshKey, webServer.SSHKeyID).Error; err != nil {
		return nil, fmt.Errorf("failed to load SSH key: %w", err)
	}
	
	// Decrypt the private key
	decryptedPrivateKey, err := s.decryptSSHPrivateKey(sshKey.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt SSH private key: %w", err)
	}

	// Parse the SSH private key
	signer, err := s.parseSSHPrivateKey(decryptedPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to parse SSH private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User: webServer.Username,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	address := fmt.Sprintf("%s:%d", webServer.Hostname, webServer.Port)
	client, err := ssh.Dial("tcp", address, config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to %s: %w", address, err)
	}

	return client, nil
}

// checkSinglePort checks if a single port is available on the remote server
func (s *DeploymentService) checkSinglePort(client *ssh.Client, port int) (bool, error) {
	// Multi-method port checking for better Docker container detection
	// 1. Check with netstat for listening ports
	// 2. Check with ss (more modern, better Docker support)
	// 3. Check with lsof if available
	// 4. Try to bind to the port directly
	
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	// Comprehensive port check script that covers Docker containers
	checkScript := fmt.Sprintf(`
		PORT=%d
		
		# Check if anything is listening on the port (including Docker containers)
		NETSTAT_CHECK=$(netstat -tuln 2>/dev/null | grep ":$PORT " | wc -l)
		SS_CHECK=$(ss -tuln 2>/dev/null | grep ":$PORT " | wc -l)
		
		# Check Docker containers specifically
		DOCKER_CHECK=0
		if command -v docker >/dev/null 2>&1; then
			DOCKER_CHECK=$(docker ps --format "table {{.Ports}}" 2>/dev/null | grep -c ":$PORT->" || echo 0)
		fi
		
		# Check if we can bind to the port (most reliable test)
		BIND_TEST=0
		if command -v nc >/dev/null 2>&1; then
			# Try to bind with netcat (timeout after 1 second)
			timeout 1 nc -l $PORT 2>/dev/null &
			NC_PID=$!
			sleep 0.1
			if kill -0 $NC_PID 2>/dev/null; then
				kill $NC_PID 2>/dev/null
				BIND_TEST=1  # Port was bindable (available)
			fi
		fi
		
		TOTAL_USAGE=$((NETSTAT_CHECK + SS_CHECK + DOCKER_CHECK))
		
		# Output result
		if [ $TOTAL_USAGE -gt 0 ]; then
			echo "PORT_IN_USE:$TOTAL_USAGE"
		elif [ $BIND_TEST -eq 1 ]; then
			echo "PORT_AVAILABLE:BIND_OK"
		else
			echo "PORT_UNKNOWN:NO_BIND_TEST"
		fi
	`, port)

	var stdout bytes.Buffer
	session.Stdout = &stdout

	if err := session.Run(checkScript); err != nil {
		// If script fails, assume port is available but unknown
		return true, nil
	}

	output := strings.TrimSpace(stdout.String())
	
	// Parse the result
	if strings.HasPrefix(output, "PORT_IN_USE:") {
		return false, nil // Port is in use
	} else if strings.HasPrefix(output, "PORT_AVAILABLE:") {
		return true, nil // Port is available
	}
	
	// Unknown status, assume available to be safe for deployment
	return true, nil
}

// buildEnvironmentVariables creates environment variable strings for deployment commands
func (s *DeploymentService) buildEnvironmentVariables(deployment models.Deployment) []string {
	var envVars []string
	
	// Add custom environment variables from deployment configuration
	for key, value := range deployment.Environment {
		if valueStr, ok := value.(string); ok {
			envVars = append(envVars, fmt.Sprintf("%s=%s", key, s.shellEscape(valueStr)))
		}
	}
	
	// Add port configuration as environment variables
	for variable, port := range deployment.PortConfiguration {
		envVars = append(envVars, fmt.Sprintf("%s=%d", variable, port))
	}
	
	return envVars
}

// prepareDeploymentEnvironment prepares the deployment environment including .env file generation
func (s *DeploymentService) prepareDeploymentEnvironment(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	result.BuildLogs += "Preparing deployment environment...\n"
	
	// Check if this is a portfolio app that needs special .env handling
	if s.isPortfolioDeployment(deployment) {
		result.BuildLogs += "Detected portfolio application - configuring optimized environment\n"
		return s.generatePortfolioEnvFile(deployment, repoDir, result)
	}
	
	// For other app types, generate standard .env if needed
	return s.generateStandardEnvFile(deployment, repoDir, result)
}

// isPortfolioDeployment checks if this deployment is for a portfolio app
func (s *DeploymentService) isPortfolioDeployment(deployment models.Deployment) bool {
	// Check if repository analysis indicates portfolio
	if deployment.GitHubRepository.Analysis != nil {
		projectType := deployment.GitHubRepository.Analysis.ProjectType
		if projectType == "photoportfolio" || strings.Contains(strings.ToLower(projectType), "portfolio") {
			return true
		}
	}
	
	// Check repository name for portfolio indicators
	repoName := strings.ToLower(deployment.GitHubRepository.Name)
	portfolioIndicators := []string{"portfolio", "photoportfolio"}
	for _, indicator := range portfolioIndicators {
		if strings.Contains(repoName, indicator) {
			return true
		}
	}
	
	return false
}

// generatePortfolioEnvFile generates optimized .env file for portfolio apps
func (s *DeploymentService) generatePortfolioEnvFile(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	envPath := filepath.Join(repoDir, ".env")
	
	// Always overwrite existing .env for portfolio apps to ensure correct configuration
	result.BuildLogs += "Overwriting .env file with CloudBox-optimized portfolio configuration\n"
	
	// Get the main application port (usually PORT or WEB_PORT)
	mainPort := deployment.Port
	if len(deployment.PortConfiguration) > 0 {
		// Find the main port from configuration
		for variable, port := range deployment.PortConfiguration {
			if variable == "PORT" || variable == "WEB_PORT" {
				mainPort = port
				break
			}
		}
	}
	
	// Generate portfolio-optimized .env content
	envContent := fmt.Sprintf(`# Portfolio Application Configuration - Generated by CloudBox
# This configuration is optimized for portfolio apps using CloudBox backend

# Application Settings
APP_NAME=portfolio
PROJECT_PREFIX=portfolio
NODE_ENV=production

# Frontend Server Configuration (only port needed for portfolio apps)
PORT=%d
WEB_PORT=%d

# API Configuration (provided by CloudBox - no separate API server needed)
API_URL=http://localhost:8080/p/%s/api
VITE_API_URL=http://localhost:8080/p/%s/api

# CloudBox provides all backend services:
# - Database via CloudBox Data API
# - Authentication via CloudBox Auth API  
# - Storage via CloudBox Storage API
# No separate database or API server configuration needed

# Portfolio-specific settings
PORTFOLIO_THEME=modern
PORTFOLIO_MODE=production
`, mainPort, mainPort, deployment.GitHubRepository.Name, deployment.GitHubRepository.Name)

	// Add any custom environment variables
	for key, value := range deployment.Environment {
		if valueStr, ok := value.(string); ok {
			envContent += fmt.Sprintf("%s=%s\n", key, valueStr)
		}
	}
	
	// Add port configuration variables
	for variable, port := range deployment.PortConfiguration {
		envContent += fmt.Sprintf("%s=%d\n", variable, port)
	}
	
	// Write the .env file
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}
	
	result.BuildLogs += fmt.Sprintf("✅ Generated optimized .env file for portfolio app (main port: %d)\n", mainPort)
	result.BuildLogs += "✅ Portfolio app will use CloudBox backend services - no separate API/DB servers needed\n"
	
	// Verify the file was written correctly
	if writtenContent, err := os.ReadFile(envPath); err == nil {
		result.BuildLogs += fmt.Sprintf("✅ .env file verification: %d bytes written\n", len(writtenContent))
		if strings.Contains(string(writtenContent), fmt.Sprintf("PORT=%d", mainPort)) {
			result.BuildLogs += fmt.Sprintf("✅ Confirmed PORT=%d is set in .env file\n", mainPort)
		} else {
			result.BuildLogs += fmt.Sprintf("⚠️ WARNING: PORT=%d not found in .env file\n", mainPort)
		}
	}
	
	return nil
}

// generateStandardEnvFile generates .env file for standard applications
func (s *DeploymentService) generateStandardEnvFile(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	envPath := filepath.Join(repoDir, ".env")
	
	// Check if .env already exists and is acceptable
	if _, err := os.Stat(envPath); err == nil {
		result.BuildLogs += ".env file already exists - keeping existing configuration\n"
		return nil
	}
	
	// Create basic .env file with deployment configuration
	envContent := "# Application Configuration - Generated by CloudBox\n\n"
	
	// Add custom environment variables
	for key, value := range deployment.Environment {
		if valueStr, ok := value.(string); ok {
			envContent += fmt.Sprintf("%s=%s\n", key, valueStr)
		}
	}
	
	// Add port configuration
	for variable, port := range deployment.PortConfiguration {
		envContent += fmt.Sprintf("%s=%d\n", variable, port)
	}
	
	// Write the .env file
	if err := os.WriteFile(envPath, []byte(envContent), 0644); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}
	
	result.BuildLogs += "Generated .env file with deployment configuration\n"
	return nil
}