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

	"github.com/cloudbox/backend/internal/models"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
)

// DeploymentService handles real deployment operations
type DeploymentService struct {
	db *gorm.DB
}

// NewDeploymentService creates a new deployment service
func NewDeploymentService(db *gorm.DB) *DeploymentService {
	return &DeploymentService{db: db}
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

	// Step 2: Build application
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

// cloneRepository clones the GitHub repository
func (s *DeploymentService) cloneRepository(deployment models.Deployment, commitHash, branch string, result *DeploymentResult) (string, error) {
	result.BuildLogs += "Cloning repository...\n"
	
	// Create temporary directory
	tempDir := filepath.Join(os.TempDir(), fmt.Sprintf("cloudbox-deploy-%d-%d", deployment.ID, time.Now().Unix()))
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create temp directory: %w", err)
	}

	// Clone repository
	cloneCmd := exec.Command("git", "clone", "--depth", "1", "-b", branch, deployment.GitHubRepository.CloneURL, tempDir)
	
	var stdout, stderr bytes.Buffer
	cloneCmd.Stdout = &stdout
	cloneCmd.Stderr = &stderr
	
	if err := cloneCmd.Run(); err != nil {
		result.BuildLogs += fmt.Sprintf("Clone failed: %s\n", stderr.String())
		return "", fmt.Errorf("git clone failed: %w", err)
	}

	result.BuildLogs += fmt.Sprintf("Repository cloned successfully to %s\n", tempDir)
	
	// Checkout specific commit if provided
	if commitHash != "" {
		checkoutCmd := exec.Command("git", "checkout", commitHash)
		checkoutCmd.Dir = tempDir
		checkoutCmd.Stdout = &stdout
		checkoutCmd.Stderr = &stderr
		
		if err := checkoutCmd.Run(); err != nil {
			result.BuildLogs += fmt.Sprintf("Checkout warning: %s\n", stderr.String())
			// Don't fail deployment for checkout issues, use branch HEAD
		} else {
			result.BuildLogs += fmt.Sprintf("Checked out commit %s\n", commitHash)
		}
	}

	return tempDir, nil
}

// buildApplication builds the application using the specified build command
func (s *DeploymentService) buildApplication(deployment models.Deployment, repoDir string, result *DeploymentResult) error {
	if deployment.BuildCommand == "" {
		result.BuildLogs += "No build command specified, skipping build step\n"
		return nil
	}

	result.BuildLogs += fmt.Sprintf("Running build command: %s\n", deployment.BuildCommand)

	// Parse build command (handle complex commands)
	cmdParts := strings.Fields(deployment.BuildCommand)
	if len(cmdParts) == 0 {
		return fmt.Errorf("empty build command")
	}

	cmd := exec.Command(cmdParts[0], cmdParts[1:]...)
	cmd.Dir = repoDir
	
	// Set environment variables
	cmd.Env = os.Environ()
	for key, value := range deployment.Environment {
		if valueStr, ok := value.(string); ok {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, valueStr))
		}
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

	// Create deployment directory on server
	deployPath := fmt.Sprintf("/var/www/%s", deployment.Name)
	result.DeployLogs += fmt.Sprintf("Creating deployment directory: %s\n", deployPath)
	
	if err := s.executeSSHCommand(client, fmt.Sprintf("sudo mkdir -p %s && sudo chown %s:%s %s", 
		deployPath, deployment.WebServer.Username, deployment.WebServer.Username, deployPath)); err != nil {
		return fmt.Errorf("failed to create deployment directory: %w", err)
	}

	// Upload files using SCP
	result.DeployLogs += "Uploading files to server...\n"
	if err := s.uploadFiles(client, repoDir, deployPath, result); err != nil {
		return fmt.Errorf("failed to upload files: %w", err)
	}

	// Stop existing application (if running)
	result.DeployLogs += "Stopping existing application...\n"
	s.executeSSHCommand(client, s.getStopCommand(deployment)) // Don't fail if stop command fails

	// Start application
	if deployment.StartCommand != "" {
		result.DeployLogs += fmt.Sprintf("Starting application with command: %s\n", deployment.StartCommand)
		
		startCmd := fmt.Sprintf("cd %s && %s", deployPath, deployment.StartCommand)
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
	// Decrypt private key
	block, _ := pem.Decode([]byte(deployment.WebServer.SSHKey.PrivateKey))
	if block == nil {
		return nil, fmt.Errorf("failed to parse SSH private key")
	}

	var privateKey interface{}
	var err error

	// Try to parse encrypted key first
	if block.Headers["Proc-Type"] == "4,ENCRYPTED" {
		// For now, assume no passphrase (we'll need to extend this)
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
		if err != nil {
			// Try PKCS8 format
			privateKey, err = x509.ParsePKCS8PrivateKey(block.Bytes)
		}
	}
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse private key: %w", err)
	}

	signer, err := ssh.NewSignerFromKey(privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
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