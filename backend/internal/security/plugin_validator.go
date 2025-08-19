package security

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/config"
)

// PluginValidator handles secure plugin validation and GitHub repository verification
type PluginValidator struct {
	cfg *config.Config
}

// NewPluginValidator creates a new plugin validator instance
func NewPluginValidator(cfg *config.Config) *PluginValidator {
	return &PluginValidator{
		cfg: cfg,
	}
}

// GitHubRepository represents a GitHub repository response
type GitHubRepository struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	FullName    string `json:"full_name"`
	Owner       Owner  `json:"owner"`
	Private     bool   `json:"private"`
	HTMLURL     string `json:"html_url"`
	Description string `json:"description"`
	Fork        bool   `json:"fork"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
}

type Owner struct {
	Login string `json:"login"`
	Type  string `json:"type"`
}

// PluginManifest represents a validated plugin manifest
type PluginManifest struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Description  string            `json:"description"`
	Author       string            `json:"author"`
	Repository   string            `json:"repository"`
	License      string            `json:"license"`
	Permissions  []string          `json:"permissions"`
	Dependencies map[string]string `json:"dependencies"`
	Signature    string            `json:"signature"`
	Checksum     string            `json:"checksum"`
}

// ApprovedRepositories defines the whitelist of approved plugin sources
var ApprovedRepositories = map[string]bool{
	"github.com/cloudbox/plugins":          true,
	"github.com/cloudbox/official-plugins": true,
	"github.com/cloudbox/community-plugins": true,
	// Add organization repositories
	"github.com/cloudbox-org/plugins": true,
}

// ValidateGitHubRepository validates if a GitHub repository is approved for plugin installation
func (pv *PluginValidator) ValidateGitHubRepository(repoURL string) (*GitHubRepository, error) {
	// Parse and validate repository URL
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return nil, fmt.Errorf("invalid repository URL: %v", err)
	}

	// Handle both https://github.com/user/repo and github.com/user/repo formats
	var path string
	if parsedURL.Host == "" {
		// Handle github.com/user/repo format (no protocol)
		if !strings.HasPrefix(repoURL, "github.com/") {
			return nil, fmt.Errorf("only GitHub repositories are allowed")
		}
		path = strings.TrimPrefix(repoURL, "github.com")
	} else {
		// Handle https://github.com/user/repo format
		if parsedURL.Host != "github.com" {
			return nil, fmt.Errorf("only GitHub repositories are allowed")
		}
		path = parsedURL.Path
	}

	// Extract owner and repo name
	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) != 2 {
		return nil, fmt.Errorf("invalid GitHub repository format")
	}

	owner := pathParts[0]
	repo := pathParts[1]

	// Validate repository name format
	if !pv.isValidRepositoryName(owner, repo) {
		return nil, fmt.Errorf("invalid repository name format")
	}

	// Check against approved repositories whitelist
	repoKey := fmt.Sprintf("github.com/%s/%s", owner, repo)
	if !ApprovedRepositories[repoKey] {
		return nil, fmt.Errorf("repository not in approved whitelist: %s", repoKey)
	}

	// Fetch repository information from GitHub API
	githubRepo, err := pv.fetchGitHubRepository(owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repository information: %v", err)
	}

	// Additional security checks
	if err := pv.validateRepositoryMetadata(githubRepo); err != nil {
		return nil, fmt.Errorf("repository failed security validation: %v", err)
	}

	return githubRepo, nil
}

// ValidatePluginManifest validates a plugin manifest file for security and integrity
func (pv *PluginValidator) ValidatePluginManifest(manifestContent []byte) (*PluginManifest, error) {
	var manifest PluginManifest
	if err := json.Unmarshal(manifestContent, &manifest); err != nil {
		return nil, fmt.Errorf("invalid manifest JSON: %v", err)
	}

	// Validate required fields
	if err := pv.validateManifestFields(&manifest); err != nil {
		return nil, err
	}

	// Validate permissions
	if err := pv.validatePermissions(manifest.Permissions); err != nil {
		return nil, err
	}

	// Validate dependencies
	if err := pv.validateDependencies(manifest.Dependencies); err != nil {
		return nil, err
	}

	// Verify checksum if provided
	if manifest.Checksum != "" {
		calculatedChecksum := pv.calculateManifestChecksum(manifestContent)
		if manifest.Checksum != calculatedChecksum {
			return nil, fmt.Errorf("manifest checksum verification failed")
		}
	}

	return &manifest, nil
}

// isValidRepositoryName validates repository name format
func (pv *PluginValidator) isValidRepositoryName(owner, repo string) bool {
	// GitHub username/organization name pattern
	ownerPattern := regexp.MustCompile(`^[a-zA-Z0-9]([a-zA-Z0-9\-])*[a-zA-Z0-9]$|^[a-zA-Z0-9]$`)
	// GitHub repository name pattern
	repoPattern := regexp.MustCompile(`^[a-zA-Z0-9._-]+$`)

	return ownerPattern.MatchString(owner) && repoPattern.MatchString(repo)
}

// fetchGitHubRepository fetches repository metadata from GitHub API
func (pv *PluginValidator) fetchGitHubRepository(owner, repo string) (*GitHubRepository, error) {
	apiURL := fmt.Sprintf("https://api.github.com/repos/%s/%s", owner, repo)
	
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, err
	}

	// Add GitHub token if available for higher rate limits
	if pv.cfg.GitHubToken != "" {
		req.Header.Set("Authorization", "token "+pv.cfg.GitHubToken)
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox-Plugin-Validator")

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

	var repoData GitHubRepository
	if err := json.Unmarshal(body, &repoData); err != nil {
		return nil, err
	}

	return &repoData, nil
}

// validateRepositoryMetadata performs additional security checks on repository
func (pv *PluginValidator) validateRepositoryMetadata(repo *GitHubRepository) error {
	// Check if repository is private (may not be suitable for public plugins)
	if repo.Private {
		return fmt.Errorf("private repositories are not allowed for plugin installation")
	}

	// Check if repository is a fork (may indicate unofficial status)
	if repo.Fork {
		return fmt.Errorf("forked repositories require manual approval")
	}

	// Validate repository age (prevent newly created malicious repos)
	createdAt, err := time.Parse(time.RFC3339, repo.CreatedAt)
	if err != nil {
		return fmt.Errorf("invalid repository creation date")
	}

	minAge := 30 * 24 * time.Hour // 30 days
	if time.Since(createdAt) < minAge {
		return fmt.Errorf("repository is too new (minimum age: 30 days)")
	}

	// Check repository activity (last update)
	updatedAt, err := time.Parse(time.RFC3339, repo.UpdatedAt)
	if err != nil {
		return fmt.Errorf("invalid repository update date")
	}

	maxInactivity := 365 * 24 * time.Hour // 1 year
	if time.Since(updatedAt) > maxInactivity {
		return fmt.Errorf("repository appears inactive (last update > 1 year)")
	}

	return nil
}

// validateManifestFields validates required manifest fields
func (pv *PluginValidator) validateManifestFields(manifest *PluginManifest) error {
	if manifest.Name == "" {
		return fmt.Errorf("plugin name is required")
	}

	if len(manifest.Name) > 100 {
		return fmt.Errorf("plugin name too long (max 100 characters)")
	}

	namePattern := regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	if !namePattern.MatchString(manifest.Name) {
		return fmt.Errorf("invalid plugin name format")
	}

	if manifest.Version == "" {
		return fmt.Errorf("plugin version is required")
	}

	// Validate semantic versioning
	versionPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.-]+)?$`)
	if !versionPattern.MatchString(manifest.Version) {
		return fmt.Errorf("invalid version format (use semantic versioning)")
	}

	if manifest.Author == "" {
		return fmt.Errorf("plugin author is required")
	}

	if manifest.Repository == "" {
		return fmt.Errorf("plugin repository is required")
	}

	return nil
}

// validatePermissions validates requested plugin permissions
func (pv *PluginValidator) validatePermissions(permissions []string) error {
	allowedPermissions := map[string]bool{
		"database:read":    true,
		"database:write":   true,
		"storage:read":     true,
		"storage:write":    true,
		"functions:deploy": true,
		"functions:execute": true,
		"webhooks:create":  true,
		"webhooks:manage":  true,
		"projects:read":    true,
		"projects:manage":  true,
		"users:read":       true,
		"users:manage":     true,
		"admin:read":       true,
		// Note: admin:write should require special approval
	}

	dangerousPermissions := []string{
		"admin:write",
		"system:execute",
		"database:admin",
		"storage:admin",
	}

	for _, permission := range permissions {
		if !allowedPermissions[permission] {
			return fmt.Errorf("unknown permission: %s", permission)
		}

		// Check for dangerous permissions that require special approval
		for _, dangerous := range dangerousPermissions {
			if permission == dangerous {
				return fmt.Errorf("permission '%s' requires manual approval", permission)
			}
		}
	}

	// Limit total number of permissions
	if len(permissions) > 10 {
		return fmt.Errorf("too many permissions requested (max 10)")
	}

	return nil
}

// validateDependencies validates plugin dependencies
func (pv *PluginValidator) validateDependencies(dependencies map[string]string) error {
	allowedDependencies := map[string]bool{
		"cloudbox-sdk": true,
		"axios":        true,
		"lodash":       true,
		"moment":       true,
		"uuid":         true,
	}

	for dep, version := range dependencies {
		if !allowedDependencies[dep] {
			return fmt.Errorf("dependency '%s' is not in approved list", dep)
		}

		// Validate version format
		if version == "" {
			return fmt.Errorf("dependency '%s' requires version specification", dep)
		}

		// Basic version validation (can be enhanced)
		versionPattern := regexp.MustCompile(`^[\^~]?\d+\.\d+\.\d+.*$`)
		if !versionPattern.MatchString(version) {
			return fmt.Errorf("invalid version format for dependency '%s': %s", dep, version)
		}
	}

	// Limit number of dependencies
	if len(dependencies) > 20 {
		return fmt.Errorf("too many dependencies (max 20)")
	}

	return nil
}

// calculateManifestChecksum calculates SHA256 checksum of manifest content
func (pv *PluginValidator) calculateManifestChecksum(content []byte) string {
	hash := sha256.Sum256(content)
	return hex.EncodeToString(hash[:])
}

// AddApprovedRepository adds a repository to the approved list (admin function)
func AddApprovedRepository(repoURL string) error {
	// Validate URL format
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %v", err)
	}

	// Handle both https://github.com/user/repo and github.com/user/repo formats
	var path string
	if parsedURL.Host == "" {
		// Handle github.com/user/repo format (no protocol)
		if !strings.HasPrefix(repoURL, "github.com/") {
			return fmt.Errorf("only GitHub repositories are supported")
		}
		path = strings.TrimPrefix(repoURL, "github.com")
	} else {
		// Handle https://github.com/user/repo format
		if parsedURL.Host != "github.com" {
			return fmt.Errorf("only GitHub repositories are supported")
		}
		path = parsedURL.Path
	}

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) != 2 {
		return fmt.Errorf("invalid GitHub repository format")
	}

	repoKey := fmt.Sprintf("github.com/%s/%s", pathParts[0], pathParts[1])
	ApprovedRepositories[repoKey] = true

	return nil
}

// RemoveApprovedRepository removes a repository from the approved list (admin function)
func RemoveApprovedRepository(repoURL string) error {
	parsedURL, err := url.Parse(repoURL)
	if err != nil {
		return fmt.Errorf("invalid repository URL: %v", err)
	}

	// Handle both https://github.com/user/repo and github.com/user/repo formats
	var path string
	if parsedURL.Host == "" {
		// Handle github.com/user/repo format (no protocol)
		if !strings.HasPrefix(repoURL, "github.com/") {
			return fmt.Errorf("only GitHub repositories are supported")
		}
		path = strings.TrimPrefix(repoURL, "github.com")
	} else {
		// Handle https://github.com/user/repo format
		if parsedURL.Host != "github.com" {
			return fmt.Errorf("only GitHub repositories are supported")
		}
		path = parsedURL.Path
	}

	pathParts := strings.Split(strings.Trim(path, "/"), "/")
	if len(pathParts) != 2 {
		return fmt.Errorf("invalid GitHub repository format")
	}

	repoKey := fmt.Sprintf("github.com/%s/%s", pathParts[0], pathParts[1])
	delete(ApprovedRepositories, repoKey)

	return nil
}