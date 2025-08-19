package models

import (
	"time"

	"github.com/lib/pq"
	"gorm.io/gorm"
)

// UserRole represents user roles in the system
type UserRole string

const (
	RoleUser       UserRole = "user"        // Regular user
	RoleAdmin      UserRole = "admin"       // Project admin (default)
	RoleSuperAdmin UserRole = "superadmin"  // Super admin (can see all projects)
)

// User represents a CloudBox user
type User struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	Name         string    `json:"name"`
	Role         UserRole  `json:"role" gorm:"type:varchar(20);default:'admin'"`
	IsActive     bool      `json:"is_active" gorm:"default:true"`
	LastLoginAt  *time.Time `json:"last_login_at"`
	
	// Relationships
	Projects     []Project      `json:"projects,omitempty"`
	RefreshTokens []RefreshToken `json:"refresh_tokens,omitempty"`
	OrganizationAdmins []OrganizationAdmin `json:"organization_admins,omitempty"`
}

// RefreshToken represents a refresh token for persistent login
type RefreshToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Token     string    `json:"-" gorm:"uniqueIndex;not null"`
	TokenHash string    `json:"-" gorm:"not null"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`
	IsActive  bool      `json:"is_active" gorm:"default:true"`

	// Session metadata
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`

	// User relation
	UserID uint `json:"user_id" gorm:"not null;index"`
	User   User `json:"user,omitempty"`
}

// Organization represents a group of projects for better organization
type Organization struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Color       string `json:"color" gorm:"default:'#3B82F6'"` // Hex color for UI
	IsActive    bool   `json:"is_active" gorm:"default:true"`

	// Contact information
	Website     string `json:"website"`
	Email       string `json:"email"`
	Phone       string `json:"phone"`
	ContactPerson string `json:"contact_person"`
	
	// Logo and branding
	LogoURL     string `json:"logo_url"`
	LogoFileID  *uint  `json:"logo_file_id,omitempty"` // Reference to uploaded file
	
	// Address information  
	Address     string `json:"address"`
	City        string `json:"city"`
	Country     string `json:"country"`
	PostalCode  string `json:"postal_code"`

	// Owner
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty"`

	// Statistics
	ProjectCount int `json:"project_count" gorm:"default:0"`
	
	// Organization admins (many-to-many relationship)
	OrganizationAdmins []OrganizationAdmin `json:"organization_admins,omitempty"`
}

// OrganizationAdmin represents the many-to-many relationship between users and organizations for admin access
type OrganizationAdmin struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// User relation
	UserID uint `json:"user_id" gorm:"not null;index"`
	User   User `json:"user,omitempty"`

	// Organization relation
	OrganizationID uint         `json:"organization_id" gorm:"not null;index"`
	Organization   Organization `json:"organization,omitempty"`

	// Admin permissions level
	Role string `json:"role" gorm:"default:'admin'"` // admin, owner (future expansion)

	// Status
	IsActive bool `json:"is_active" gorm:"default:true"`

	// Metadata
	AssignedBy   uint       `json:"assigned_by" gorm:"not null"` // ID of superadmin who assigned
	AssignedAt   time.Time  `json:"assigned_at" gorm:"not null"`
	RevokedBy    *uint      `json:"revoked_by,omitempty"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
}

// Project represents a CloudBox project
type Project struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Slug        string `json:"slug" gorm:"not null"`
	Notes       string `json:"notes" gorm:"type:text"` // Project notes
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	
	// Owner
	UserID uint `json:"user_id" gorm:"not null"`
	User   User `json:"user,omitempty"`

	// Organization (required)
	OrganizationID uint         `json:"organization_id" gorm:"index"`
	Organization   Organization `json:"organization,omitempty"`
	
	// Relationships
	APIKeys     []APIKey     `json:"api_keys,omitempty"`
	CORSConfig  *CORSConfig  `json:"cors_config,omitempty"`
	Deployments []Deployment `json:"deployments,omitempty"`
	Backups     []Backup     `json:"backups,omitempty"`
	
	// GitHub configuration per project
	GitHubConfig *ProjectGitHubConfig `json:"github_config,omitempty"`
}

// ProjectGitHubConfig stores GitHub OAuth settings per project
type ProjectGitHubConfig struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;uniqueIndex"`
	Project   Project `json:"project,omitempty"`
	
	// GitHub OAuth configuration
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret" gorm:"type:text"` // Should be encrypted
	IsEnabled    bool   `json:"is_enabled" gorm:"default:false"`
	
	// OAuth URLs
	CallbackURL string `json:"callback_url"`
	
	// Metadata
	CreatedBy uint `json:"created_by" gorm:"not null"`
	UpdatedBy uint `json:"updated_by"`
}

// APIKey represents an API key for a project
type APIKey struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string     `json:"name" gorm:"not null"`
	KeyHash     string     `json:"-" gorm:"uniqueIndex;not null"` // Only store hashed version
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ExpiresAt   *time.Time `json:"expires_at"`
	
	// Permissions
	Permissions pq.StringArray `json:"permissions" gorm:"type:text[]"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`
}

// CORSConfig represents CORS configuration for a project
type CORSConfig struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	AllowedOrigins     pq.StringArray `json:"allowed_origins" gorm:"type:text[]"`
	AllowedMethods     pq.StringArray `json:"allowed_methods" gorm:"type:text[]"`
	AllowedHeaders     pq.StringArray `json:"allowed_headers" gorm:"type:text[]"`
	ExposedHeaders     pq.StringArray `json:"exposed_headers" gorm:"type:text[]"`
	AllowCredentials   bool     `json:"allow_credentials" gorm:"default:false"`
	MaxAge             int      `json:"max_age" gorm:"default:3600"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"uniqueIndex;not null"`
	Project   Project `json:"project,omitempty"`
}

// Collection represents a dynamic data collection (like a table)
type Collection struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Schema      map[string]interface{} `json:"schema" gorm:"type:jsonb;serializer:json"` // JSON schema for validation
	Indexes     []string `json:"indexes" gorm:"type:jsonb;serializer:json"` // Database indexes
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`
	
	// Statistics
	DocumentCount int64     `json:"document_count" gorm:"default:0"`
	LastModified  time.Time `json:"last_modified"`
}

// Document represents a document in a collection
type Document struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID or custom ID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	// Collection info
	CollectionName string `json:"collection_name" gorm:"not null;index"`
	ProjectID      uint   `json:"project_id" gorm:"not null;index"`
	
	// Document data (JSON)
	Data map[string]interface{} `json:"data" gorm:"type:jsonb;serializer:json"`
	
	// Metadata
	Version int    `json:"version" gorm:"default:1"`
	Author  string `json:"author"` // User/API key that created/modified
}

// GitHubRepository represents a connected GitHub repository
type GitHubRepository struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	FullName    string `json:"full_name" gorm:"not null"` // owner/repo
	CloneURL    string `json:"clone_url" gorm:"not null"`
	Branch      string `json:"branch" gorm:"default:'main'"`
	IsPrivate   bool   `json:"is_private" gorm:"default:false"`
	Description string `json:"description"`

	// GitHub webhook
	WebhookID     *int64 `json:"webhook_id"`
	WebhookSecret string `json:"webhook_secret"`

	// SSH Key for private repository access (optional)
	SSHKeyID *uint   `json:"ssh_key_id,omitempty"`
	SSHKey   *SSHKey `json:"ssh_key,omitempty"`
	
	// GitHub OAuth for repository access
	AccessToken      string    `json:"-" gorm:"column:access_token"` // Hidden from JSON for security
	TokenExpiresAt   *time.Time `json:"token_expires_at"`
	RefreshToken     string    `json:"-" gorm:"column:refresh_token"` // Hidden from JSON for security  
	TokenScopes      string    `json:"token_scopes"` // Comma-separated scopes
	AuthorizedAt     *time.Time `json:"authorized_at"`
	AuthorizedBy     string    `json:"authorized_by"` // GitHub username who authorized

	// SDK Configuration
	SDKVersion    string                 `json:"sdk_version"`
	AppPort       int                    `json:"app_port" gorm:"default:3000"`
	BuildCommand  string                 `json:"build_command" gorm:"default:'npm run build'"`
	StartCommand  string                 `json:"start_command" gorm:"default:'npm start'"`
	Environment   map[string]interface{} `json:"environment" gorm:"type:jsonb;serializer:json"`

	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`

	// GitHub specific fields
	GitHubID        int64      `json:"github_id" gorm:"column:git_hub_id"`
	DefaultBranch   string     `json:"default_branch"`
	Language        string     `json:"language"`
	Size            int        `json:"size"`
	StargazersCount int        `json:"stargazers_count"`
	ForksCount      int        `json:"forks_count"`
	
	// Status
	IsActive            bool       `json:"is_active" gorm:"default:true"`
	LastSyncAt          *time.Time `json:"last_sync_at"`
	LastCommitHash      string     `json:"last_commit_hash"`
	PendingCommitHash   string     `json:"pending_commit_hash"`   // New commit available for deployment
	PendingCommitBranch string     `json:"pending_commit_branch"` // Branch of pending commit
	HasPendingUpdate    bool       `json:"has_pending_update" gorm:"default:false"` // Badge indicator

	// Analysis relation
	Analysis *RepositoryAnalysis `json:"analysis,omitempty"`
}

// SystemSetting represents a system-wide configuration setting
type SystemSetting struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	
	// Setting identification
	Key      string `json:"key" gorm:"uniqueIndex;not null"`
	Category string `json:"category" gorm:"not null;default:'general'"`
	
	// Setting values
	Value       string `json:"value"`
	ValueType   string `json:"value_type" gorm:"not null;default:'string'"` // string, boolean, integer, json
	
	// Setting metadata
	Name            string `json:"name" gorm:"not null"`
	Description     string `json:"description"`
	IsSecret        bool   `json:"is_secret" gorm:"default:false"`
	IsRequired      bool   `json:"is_required" gorm:"default:false"`
	DefaultValue    string `json:"default_value"`
	ValidationRules string `json:"validation_rules"` // JSON string
	
	// Setting organization
	SortOrder int  `json:"sort_order" gorm:"default:0"`
	IsActive  bool `json:"is_active" gorm:"default:true"`
}

// InstallOption represents different installation methods for a repository
// PortRequirement represents a port requirement from script analysis
type PortRequirement struct {
	Name        string `json:"name"`         // "Web Server", "API Server", "MongoDB"
	Port        int    `json:"port"`         // 80, 4000, 27017
	Protocol    string `json:"protocol"`     // "tcp", "udp"
	Required    bool   `json:"required"`     // true if absolutely required
	Description string `json:"description"`  // "Web server port"
	Variable    string `json:"variable"`     // Environment variable name, e.g. "WEB_PORT"
}

type InstallOption struct {
	Name         string                 `json:"name"`          // "npm", "yarn", "pnpm", "docker"
	Command      string                 `json:"command"`       // "npm install"
	BuildCommand string                 `json:"build_command"` // "npm run build"
	StartCommand string                 `json:"start_command"` // "npm start"
	DevCommand   string                 `json:"dev_command"`   // "npm run dev"
	Port         int                    `json:"port"`          // Default port
	Environment  map[string]interface{} `json:"environment"`   // Default environment variables
	IsRecommended bool                  `json:"is_recommended"` // Recommended option
	Description  string                 `json:"description"`   // User-friendly description
	PortRequirements []PortRequirement  `json:"port_requirements"` // Required ports for this install option
}

// RepositoryAnalysis represents the detailed analysis of a repository
type RepositoryAnalysis struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Repository relation
	GitHubRepositoryID uint             `json:"github_repository_id" gorm:"column:github_repository_id;uniqueIndex;not null"`
	GitHubRepository   GitHubRepository `json:"github_repository,omitempty"`

	// Analysis metadata
	AnalyzedAt     time.Time `json:"analyzed_at" gorm:"not null"`
	AnalyzedBranch string    `json:"analyzed_branch" gorm:"not null"`
	AnalysisStatus string    `json:"analysis_status" gorm:"default:'completed'"` // pending, completed, failed

	// Project detection
	ProjectType    string   `json:"project_type"`    // react, vue, angular, next, nuxt, etc.
	Framework      string   `json:"framework"`       // vite, webpack, create-react-app, etc.
	Language       string   `json:"language"`        // javascript, typescript, python, go, etc.
	PackageManager string   `json:"package_manager"` // npm, yarn, pnpm, pip, go mod, etc.
	
	// Build configuration
	BuildCommand   string `json:"build_command"`
	StartCommand   string `json:"start_command"`
	DevCommand     string `json:"dev_command"`
	InstallCommand string `json:"install_command"`
	TestCommand    string `json:"test_command"`
	
	// Runtime configuration
	Port        int                    `json:"port"`
	Environment map[string]interface{} `json:"environment" gorm:"type:jsonb;serializer:json"`
	
	// Docker support
	HasDocker     bool   `json:"has_docker"`
	DockerCommand string `json:"docker_command"`
	DockerPort    int    `json:"docker_port"`
	
	// Dependencies and features
	Dependencies    []string `json:"dependencies" gorm:"type:jsonb;serializer:json"`     // Main dependencies found
	DevDependencies []string `json:"dev_dependencies" gorm:"type:jsonb;serializer:json"` // Dev dependencies found
	Scripts         []string `json:"scripts" gorm:"type:jsonb;serializer:json"`          // Available npm scripts
	
	// File structure
	ImportantFiles  []string `json:"important_files" gorm:"type:jsonb;serializer:json"`  // package.json, Dockerfile, etc.
	ConfigFiles     []string `json:"config_files" gorm:"type:jsonb;serializer:json"`     // vite.config.js, next.config.js, etc.
	EnvironmentFiles []string `json:"environment_files" gorm:"type:jsonb;serializer:json"` // .env.example, .env.local, etc.
	
	// Installation options
	InstallOptions []InstallOption `json:"install_options" gorm:"type:jsonb;serializer:json"`
	
	// Analysis insights
	Insights      []string `json:"insights" gorm:"type:jsonb;serializer:json"`      // Helpful suggestions
	Warnings      []string `json:"warnings" gorm:"type:jsonb;serializer:json"`      // Potential issues
	Requirements  []string `json:"requirements" gorm:"type:jsonb;serializer:json"`  // System requirements
	
	// Performance metrics
	EstimatedBuildTime int64 `json:"estimated_build_time"` // seconds
	EstimatedSize      int64 `json:"estimated_size"`       // bytes
	Complexity         int   `json:"complexity"`           // 1-10 scale
	
	// Analysis errors
	AnalysisErrors []string `json:"analysis_errors" gorm:"type:jsonb;serializer:json"` // Errors during analysis
	
	// Project relation for easier querying
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`
}

// WebServer represents a target deployment server
type WebServer struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	Hostname    string `json:"hostname" gorm:"not null"`
	Port        int    `json:"port" gorm:"default:22"`
	Username    string `json:"username" gorm:"not null"`
	Description string `json:"description"`

	// Server configuration
	ServerType    string `json:"server_type" gorm:"default:'vps'"` // vps, dedicated, cloud
	OS            string `json:"os" gorm:"default:'ubuntu'"`
	DockerEnabled bool   `json:"docker_enabled" gorm:"default:true"`
	NginxEnabled  bool   `json:"nginx_enabled" gorm:"default:true"`

	// Deployment paths
	DeployPath    string `json:"deploy_path" gorm:"default:'~/deploys'"`
	BackupPath    string `json:"backup_path" gorm:"default:'/var/backups'"`
	LogPath       string `json:"log_path" gorm:"default:'/var/log/deployments'"`

	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`

	// SSH Key relation
	SSHKeyID uint   `json:"ssh_key_id" gorm:"not null"`
	SSHKey   SSHKey `json:"ssh_key,omitempty"`

	// Status
	IsActive       bool       `json:"is_active" gorm:"default:true"`
	LastConnectedAt *time.Time `json:"last_connected_at"`
	ConnectionStatus string    `json:"connection_status" gorm:"default:'unknown'"` // connected, disconnected, error
}

// SSHKey represents SSH keys for server access
type SSHKey struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	
	// SSH Key data
	PublicKey    string `json:"public_key" gorm:"not null"`
	PrivateKey   string `json:"-" gorm:"not null"` // Encrypted storage
	Fingerprint  string `json:"fingerprint" gorm:"not null"`
	KeyType      string `json:"key_type" gorm:"default:'rsa'"` // rsa, ed25519
	KeySize      int    `json:"key_size" gorm:"default:2048"`

	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`

	// Usage tracking
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	LastUsedAt  *time.Time `json:"last_used_at"`
	ServerCount int        `json:"server_count" gorm:"default:0"`
}

// Deployment represents a deployment from repository to server
type Deployment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Version     string `json:"version" gorm:"not null"`
	
	// Deployment configuration
	Domain        string                 `json:"domain"`
	Subdomain     string                 `json:"subdomain"`
	Port          int                    `json:"port"`
	DeployPath    string                 `json:"deploy_path"`    // Override deployment path (default: ~/deploys/{name})
	Environment   map[string]interface{} `json:"environment" gorm:"type:jsonb;serializer:json"`
	BuildCommand  string                 `json:"build_command"`
	StartCommand  string                 `json:"start_command"`
	PortConfiguration map[string]int     `json:"port_configuration" gorm:"type:jsonb;serializer:json"` // Port mappings: variable -> port
	
	// Deployment status
	Status        string     `json:"status" gorm:"default:'pending'"` // pending, building, deploying, deployed, failed, stopped
	DeployedAt    *time.Time `json:"deployed_at"`
	BuildLogs     string     `json:"build_logs"`
	DeployLogs    string     `json:"deploy_logs"`
	ErrorLogs     string     `json:"error_logs"`

	// Git information
	CommitHash    string `json:"commit_hash"`
	CommitMessage string `json:"commit_message"`
	CommitAuthor  string `json:"commit_author"`
	Branch        string `json:"branch"`

	// Performance metrics
	BuildTime   *int64 `json:"build_time"`   // milliseconds
	DeployTime  *int64 `json:"deploy_time"`  // milliseconds
	FileCount   int64  `json:"file_count"`
	TotalSize   int64  `json:"total_size"`

	// Relations
	ProjectID          uint              `json:"project_id" gorm:"not null"`
	Project            Project           `json:"project,omitempty"`
	GitHubRepositoryID uint              `json:"github_repository_id" gorm:"not null"`
	GitHubRepository   GitHubRepository  `json:"github_repository,omitempty"`
	WebServerID        uint              `json:"web_server_id" gorm:"not null"`
	WebServer          WebServer         `json:"web_server,omitempty"`

	// Auto-deployment settings
	IsAutoDeployEnabled bool `json:"is_auto_deploy_enabled" gorm:"default:false"`
	TriggerBranch       string `json:"trigger_branch" gorm:"default:'main'"`
}

// Backup represents a project backup
type Backup struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
	
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Type        string `json:"type" gorm:"not null"` // manual, automatic
	Status      string `json:"status" gorm:"default:pending"` // pending, creating, completed, failed
	
	// Backup metadata
	Size         int64     `json:"size"`
	FilePath     string    `json:"file_path"`
	Checksum     string    `json:"checksum"`
	CompletedAt  *time.Time `json:"completed_at"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`
}

// Bucket represents a file storage bucket
type Bucket struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	Name        string   `json:"name" gorm:"not null"`
	Description string   `json:"description"`
	MaxFileSize int64    `json:"max_file_size" gorm:"default:52428800"` // 50MB default
	AllowedTypes []string `json:"allowed_types" gorm:"type:jsonb;serializer:json"` // MIME types
	IsPublic    bool     `json:"is_public" gorm:"default:false"`

	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`

	// Statistics
	FileCount int64     `json:"file_count" gorm:"default:0"`
	TotalSize int64     `json:"total_size" gorm:"default:0"`
	LastModified time.Time `json:"last_modified"`
}

// File represents an uploaded file
type File struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// File metadata
	OriginalName string `json:"original_name" gorm:"not null"`
	FileName     string `json:"file_name" gorm:"not null"` // Stored filename
	FilePath     string `json:"file_path" gorm:"not null"` // Full path on disk
	MimeType     string `json:"mime_type" gorm:"not null"`
	Size         int64  `json:"size" gorm:"not null"`
	Checksum     string `json:"checksum"` // MD5 or SHA256

	// Storage info
	BucketName string `json:"bucket_name" gorm:"not null;index"`
	FolderPath string `json:"folder_path" gorm:"index"` // Path within bucket (empty for root)
	ProjectID  uint   `json:"project_id" gorm:"not null;index"`

	// Access control
	IsPublic bool   `json:"is_public" gorm:"default:false"`
	Author   string `json:"author"` // User/API key that uploaded

	// URLs
	PublicURL  string `json:"public_url,omitempty"`
	PrivateURL string `json:"private_url,omitempty"`
}

// AppUser represents an application user (different from CloudBox admin users)
type AppUser struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// User credentials
	Email        string `json:"email" gorm:"not null;index"`
	PasswordHash string `json:"-" gorm:"not null"`
	Name         string `json:"name"`
	Username     string `json:"username" gorm:"index"`

	// User metadata
	ProfileData map[string]interface{} `json:"profile_data" gorm:"type:jsonb;serializer:json"`
	Preferences map[string]interface{} `json:"preferences" gorm:"type:jsonb;serializer:json"`
	
	// Status
	IsActive        bool       `json:"is_active" gorm:"default:true"`
	Status          string     `json:"status" gorm:"-"` // Computed field, not stored in DB
	IsEmailVerified bool       `json:"is_email_verified" gorm:"default:false"`
	LastLoginAt     *time.Time `json:"last_login_at"`
	LastSeenAt      *time.Time `json:"last_seen_at"`

	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`

	// Security
	LoginAttempts   int        `json:"login_attempts" gorm:"default:0"`
	LockedUntil     *time.Time `json:"locked_until"`
	PasswordResetToken string  `json:"-"`
	PasswordResetExpires *time.Time `json:"-"`
	EmailVerificationToken string `json:"-"`
}

// Function represents a serverless function
type Function struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Function identification
	Name        string `json:"name" gorm:"not null;uniqueIndex:idx_project_function_name"`
	Description string `json:"description"`
	
	// Function code
	Runtime     string `json:"runtime" gorm:"not null;default:'nodejs18'"`     // nodejs18, python3.9, go1.19
	Language    string `json:"language" gorm:"not null;default:'javascript'"` // javascript, python, go
	Code        string `json:"code" gorm:"type:text;not null"`               // The function code
	EntryPoint  string `json:"entry_point" gorm:"default:'index.handler'"`    // Entry point for the function
	
	// Configuration
	Timeout        int                    `json:"timeout" gorm:"default:30"`       // seconds
	Memory         int                    `json:"memory" gorm:"default:128"`       // MB
	Environment    map[string]interface{} `json:"environment" gorm:"type:jsonb;serializer:json"`
	Commands       []string               `json:"commands" gorm:"type:jsonb;serializer:json"` // Build commands
	Dependencies   map[string]interface{} `json:"dependencies" gorm:"type:jsonb;serializer:json"`
	
	// Status and deployment
	Status         string     `json:"status" gorm:"default:'draft'"`        // draft, building, deployed, error
	Version        int        `json:"version" gorm:"default:1"`
	LastDeployedAt *time.Time `json:"last_deployed_at"`
	
	// Runtime info
	BuildLogs      string `json:"build_logs" gorm:"type:text"`
	DeploymentLogs string `json:"deployment_logs" gorm:"type:text"`
	
	// Relationships
	ProjectID uint    `json:"project_id" gorm:"not null;uniqueIndex:idx_project_function_name"`
	Project   Project `json:"project,omitempty"`
}

// TemplateDeployment represents a deployment created from a template
type TemplateDeployment struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Template information
	TemplateName string                 `json:"template_name" gorm:"not null"`
	Variables    map[string]interface{} `json:"variables" gorm:"type:jsonb"`
	
	// Status and relationships
	Status    string `json:"status" gorm:"type:varchar(50);default:'created'"`
	ProjectID uint   `json:"project_id" gorm:"not null"`
	Project   Project `json:"project,omitempty"`
	
	// Related resources
	GitHubRepositoryID uint             `json:"github_repository_id" gorm:"not null"`
	GitHubRepository   GitHubRepository `json:"github_repository,omitempty"`
	DeploymentID       uint             `json:"deployment_id" gorm:"not null"`
	Deployment         Deployment       `json:"deployment,omitempty"`
}

// CloudBoxCompatibility represents CloudBox SDK compatibility check results
type CloudBoxCompatibility struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Repository information
	GitHubRepositoryID uint             `json:"github_repository_id" gorm:"uniqueIndex;not null"`
	GitHubRepository   GitHubRepository `json:"github_repository,omitempty"`
	
	// Compatibility check results
	IsCompatible     bool     `json:"is_compatible"`
	HasCloudBoxSDK   bool     `json:"has_cloudbox_sdk"`
	SDKVersion       string   `json:"sdk_version"`
	PackageManager   string   `json:"package_manager"` // npm, yarn, pnpm
	FrameworkType    string   `json:"framework_type"`  // react, vue, nextjs, etc.
	
	// CloudBox configuration detected
	HasCloudBoxConfig bool                   `json:"has_cloudbox_config"`
	ConfigFile        string                 `json:"config_file"`
	DetectedConfig    map[string]interface{} `json:"detected_config" gorm:"type:jsonb"`
	
	// Environment variables detected
	EnvVariables     pq.StringArray         `json:"env_variables" gorm:"type:text[]"`
	RequiredEnvVars  pq.StringArray         `json:"required_env_vars" gorm:"type:text[]"`
	
	// Installation requirements
	InstallCommands  pq.StringArray         `json:"install_commands" gorm:"type:text[]"`
	BuildCommands    pq.StringArray         `json:"build_commands" gorm:"type:text[]"`
	StartCommands    pq.StringArray         `json:"start_commands" gorm:"type:text[]"`
	
	// Compatibility issues
	Issues           []string               `json:"issues" gorm:"type:jsonb"`
	Recommendations  []string               `json:"recommendations" gorm:"type:jsonb"`
	
	// Check metadata
	CheckedAt        time.Time              `json:"checked_at"`
	CheckVersion     string                 `json:"check_version" gorm:"default:'1.0.0'"`
	ErrorMessage   string `json:"error_message"`
	
	// Function URL and access
	FunctionURL string `json:"function_url"`
	IsActive    bool   `json:"is_active" gorm:"default:true"`
	IsPublic    bool   `json:"is_public" gorm:"default:false"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;uniqueIndex:idx_project_function_name"`
	Project   Project `json:"project,omitempty"`
}

// FunctionExecution represents a function execution/invocation
type FunctionExecution struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Execution details
	FunctionID     uint                   `json:"function_id" gorm:"not null;index"`
	Function       Function               `json:"function,omitempty"`
	ExecutionID    string                 `json:"execution_id" gorm:"not null;uniqueIndex"` // UUID for tracking
	
	// Request/response data
	RequestData    map[string]interface{} `json:"request_data" gorm:"type:jsonb;serializer:json"`
	ResponseData   map[string]interface{} `json:"response_data" gorm:"type:jsonb;serializer:json"`
	Headers        map[string]interface{} `json:"headers" gorm:"type:jsonb;serializer:json"`
	Method         string                 `json:"method" gorm:"not null"`
	Path           string                 `json:"path"`
	
	// Execution results
	Status         string    `json:"status" gorm:"not null"`        // success, error, timeout
	StatusCode     int       `json:"status_code" gorm:"default:200"`
	ExecutionTime  int64     `json:"execution_time"`                // milliseconds
	MemoryUsage    int64     `json:"memory_usage"`                  // bytes
	StartedAt      time.Time `json:"started_at"`
	CompletedAt    *time.Time `json:"completed_at"`
	
	// Logs and errors
	Logs         string `json:"logs" gorm:"type:text"`
	ErrorMessage string `json:"error_message"`
	
	// Metadata
	UserAgent    string `json:"user_agent"`
	ClientIP     string `json:"client_ip"`
	Source       string `json:"source" gorm:"default:'http'"` // http, webhook, cron, manual
	
	// Project relation  
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`
}

// FunctionDomain represents custom domains for functions
type FunctionDomain struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Domain info
	Domain      string `json:"domain" gorm:"not null;uniqueIndex"`
	IsVerified  bool   `json:"is_verified" gorm:"default:false"`
	Certificate string `json:"certificate"` // SSL certificate
	
	// Function mapping
	FunctionID uint     `json:"function_id" gorm:"not null;index"`
	Function   Function `json:"function,omitempty"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`
}

// AppSession represents a user session
type AppSession struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Session info
	UserID    string    `json:"user_id" gorm:"not null;index"`
	Token     string    `json:"-" gorm:"not null;uniqueIndex"`
	ExpiresAt time.Time `json:"expires_at" gorm:"not null"`

	// Session metadata
	IPAddress string                 `json:"ip_address"`
	UserAgent string                 `json:"user_agent"`
	DeviceInfo map[string]interface{} `json:"device_info" gorm:"type:jsonb;serializer:json"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`

	// Status
	IsActive     bool       `json:"is_active" gorm:"default:true"`
	LastActivity *time.Time `json:"last_activity"`
}

// Channel represents a messaging channel
type Channel struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Channel info
	Name        string `json:"name" gorm:"not null"`
	Description string `json:"description"`
	Type        string `json:"type" gorm:"not null;default:'public'"` // public, private, direct
	Topic       string `json:"topic"`

	// Channel settings
	IsActive        bool                   `json:"is_active" gorm:"default:true"`
	MaxMembers      int                    `json:"max_members" gorm:"default:0"` // 0 = unlimited
	Settings        map[string]interface{} `json:"settings" gorm:"type:jsonb;serializer:json"`
	
	// Project relation
	ProjectID uint    `json:"project_id" gorm:"not null;index"`
	Project   Project `json:"project,omitempty"`

	// Creator
	CreatedBy string `json:"created_by"` // AppUser ID
	
	// Statistics
	MemberCount  int       `json:"member_count" gorm:"default:0"`
	MessageCount int64     `json:"message_count" gorm:"default:0"`
	LastActivity time.Time `json:"last_activity"`
}

// ChannelMember represents channel membership
type ChannelMember struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Membership info
	ChannelID string `json:"channel_id" gorm:"not null;index"`
	UserID    string `json:"user_id" gorm:"not null;index"`
	Role      string `json:"role" gorm:"default:'member'"` // owner, admin, member
	
	// Project relation
	ProjectID uint `json:"project_id" gorm:"not null;index"`

	// Status
	IsActive    bool       `json:"is_active" gorm:"default:true"`
	JoinedAt    time.Time  `json:"joined_at"`
	LastReadAt  *time.Time `json:"last_read_at"`
	IsMuted     bool       `json:"is_muted" gorm:"default:false"`
	
	// Permissions
	CanRead    bool `json:"can_read" gorm:"default:true"`
	CanWrite   bool `json:"can_write" gorm:"default:true"`
	CanInvite  bool `json:"can_invite" gorm:"default:false"`
	CanModerate bool `json:"can_moderate" gorm:"default:false"`
}

// Message represents a message in a channel
type Message struct {
	ID        string         `json:"id" gorm:"primaryKey;type:varchar(255)"` // UUID
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Message content
	Content  string                 `json:"content" gorm:"not null"`
	Type     string                 `json:"type" gorm:"default:'text'"` // text, image, file, system
	Metadata map[string]interface{} `json:"metadata" gorm:"type:jsonb;serializer:json"`

	// References
	ChannelID string `json:"channel_id" gorm:"not null;index"`
	UserID    string `json:"user_id" gorm:"not null;index"`
	
	// Thread support
	ParentID   *string `json:"parent_id" gorm:"index"` // For replies
	ThreadID   *string `json:"thread_id" gorm:"index"` // Thread identifier
	ReplyCount int     `json:"reply_count" gorm:"default:0"`

	// Project relation
	ProjectID uint `json:"project_id" gorm:"not null;index"`

	// Status
	IsEdited   bool       `json:"is_edited" gorm:"default:false"`
	EditedAt   *time.Time `json:"edited_at"`
	IsDeleted  bool       `json:"is_deleted" gorm:"default:false"`
	MessageDeletedAt  *time.Time `json:"message_deleted_at"`

	// Reactions and interactions
	ReactionCount int `json:"reaction_count" gorm:"default:0"`
}

// MessageReaction represents reactions to messages
type MessageReaction struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Reaction info
	MessageID string `json:"message_id" gorm:"not null;index"`
	UserID    string `json:"user_id" gorm:"not null;index"`
	Emoji     string `json:"emoji" gorm:"not null"` // emoji unicode or :name:
	
	// Project relation
	ProjectID uint `json:"project_id" gorm:"not null;index"`
}

// MessageRead represents read receipts
type MessageRead struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Read receipt info
	MessageID string    `json:"message_id" gorm:"not null;index"`
	UserID    string    `json:"user_id" gorm:"not null;index"`
	ReadAt    time.Time `json:"read_at" gorm:"not null"`
	
	// Project relation
	ProjectID uint `json:"project_id" gorm:"not null;index"`
}

// AuditLogAction represents the type of action performed
type AuditLogAction string

const (
	AuditActionProjectCreate AuditLogAction = "project.create"
	AuditActionProjectUpdate AuditLogAction = "project.update"
	AuditActionProjectDelete AuditLogAction = "project.delete"
	AuditActionOrgCreate     AuditLogAction = "organization.create"
	AuditActionOrgUpdate     AuditLogAction = "organization.update"
	AuditActionOrgDelete     AuditLogAction = "organization.delete"
	AuditActionUserCreate    AuditLogAction = "user.create"
	AuditActionUserUpdate    AuditLogAction = "user.update"
	AuditActionUserDelete    AuditLogAction = "user.delete"
	AuditActionLogin         AuditLogAction = "auth.login"
	AuditActionLogout        AuditLogAction = "auth.logout"
	AuditActionAPIKeyCreate  AuditLogAction = "apikey.create"
	AuditActionAPIKeyDelete  AuditLogAction = "apikey.delete"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time      `json:"created_at" gorm:"index"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Action details
	Action      AuditLogAction `json:"action" gorm:"not null;index"`
	Resource    string         `json:"resource" gorm:"not null"` // e.g., "project", "user"
	ResourceID  string         `json:"resource_id" gorm:"index"` // ID of the affected resource
	Description string         `json:"description"`              // Human-readable description

	// Actor (who performed the action)
	ActorID   uint   `json:"actor_id" gorm:"not null;index"`
	ActorName string `json:"actor_name" gorm:"not null"`
	ActorRole string `json:"actor_role" gorm:"not null"`

	// Context
	IPAddress string `json:"ip_address"`
	UserAgent string `json:"user_agent"`
	Method    string `json:"method"` // HTTP method
	Path      string `json:"path"`   // Request path

	// Additional data (JSON)
	Metadata string `json:"metadata,omitempty"` // JSON string for additional context

	// Project context (if applicable)
	ProjectID *uint `json:"project_id,omitempty" gorm:"index"`

	// Success/failure
	Success   bool   `json:"success" gorm:"default:true"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

// PluginRegistry represents a plugin in the registry
type PluginRegistry struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin identification
	Name         string `json:"name" gorm:"uniqueIndex;not null"`
	Version      string `json:"version" gorm:"not null"`
	Description  string `json:"description"`
	Author       string `json:"author" gorm:"not null"`
	Repository   string `json:"repository" gorm:"not null"`
	License      string `json:"license"`
	Type         string `json:"type" gorm:"default:'dashboard-plugin'"`
	MainFile     string `json:"main_file"`

	// Security and validation
	Checksum    string `json:"checksum"`
	Signature   string `json:"signature"`
	IsVerified  bool   `json:"is_verified" gorm:"default:false"`
	IsApproved  bool   `json:"is_approved" gorm:"default:false"`

	// Metadata
	Permissions      pq.StringArray         `json:"permissions" gorm:"type:text[]"`
	Dependencies     map[string]interface{} `json:"dependencies" gorm:"type:jsonb;serializer:json"`
	UIConfig         map[string]interface{} `json:"ui_config" gorm:"type:jsonb;serializer:json"`

	// Status and lifecycle
	Status        string    `json:"status" gorm:"default:'available'"`
	DownloadCount int       `json:"download_count" gorm:"default:0"`
	InstallCount  int       `json:"install_count" gorm:"default:0"`
	PublishedAt   *time.Time `json:"published_at"`
	DeprecatedAt  *time.Time `json:"deprecated_at"`

	// Registry metadata
	RegistrySource   string                 `json:"registry_source" gorm:"default:'github'"`
	SourceMetadata   map[string]interface{} `json:"source_metadata" gorm:"type:jsonb;serializer:json"`
}

// TableName overrides the default table name for PluginRegistry
func (PluginRegistry) TableName() string {
	return "plugin_registry"
}

// PluginInstallation represents an installed plugin in a project
type PluginInstallation struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin information
	PluginName    string `json:"plugin_name" gorm:"not null"`
	PluginVersion string `json:"plugin_version" gorm:"not null"`
	ProjectID     uint   `json:"project_id" gorm:"not null"`

	// Installation state
	Status           string `json:"status" gorm:"default:'disabled'"`
	InstallationPath string `json:"installation_path"`

	// Installation metadata
	InstalledBy      uint       `json:"installed_by" gorm:"not null"`
	InstalledAt      time.Time  `json:"installed_at"`
	LastEnabledAt    *time.Time `json:"last_enabled_at"`
	LastDisabledAt   *time.Time `json:"last_disabled_at"`

	// Configuration
	Config      map[string]interface{} `json:"config" gorm:"type:jsonb;serializer:json"`
	Environment map[string]interface{} `json:"environment" gorm:"type:jsonb;serializer:json"`

	// Error tracking
	ErrorMessage  string     `json:"error_message"`
	LastErrorAt   *time.Time `json:"last_error_at"`

	// Relationships
	Project Project `json:"project,omitempty"`
}

// PluginState represents real-time plugin state
type PluginState struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin identification
	PluginName string `json:"plugin_name" gorm:"not null"`
	ProjectID  uint   `json:"project_id" gorm:"not null"`

	// Current state
	CurrentStatus string `json:"current_status" gorm:"default:'disabled'"`
	ProcessID     *int   `json:"process_id"`
	Port          *int   `json:"port"`

	// Health monitoring
	LastHealthCheck *time.Time             `json:"last_health_check"`
	HealthStatus    string                 `json:"health_status" gorm:"default:'unknown'"`
	HealthDetails   map[string]interface{} `json:"health_details" gorm:"type:jsonb;serializer:json"`

	// Performance metrics
	CPUUsage      *float64 `json:"cpu_usage"`
	MemoryUsage   *int64   `json:"memory_usage"`
	UptimeSeconds *int     `json:"uptime_seconds"`

	// State transitions
	StateChangedAt time.Time `json:"state_changed_at"`
	StateChangedBy *uint     `json:"state_changed_by"`

	// Relationships
	Project Project `json:"project,omitempty"`
}

// ApprovedRepository represents an approved plugin repository
type ApprovedRepository struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Repository information
	RepositoryURL   string `json:"repository_url" gorm:"uniqueIndex;not null"`
	RepositoryOwner string `json:"repository_owner" gorm:"not null"`
	RepositoryName  string `json:"repository_name" gorm:"not null"`

	// Approval metadata
	ApprovedBy     uint   `json:"approved_by" gorm:"not null"`
	ApprovedAt     time.Time `json:"approved_at"`
	ApprovalReason string `json:"approval_reason"`

	// Repository metadata
	RepositoryType string `json:"repository_type" gorm:"default:'github'"`
	IsOfficial     bool   `json:"is_official" gorm:"default:false"`
	TrustLevel     int    `json:"trust_level" gorm:"default:1"`

	// Status
	IsActive           bool       `json:"is_active" gorm:"default:true"`
	LastValidatedAt    *time.Time `json:"last_validated_at"`
	ValidationStatus   string     `json:"validation_status" gorm:"default:'pending'"`

	// Security tracking
	SecurityIssues     []string   `json:"security_issues" gorm:"type:jsonb;serializer:json"`
	LastSecurityScan   *time.Time `json:"last_security_scan"`
}

// PluginDownload represents a plugin download/installation attempt
type PluginDownload struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin information
	PluginName    string `json:"plugin_name" gorm:"not null"`
	PluginVersion string `json:"plugin_version" gorm:"not null"`
	ProjectID     uint   `json:"project_id" gorm:"not null"`
	UserID        uint   `json:"user_id" gorm:"not null"`

	// Download details
	DownloadSource string `json:"download_source" gorm:"not null"`
	DownloadStatus string `json:"download_status" gorm:"default:'started'"`

	// File information
	FileSize           *int64 `json:"file_size"`
	DownloadTimeMs     *int   `json:"download_time_ms"`
	ChecksumVerified   bool   `json:"checksum_verified" gorm:"default:false"`
	SignatureVerified  bool   `json:"signature_verified" gorm:"default:false"`

	// Error tracking
	ErrorMessage string `json:"error_message"`
	ErrorCode    string `json:"error_code"`

	// Network information
	ClientIP  string `json:"client_ip"`
	UserAgent string `json:"user_agent"`

	// Timestamps
	StartedAt   time.Time  `json:"started_at"`
	CompletedAt *time.Time `json:"completed_at"`
	FailedAt    *time.Time `json:"failed_at"`

	// Relationships
	Project Project `json:"project,omitempty"`
	User    User    `json:"user,omitempty"`
}

// PluginAuditLog represents security audit trail for plugin operations (from plugins.go)
type PluginAuditLog struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	Action      string    `json:"action"`      // enable, disable, install, uninstall
	PluginName  string    `json:"plugin_name"`
	OldStatus   string    `json:"old_status,omitempty"`
	NewStatus   string    `json:"new_status,omitempty"`
	IPAddress   string    `json:"ip_address"`
	UserAgent   string    `json:"user_agent"`
	Success     bool      `json:"success"`
	ErrorMsg    string    `json:"error_msg,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
}

// PluginSubmission represents a plugin submission for marketplace approval
type PluginSubmission struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin information
	PluginName    string                 `json:"plugin_name" gorm:"not null"`
	Repository    string                 `json:"repository" gorm:"not null"`
	Version       string                 `json:"version" gorm:"not null"`
	Description   string                 `json:"description" gorm:"not null"`
	Author        string                 `json:"author" gorm:"not null"`
	Category      string                 `json:"category" gorm:"not null"`
	Tags          pq.StringArray         `json:"tags" gorm:"type:text[]"`
	License       string                 `json:"license" gorm:"not null"`
	Website       string                 `json:"website"`
	SupportEmail  string                 `json:"support_email" gorm:"not null"`
	Screenshots   pq.StringArray         `json:"screenshots" gorm:"type:text[]"`
	DemoURL       string                 `json:"demo_url"`
	Permissions   pq.StringArray         `json:"permissions" gorm:"type:text[]"`
	Dependencies  map[string]interface{} `json:"dependencies" gorm:"type:jsonb;serializer:json"`
	Configuration map[string]interface{} `json:"configuration" gorm:"type:jsonb;serializer:json"`

	// Submission status
	Status         string `json:"status" gorm:"default:'submitted'"` // submitted, reviewing, approved, rejected
	ReviewComments string `json:"review_comments"`
	ReviewScore    int    `json:"review_score"`

	// Submission metadata
	SubmittedBy uint      `json:"submitted_by" gorm:"not null"`
	SubmittedAt time.Time `json:"submitted_at"`
	ReviewedBy  uint      `json:"reviewed_by"`
	ReviewedAt  time.Time `json:"reviewed_at"`

	// Relationships
	Submitter User `json:"submitter,omitempty" gorm:"foreignKey:SubmittedBy"`
	Reviewer  User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// RepositoryApprovalRequest represents a request to approve a repository for plugin submissions
type RepositoryApprovalRequest struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Repository information
	RepositoryURL    string `json:"repository_url" gorm:"not null"`
	OrganizationName string `json:"organization_name" gorm:"not null"`
	ContactEmail     string `json:"contact_email" gorm:"not null"`
	SecurityLevel    string `json:"security_level"`
	AutoApprove      bool   `json:"auto_approve"`
	Verified         bool   `json:"verified"`
	Reason           string `json:"reason" gorm:"not null"`

	// Request status
	Status         string `json:"status" gorm:"default:'pending'"` // pending, approved, rejected
	ReviewComments string `json:"review_comments"`

	// Request metadata
	RequestedBy uint      `json:"requested_by" gorm:"not null"`
	RequestedAt time.Time `json:"requested_at"`
	ReviewedBy  uint      `json:"reviewed_by"`
	ReviewedAt  time.Time `json:"reviewed_at"`

	// Relationships
	Requester User `json:"requester,omitempty" gorm:"foreignKey:RequestedBy"`
	Reviewer  User `json:"reviewer,omitempty" gorm:"foreignKey:ReviewedBy"`
}

// PluginMarketplace represents a plugin in the marketplace
type PluginMarketplace struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`

	// Plugin information - matching plugin_registry table structure
	Name         string                 `json:"plugin_name" gorm:"column:name;uniqueIndex;not null"`
	Repository   string                 `json:"repository" gorm:"not null"`
	Version      string                 `json:"version" gorm:"not null"`
	Description  string                 `json:"description"`
	Author       string                 `json:"author" gorm:"not null"`
	License      string                 `json:"license"`
	Type         string                 `json:"type" gorm:"default:'dashboard-plugin'"`
	MainFile     string                 `json:"main_file"`
	Checksum     string                 `json:"checksum"`
	Signature    string                 `json:"signature"`
	Permissions  pq.StringArray         `json:"permissions" gorm:"type:text[]"`
	Dependencies map[string]interface{} `json:"dependencies" gorm:"type:jsonb;serializer:json"`
	UIConfig     map[string]interface{} `json:"ui_config" gorm:"type:jsonb;serializer:json"`
	
	// Status and metadata - matching plugin_registry table
	IsVerified      bool   `json:"verified" gorm:"column:is_verified;default:false"`
	IsApproved      bool   `json:"approved" gorm:"column:is_approved;default:false"`
	Status          string `json:"status" gorm:"default:'available'"`
	DownloadCount   int    `json:"download_count" gorm:"default:0"`
	InstallCount    int    `json:"install_count" gorm:"default:0"`
	RegistrySource  string `json:"registry_source" gorm:"default:'github'"`
	SourceMetadata  map[string]interface{} `json:"source_metadata" gorm:"type:jsonb;serializer:json"`

	// Timestamps
	PublishedAt   *time.Time `json:"published_at"`
	DeprecatedAt  *time.Time `json:"deprecated_at"`
}

// TableName overrides the default table name for PluginMarketplace
func (PluginMarketplace) TableName() string {
	return "plugin_registry"
}

// AfterFind hook to populate computed fields
func (u *AppUser) AfterFind(tx *gorm.DB) (err error) {
	if u.IsActive {
		u.Status = "active"
	} else {
		u.Status = "suspended"
	}
	return
}