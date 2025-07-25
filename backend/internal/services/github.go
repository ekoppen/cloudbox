package services

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cloudbox/backend/internal/models"
	"gorm.io/gorm"
)

// GitHubService handles GitHub API interactions
type GitHubService struct {
	db         *gorm.DB
	httpClient *http.Client
	baseURL    string
}

// GitHubRepositoryInfo represents repository information from GitHub API
type GitHubRepositoryInfo struct {
	ID              int64     `json:"id"`
	NodeID          string    `json:"node_id"`
	Name            string    `json:"name"`
	FullName        string    `json:"full_name"`
	Private         bool      `json:"private"`
	Owner           GitHubUser `json:"owner"`
	Description     *string   `json:"description"`
	CloneURL        string    `json:"clone_url"`
	SSHURL          string    `json:"ssh_url"`
	HTMLURL         string    `json:"html_url"`
	DefaultBranch   string    `json:"default_branch"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	PushedAt        *time.Time `json:"pushed_at"`
	Size            int       `json:"size"`
	StargazersCount int       `json:"stargazers_count"`
	Language        *string   `json:"language"`
	ForksCount      int       `json:"forks_count"`
	Archived        bool      `json:"archived"`
	Disabled        bool      `json:"disabled"`
	OpenIssuesCount int       `json:"open_issues_count"`
	Topics          []string  `json:"topics"`
	Visibility      string    `json:"visibility"`
}

// GitHubUser represents a GitHub user
type GitHubUser struct {
	Login     string `json:"login"`
	ID        int64  `json:"id"`
	NodeID    string `json:"node_id"`
	AvatarURL string `json:"avatar_url"`
	Type      string `json:"type"`
}

// GitHubBranch represents a GitHub branch
type GitHubBranch struct {
	Name      string           `json:"name"`
	Commit    GitHubCommit     `json:"commit"`
	Protected bool             `json:"protected"`
}

// GitHubCommit represents a GitHub commit
type GitHubCommit struct {
	SHA string `json:"sha"`
	URL string `json:"url"`
}

// GitHubCommitInfo represents detailed commit information
type GitHubCommitInfo struct {
	SHA         string                `json:"sha"`
	NodeID      string                `json:"node_id"`
	Commit      GitHubCommitDetails   `json:"commit"`
	URL         string                `json:"url"`
	HTMLURL     string                `json:"html_url"`
	Author      *GitHubUser           `json:"author"`
	Committer   *GitHubUser           `json:"committer"`
	Parents     []GitHubCommit        `json:"parents"`
}

// GitHubCommitDetails represents commit details
type GitHubCommitDetails struct {
	Author      GitHubCommitUser `json:"author"`
	Committer   GitHubCommitUser `json:"committer"`
	Message     string           `json:"message"`
	Tree        GitHubCommit     `json:"tree"`
	URL         string           `json:"url"`
	CommentCount int             `json:"comment_count"`
}

// GitHubCommitUser represents commit author/committer
type GitHubCommitUser struct {
	Name  string    `json:"name"`
	Email string    `json:"email"`
	Date  time.Time `json:"date"`
}

// NewGitHubService creates a new GitHub service
func NewGitHubService(db *gorm.DB) *GitHubService {
	return &GitHubService{
		db: db,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: "https://api.github.com",
	}
}

// SyncRepository syncs repository information from GitHub
func (s *GitHubService) SyncRepository(repository *models.GitHubRepository, accessToken string) error {
	// Extract owner and repo from full name
	parts := strings.Split(repository.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository full name format: %s", repository.FullName)
	}
	owner, repo := parts[0], parts[1]

	// Fetch repository information from GitHub
	repoInfo, err := s.getRepositoryInfo(owner, repo, accessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch repository info: %w", err)
	}

	// Update repository with GitHub information
	updates := map[string]interface{}{
		"name":             repoInfo.Name,
		"full_name":        repoInfo.FullName,
		"clone_url":        repoInfo.CloneURL,
		"is_private":       repoInfo.Private,
		"description":      repoInfo.Description,
		"default_branch":   repoInfo.DefaultBranch,
		"github_id":        repoInfo.ID,
		"last_sync_at":     time.Now(),
		"is_active":        !repoInfo.Archived && !repoInfo.Disabled,
	}

	// Update language if available
	if repoInfo.Language != nil {
		updates["language"] = *repoInfo.Language
	}

	// Update size and stats
	updates["size"] = repoInfo.Size
	updates["stargazers_count"] = repoInfo.StargazersCount
	updates["forks_count"] = repoInfo.ForksCount

	if err := s.db.Model(repository).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update repository: %w", err)
	}

	return nil
}

// GetRepositoryBranches fetches all branches for a repository
func (s *GitHubService) GetRepositoryBranches(repository *models.GitHubRepository, accessToken string) ([]GitHubBranch, error) {
	parts := strings.Split(repository.FullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository full name format: %s", repository.FullName)
	}
	owner, repo := parts[0], parts[1]

	url := fmt.Sprintf("%s/repos/%s/%s/branches", s.baseURL, owner, repo)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token provided
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var branches []GitHubBranch
	if err := json.NewDecoder(resp.Body).Decode(&branches); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return branches, nil
}

// GetLatestCommit fetches the latest commit for a branch
func (s *GitHubService) GetLatestCommit(repository *models.GitHubRepository, branch, accessToken string) (*GitHubCommitInfo, error) {
	parts := strings.Split(repository.FullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository full name format: %s", repository.FullName)
	}
	owner, repo := parts[0], parts[1]

	url := fmt.Sprintf("%s/repos/%s/%s/commits/%s", s.baseURL, owner, repo, branch)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token provided
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var commit GitHubCommitInfo
	if err := json.NewDecoder(resp.Body).Decode(&commit); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &commit, nil
}

// ValidateRepository checks if a repository exists and is accessible
func (s *GitHubService) ValidateRepository(fullName, accessToken string) (*GitHubRepositoryInfo, error) {
	parts := strings.Split(fullName, "/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid repository full name format: %s", fullName)
	}
	owner, repo := parts[0], parts[1]

	return s.getRepositoryInfo(owner, repo, accessToken)
}

// SearchRepositories searches for repositories by query
func (s *GitHubService) SearchRepositories(query, accessToken string, limit int) ([]GitHubRepositoryInfo, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	url := fmt.Sprintf("%s/search/repositories?q=%s&sort=stars&order=desc&per_page=%d", 
		s.baseURL, query, limit)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token provided
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var searchResult struct {
		TotalCount        int                     `json:"total_count"`
		IncompleteResults bool                    `json:"incomplete_results"`
		Items             []GitHubRepositoryInfo  `json:"items"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&searchResult); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return searchResult.Items, nil
}

// CreateWebhook creates a webhook for a repository
func (s *GitHubService) CreateWebhook(repository *models.GitHubRepository, webhookURL, accessToken string) error {
	parts := strings.Split(repository.FullName, "/")
	if len(parts) != 2 {
		return fmt.Errorf("invalid repository full name format: %s", repository.FullName)
	}
	owner, repo := parts[0], parts[1]

	webhookData := map[string]interface{}{
		"name": "web",
		"active": true,
		"events": []string{"push", "pull_request"},
		"config": map[string]interface{}{
			"url":          webhookURL,
			"content_type": "json",
			"secret":       repository.WebhookSecret,
			"insecure_ssl": "0",
		},
	}

	jsonData, err := json.Marshal(webhookData)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook data: %w", err)
	}

	url := fmt.Sprintf("%s/repos/%s/%s/hooks", s.baseURL, owner, repo)
	
	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	return nil
}

// getRepositoryInfo fetches repository information from GitHub API
func (s *GitHubService) getRepositoryInfo(owner, repo, accessToken string) (*GitHubRepositoryInfo, error) {
	url := fmt.Sprintf("%s/repos/%s/%s", s.baseURL, owner, repo)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication if token provided
	if accessToken != "" {
		req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	}
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("repository not found: %s/%s", owner, repo)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var repoInfo GitHubRepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&repoInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &repoInfo, nil
}

// GetUserRepositories fetches repositories for the authenticated user
func (s *GitHubService) GetUserRepositories(accessToken string, visibility string, limit int) ([]GitHubRepositoryInfo, error) {
	if limit <= 0 {
		limit = 30
	}
	if limit > 100 {
		limit = 100
	}

	url := fmt.Sprintf("%s/user/repos?visibility=%s&sort=updated&per_page=%d", 
		s.baseURL, visibility, limit)
	
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("token %s", accessToken))
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "CloudBox/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error (status %d): %s", resp.StatusCode, string(body))
	}

	var repositories []GitHubRepositoryInfo
	if err := json.NewDecoder(resp.Body).Decode(&repositories); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return repositories, nil
}