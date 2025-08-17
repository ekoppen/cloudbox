package services

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
	"github.com/cloudbox/backend/internal/models"
	"github.com/cloudbox/backend/internal/security"
	"gorm.io/gorm"
)

type PluginService struct {
	db        *gorm.DB
	cfg       *config.Config
	validator *security.PluginValidator
}

func NewPluginService(db *gorm.DB, cfg *config.Config) *PluginService {
	return &PluginService{
		db:        db,
		cfg:       cfg,
		validator: security.NewPluginValidator(cfg),
	}
}

type GitHubReleaseAsset struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	ContentType string `json:"content_type"`
	Size        int    `json:"size"`
	BrowserDownloadURL string `json:"browser_download_url"`
	State       string `json:"state"`
}

type GitHubRelease struct {
	ID        int64                `json:"id"`
	TagName   string               `json:"tag_name"`
	Name      string               `json:"name"`
	Body      string               `json:"body"`
	Draft     bool                 `json:"draft"`
	Prerelease bool                `json:"prerelease"`
	CreatedAt string               `json:"created_at"`
	PublishedAt string             `json:"published_at"`
	Assets    []GitHubReleaseAsset `json:"assets"`
}

type PluginManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	License      string            `json:"license"`
	Repository   string            `json:"repository"`
	Type         string            `json:"type"`
	Main         string            `json:"main"`
	Permissions  []string          `json:"permissions"`
	Dependencies map[string]string `json:"dependencies"`
	UI           map[string]interface{} `json:"ui"`
	// Security fields
	Signature    string            `json:"signature,omitempty"`
	Checksum     string            `json:"checksum,omitempty"`
}

// DownloadAndInstallPlugin downloads a plugin from GitHub and installs it
func (ps *PluginService) DownloadAndInstallPlugin(repoURL, version string, projectID, userID uint) (*models.PluginInstallation, error) {
	// Validate repository
	repo, err := ps.validator.ValidateGitHubRepository(repoURL)
	if err != nil {
		return nil, fmt.Errorf("repository validation failed: %v", err)
	}

	// Create download record
	download := &models.PluginDownload{
		PluginName:     repo.Name,
		PluginVersion:  version,
		ProjectID:      projectID,
		UserID:         userID,
		DownloadSource: repoURL,
		DownloadStatus: "started",
		StartedAt:      time.Now(),
	}
	ps.db.Create(download)

	// Get latest release if version not specified
	if version == "" {
		release, err := ps.getLatestRelease(repo.Owner.Login, repo.Name)
		if err != nil {
			ps.updateDownloadStatus(download, "failed", err.Error())
			return nil, fmt.Errorf("failed to get latest release: %v", err)
		}
		version = release.TagName
	}

	// Download plugin files
	pluginPath := filepath.Join("./plugins", repo.Name)
	err = ps.downloadPluginFromGitHub(repo.Owner.Login, repo.Name, version, pluginPath)
	if err != nil {
		ps.updateDownloadStatus(download, "failed", err.Error())
		return nil, fmt.Errorf("failed to download plugin: %v", err)
	}

	// Validate plugin manifest
	manifestPath := filepath.Join(pluginPath, "plugin.json")
	manifest, err := ps.validatePluginManifest(manifestPath)
	if err != nil {
		ps.updateDownloadStatus(download, "failed", err.Error())
		return nil, fmt.Errorf("invalid plugin manifest: %v", err)
	}

	// Create installation record
	installation := &models.PluginInstallation{
		PluginName:       manifest.Name,
		PluginVersion:    manifest.Version,
		ProjectID:        projectID,
		Status:           "disabled", // Disabled by default for security
		InstallationPath: pluginPath,
		InstalledBy:      userID,
		InstalledAt:      time.Now(),
		Config:          make(map[string]interface{}),
		Environment:     make(map[string]interface{}),
	}

	err = ps.db.Create(installation).Error
	if err != nil {
		ps.updateDownloadStatus(download, "failed", err.Error())
		return nil, fmt.Errorf("failed to create installation record: %v", err)
	}

	// Create plugin state
	state := &models.PluginState{
		PluginName:     manifest.Name,
		ProjectID:      projectID,
		CurrentStatus:  "disabled",
		StateChangedAt: time.Now(),
		StateChangedBy: &userID,
		HealthStatus:   "unknown",
		HealthDetails:  make(map[string]interface{}),
	}
	ps.db.Create(state)

	// Update download as completed
	ps.updateDownloadStatus(download, "completed", "")

	log.Printf("Plugin %s v%s installed successfully for project %d", manifest.Name, manifest.Version, projectID)
	return installation, nil
}

// UninstallPlugin completely removes a plugin installation
func (ps *PluginService) UninstallPlugin(pluginName string, projectID uint) error {
	// Get installation record
	var installation models.PluginInstallation
	err := ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		return fmt.Errorf("plugin not found: %v", err)
	}

	// Stop plugin if running
	err = ps.StopPlugin(pluginName, projectID)
	if err != nil {
		log.Printf("Warning: Failed to stop plugin before uninstall: %v", err)
	}

	// Remove plugin files
	if installation.InstallationPath != "" {
		err = os.RemoveAll(installation.InstallationPath)
		if err != nil {
			log.Printf("Warning: Failed to remove plugin files: %v", err)
		}
	}

	// Delete database records
	err = ps.db.Transaction(func(tx *gorm.DB) error {
		// Delete plugin state
		if err := tx.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).Delete(&models.PluginState{}).Error; err != nil {
			return err
		}

		// Delete installation record
		if err := tx.Delete(&installation).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("failed to remove plugin from database: %v", err)
	}

	log.Printf("Plugin %s uninstalled successfully from project %d", pluginName, projectID)
	return nil
}

// StartPlugin starts a plugin process
func (ps *PluginService) StartPlugin(pluginName string, projectID uint) error {
	// Get installation
	var installation models.PluginInstallation
	err := ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		return fmt.Errorf("plugin not found: %v", err)
	}

	if installation.Status != "enabled" {
		return fmt.Errorf("plugin is not enabled")
	}

	// Load plugin manifest
	manifestPath := filepath.Join(installation.InstallationPath, "plugin.json")
	_, err = ps.loadPluginManifest(manifestPath)
	if err != nil {
		return fmt.Errorf("failed to load plugin manifest: %v", err)
	}

	// In a real implementation, this would:
	// 1. Start the plugin process using the main file
	// 2. Set up environment variables
	// 3. Monitor process health
	// 4. Update plugin state with process information

	// Update plugin state
	now := time.Now()
	var state models.PluginState
	err = ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&state).Error
	if err == gorm.ErrRecordNotFound {
		state = models.PluginState{
			PluginName:     pluginName,
			ProjectID:      projectID,
			CurrentStatus:  "running",
			StateChangedAt: now,
			HealthStatus:   "healthy",
			HealthDetails:  make(map[string]interface{}),
		}
		ps.db.Create(&state)
	} else if err == nil {
		state.CurrentStatus = "running"
		state.StateChangedAt = now
		state.HealthStatus = "healthy"
		state.LastHealthCheck = &now
		ps.db.Save(&state)
	}

	log.Printf("Plugin %s started successfully (mock) for project %d", pluginName, projectID)
	return nil
}

// StopPlugin stops a plugin process
func (ps *PluginService) StopPlugin(pluginName string, projectID uint) error {
	// Get plugin state
	var state models.PluginState
	err := ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&state).Error
	if err != nil {
		return fmt.Errorf("plugin state not found: %v", err)
	}

	// In a real implementation, this would:
	// 1. Gracefully stop the plugin process
	// 2. Clean up resources
	// 3. Update process information

	// Update plugin state
	now := time.Now()
	state.CurrentStatus = "stopped"
	state.StateChangedAt = now
	state.ProcessID = nil
	state.Port = nil
	state.HealthStatus = "unknown"
	ps.db.Save(&state)

	log.Printf("Plugin %s stopped successfully for project %d", pluginName, projectID)
	return nil
}

// CheckPluginHealth performs health check on a plugin
func (ps *PluginService) CheckPluginHealth(pluginName string, projectID uint) error {
	// Get plugin state
	var state models.PluginState
	err := ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&state).Error
	if err != nil {
		return fmt.Errorf("plugin state not found: %v", err)
	}

	if state.CurrentStatus != "running" {
		return fmt.Errorf("plugin is not running")
	}

	// In a real implementation, this would:
	// 1. Check if process is still running
	// 2. Ping health endpoint if available
	// 3. Check resource usage
	// 4. Validate plugin functionality

	// Mock health check
	now := time.Now()
	state.LastHealthCheck = &now
	state.HealthStatus = "healthy"
	state.HealthDetails = map[string]interface{}{
		"last_check": now.Format(time.RFC3339),
		"checks": map[string]interface{}{
			"process":  "running",
			"memory":   "normal",
			"cpu":      "normal",
			"response": "ok",
		},
	}
	ps.db.Save(&state)

	return nil
}

// GetPluginLogs retrieves plugin logs
func (ps *PluginService) GetPluginLogs(pluginName string, projectID uint, lines int) ([]string, error) {
	// In a real implementation, this would:
	// 1. Read plugin log files
	// 2. Parse and format log entries
	// 3. Return recent log lines

	// Mock log data
	logs := []string{
		"[2024-08-17 12:00:00] Plugin started successfully",
		"[2024-08-17 12:00:01] Configuration loaded",
		"[2024-08-17 12:00:02] Connected to database",
		"[2024-08-17 12:00:03] Plugin ready to serve requests",
	}

	if lines > 0 && lines < len(logs) {
		return logs[len(logs)-lines:], nil
	}

	return logs, nil
}

// UpdatePluginFromRegistry updates a plugin to the latest version
func (ps *PluginService) UpdatePluginFromRegistry(pluginName string, projectID uint) error {
	// Get current installation
	var installation models.PluginInstallation
	err := ps.db.Where("plugin_name = ? AND project_id = ?", pluginName, projectID).First(&installation).Error
	if err != nil {
		return fmt.Errorf("plugin not found: %v", err)
	}

	// In a real implementation, this would:
	// 1. Check for newer versions in registry
	// 2. Download and validate new version
	// 3. Backup current installation
	// 4. Install new version
	// 5. Migrate configuration if needed
	// 6. Restart plugin if it was running

	log.Printf("Plugin %s update check completed (mock) for project %d", pluginName, projectID)
	return nil
}

// Helper methods

// getLatestRelease fetches the latest release from GitHub
func (ps *PluginService) getLatestRelease(owner, repo string) (*GitHubRelease, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s/releases/latest", owner, repo)
	
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	if ps.cfg.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+ps.cfg.GitHubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox-Plugin-Service")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API error: status %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var release GitHubRelease
	err = json.Unmarshal(body, &release)
	if err != nil {
		return nil, err
	}

	return &release, nil
}

// downloadPluginFromGitHub downloads plugin source code
func (ps *PluginService) downloadPluginFromGitHub(owner, repo, version, destPath string) error {
	// Create destination directory
	err := os.MkdirAll(destPath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create plugin directory: %v", err)
	}

	// Download source archive
	downloadURL := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s.zip", owner, repo, version)
	
	client := &http.Client{Timeout: 5 * time.Minute}
	req, err := http.NewRequest("GET", downloadURL, nil)
	if err != nil {
		return err
	}

	if ps.cfg.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+ps.cfg.GitHubToken)
	}
	req.Header.Set("User-Agent", "CloudBox-Plugin-Service")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("download failed: status %d", resp.StatusCode)
	}

	// Save to temporary file
	tmpFile, err := ioutil.TempFile("", "plugin-*.zip")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, resp.Body)
	if err != nil {
		return err
	}

	// Extract archive
	return ps.extractPluginArchive(tmpFile.Name(), destPath, fmt.Sprintf("%s-%s", repo, strings.TrimPrefix(version, "v")))
}

// extractPluginArchive extracts a plugin ZIP archive
func (ps *PluginService) extractPluginArchive(archivePath, destPath, stripPrefix string) error {
	reader, err := zip.OpenReader(archivePath)
	if err != nil {
		return err
	}
	defer reader.Close()

	for _, file := range reader.File {
		// Skip directories and strip prefix
		if file.FileInfo().IsDir() {
			continue
		}

		path := file.Name
		if stripPrefix != "" && strings.HasPrefix(path, stripPrefix+"/") {
			path = strings.TrimPrefix(path, stripPrefix+"/")
		}

		// Skip files outside plugin directory
		if strings.Contains(path, "..") {
			continue
		}

		destFile := filepath.Join(destPath, path)
		err = os.MkdirAll(filepath.Dir(destFile), 0755)
		if err != nil {
			return err
		}

		// Extract file
		src, err := file.Open()
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := os.Create(destFile)
		if err != nil {
			return err
		}
		defer dst.Close()

		_, err = io.Copy(dst, src)
		if err != nil {
			return err
		}
	}

	return nil
}

// validatePluginManifest validates a plugin.json file
func (ps *PluginService) validatePluginManifest(manifestPath string) (*PluginManifest, error) {
	// Check if manifest exists
	if _, err := os.Stat(manifestPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("plugin.json not found")
	}

	// Read manifest
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read manifest: %v", err)
	}

	// Parse JSON
	var manifest PluginManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, fmt.Errorf("invalid JSON: %v", err)
	}

	// Use security validator for detailed validation
	securityManifest, err := ps.validator.ValidatePluginManifest(data)
	if err != nil {
		return nil, err
	}

	// Convert security manifest to our format
	manifest.Name = securityManifest.Name
	manifest.Version = securityManifest.Version
	manifest.Description = securityManifest.Description
	manifest.Author = securityManifest.Author
	manifest.License = securityManifest.License
	manifest.Repository = securityManifest.Repository
	manifest.Permissions = securityManifest.Permissions
	manifest.Dependencies = securityManifest.Dependencies
	manifest.Signature = securityManifest.Signature
	manifest.Checksum = securityManifest.Checksum

	return &manifest, nil
}

// loadPluginManifest loads a plugin manifest without validation
func (ps *PluginService) loadPluginManifest(manifestPath string) (*PluginManifest, error) {
	data, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		return nil, err
	}

	var manifest PluginManifest
	err = json.Unmarshal(data, &manifest)
	if err != nil {
		return nil, err
	}

	return &manifest, nil
}

// updateDownloadStatus updates the status of a plugin download
func (ps *PluginService) updateDownloadStatus(download *models.PluginDownload, status, errorMsg string) {
	download.DownloadStatus = status
	download.ErrorMessage = errorMsg

	now := time.Now()
	if status == "completed" {
		download.CompletedAt = &now
	} else if status == "failed" {
		download.FailedAt = &now
		download.ErrorCode = "DOWNLOAD_FAILED"
	}

	ps.db.Save(download)
}

// calculateFileChecksum calculates SHA256 checksum of a file
func (ps *PluginService) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	_, err = io.Copy(hash, file)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}