# CloudBox Plugin System Production Architecture

## Executive Summary

This document provides a comprehensive technical architecture for transforming the CloudBox plugin system from its current state to a production-ready system with secure GitHub marketplace integration, proper authentication, and complete plugin lifecycle management.

### Current Issues Identified
1. **Plugin enable/disable returns 500 errors** - Authentication and persistence failures
2. **Install plugin button non-functional** - No marketplace integration implementation
3. **Security vulnerabilities** - File-based status storage without validation
4. **Missing GitHub integration** - No marketplace or secure installation mechanism

### Key Architectural Decisions
- **Technology Stack**: Go/Gin backend, SvelteKit frontend, PostgreSQL database
- **Authentication**: JWT Bearer tokens for admin operations
- **Plugin Source**: GitHub repositories only (curated marketplace)
- **Security Model**: Repository validation, manifest verification, sandboxed execution
- **Database Strategy**: Plugin metadata and state persistence in PostgreSQL

### System Component Overview
- **Plugin Marketplace Service**: GitHub integration and plugin discovery
- **Plugin Installation Engine**: Secure download, validation, and installation
- **Plugin State Manager**: Database-backed enable/disable with persistence
- **Security Validation Service**: Repository verification and manifest validation

## For Backend Engineers

### Database Schema Design

#### Core Plugin Management Tables
```sql
-- Plugin registry table
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    author VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    github_repo VARCHAR(255) NOT NULL,
    github_owner VARCHAR(255) NOT NULL,
    github_ref VARCHAR(255) DEFAULT 'main',
    manifest_hash VARCHAR(64) NOT NULL,
    status plugin_status DEFAULT 'disabled',
    install_path VARCHAR(500),
    installed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Plugin status enum
CREATE TYPE plugin_status AS ENUM ('enabled', 'disabled', 'installing', 'error', 'uninstalling');

-- Plugin marketplace table (curated plugins)
CREATE TABLE plugin_marketplace (
    id SERIAL PRIMARY KEY,
    repo_url VARCHAR(255) UNIQUE NOT NULL,
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    approved BOOLEAN DEFAULT false,
    featured BOOLEAN DEFAULT false,
    category plugin_category DEFAULT 'utility',
    tags TEXT[],
    download_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id)
);

-- Plugin categories enum
CREATE TYPE plugin_category AS ENUM (
    'dashboard', 'project-tools', 'database', 'api-tools', 
    'deployment', 'monitoring', 'security', 'utility'
);

-- Plugin permissions table
CREATE TABLE plugin_permissions (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER REFERENCES plugins(id) ON DELETE CASCADE,
    permission VARCHAR(100) NOT NULL,
    granted_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Plugin installation log table
CREATE TABLE plugin_installation_logs (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER REFERENCES plugins(id),
    operation plugin_operation NOT NULL,
    status installation_status NOT NULL,
    message TEXT,
    details JSONB,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Plugin operation enum
CREATE TYPE plugin_operation AS ENUM ('install', 'uninstall', 'enable', 'disable', 'update');

-- Installation status enum
CREATE TYPE installation_status AS ENUM ('pending', 'running', 'success', 'failed', 'cancelled');
```

#### Migration Script
```sql
-- Create plugin-related tables and types
-- File: 011_plugin_management.sql

BEGIN;

-- Create enums first
CREATE TYPE plugin_status AS ENUM ('enabled', 'disabled', 'installing', 'error', 'uninstalling');
CREATE TYPE plugin_category AS ENUM (
    'dashboard', 'project-tools', 'database', 'api-tools', 
    'deployment', 'monitoring', 'security', 'utility'
);
CREATE TYPE plugin_operation AS ENUM ('install', 'uninstall', 'enable', 'disable', 'update');
CREATE TYPE installation_status AS ENUM ('pending', 'running', 'success', 'failed', 'cancelled');

-- Main plugin registry
CREATE TABLE plugins (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) UNIQUE NOT NULL,
    display_name VARCHAR(255) NOT NULL,
    description TEXT,
    author VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    github_repo VARCHAR(255) NOT NULL,
    github_owner VARCHAR(255) NOT NULL,
    github_ref VARCHAR(255) DEFAULT 'main',
    manifest_hash VARCHAR(64) NOT NULL,
    status plugin_status DEFAULT 'disabled',
    install_path VARCHAR(500),
    ui_config JSONB,
    permissions TEXT[],
    dependencies JSONB,
    installed_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Curated marketplace
CREATE TABLE plugin_marketplace (
    id SERIAL PRIMARY KEY,
    repo_url VARCHAR(255) UNIQUE NOT NULL,
    repo_owner VARCHAR(255) NOT NULL,
    repo_name VARCHAR(255) NOT NULL,
    approved BOOLEAN DEFAULT false,
    featured BOOLEAN DEFAULT false,
    category plugin_category DEFAULT 'utility',
    tags TEXT[],
    download_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    approved_at TIMESTAMP,
    approved_by INTEGER REFERENCES users(id)
);

-- Installation tracking
CREATE TABLE plugin_installation_logs (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER REFERENCES plugins(id),
    operation plugin_operation NOT NULL,
    status installation_status NOT NULL,
    message TEXT,
    details JSONB,
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP
);

-- Indexes for performance
CREATE INDEX idx_plugins_status ON plugins(status);
CREATE INDEX idx_plugins_name ON plugins(name);
CREATE INDEX idx_marketplace_approved ON plugin_marketplace(approved);
CREATE INDEX idx_installation_logs_plugin_id ON plugin_installation_logs(plugin_id);

COMMIT;
```

### API Endpoint Specifications

#### Plugin Management Endpoints
```go
// GET /api/v1/admin/plugins
// List all installed plugins with status and metadata
func (h *PluginHandler) GetAllPlugins(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Response: Array of plugin objects with full metadata
    // Error Handling: 401/403 for auth, 500 for database errors
}

// POST /api/v1/admin/plugins/{pluginName}/enable
// Enable a specific plugin by name
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Parameters: pluginName (path)
    // Response: Success message with updated status
    // Database: Update plugins.status = 'enabled'
    // Error Handling: 404 if plugin not found, 409 if already enabled
}

// POST /api/v1/admin/plugins/{pluginName}/disable
// Disable a specific plugin by name
func (h *PluginHandler) DisablePlugin(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Parameters: pluginName (path)
    // Response: Success message with updated status
    // Database: Update plugins.status = 'disabled'
    // Error Handling: 404 if plugin not found, 409 if already disabled
}

// GET /api/v1/admin/plugins/marketplace
// Get curated plugin marketplace listings
func (h *PluginHandler) GetMarketplacePlugins(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Response: Array of approved marketplace plugins
    // Includes: GitHub metadata, download counts, categories
}

// POST /api/v1/admin/plugins/install
// Install plugin from GitHub repository
func (h *PluginHandler) InstallPlugin(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Body: { "repo_url": "github.com/user/repo", "ref": "main" }
    // Response: Installation job ID and status
    // Process: Async installation with progress tracking
}

// DELETE /api/v1/admin/plugins/{pluginName}
// Uninstall and remove plugin completely
func (h *PluginHandler) UninstallPlugin(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Parameters: pluginName (path)
    // Response: Success message
    // Process: Clean up files, database entries, stop processes
}

// GET /api/v1/admin/plugins/{pluginName}/logs
// Get installation/operation logs for specific plugin
func (h *PluginHandler) GetPluginLogs(c *gin.Context) {
    // Authentication: Admin or SuperAdmin required
    // Parameters: pluginName (path), optional query filters
    // Response: Array of log entries with timestamps and details
}
```

#### Plugin Installation Engine Implementation
```go
type PluginInstaller struct {
    db           *gorm.DB
    cfg          *config.Config
    github       *github.Client
    downloadPath string
}

type InstallationJob struct {
    ID          string                 `json:"id"`
    PluginName  string                 `json:"plugin_name"`
    RepoURL     string                 `json:"repo_url"`
    Status      string                 `json:"status"`
    Progress    int                    `json:"progress"`
    Message     string                 `json:"message"`
    StartedAt   time.Time             `json:"started_at"`
    CompletedAt *time.Time            `json:"completed_at,omitempty"`
    Error       string                 `json:"error,omitempty"`
}

// Core installation workflow
func (pi *PluginInstaller) InstallFromGitHub(repoURL, ref string) (*InstallationJob, error) {
    job := &InstallationJob{
        ID:        generateJobID(),
        RepoURL:   repoURL,
        Status:    "pending",
        StartedAt: time.Now(),
    }
    
    // Start async installation
    go pi.performInstallation(job, repoURL, ref)
    
    return job, nil
}

func (pi *PluginInstaller) performInstallation(job *InstallationJob, repoURL, ref string) {
    defer pi.finalizeJob(job)
    
    // Step 1: Validate repository (10%)
    if err := pi.validateRepository(repoURL); err != nil {
        job.Error = fmt.Sprintf("Repository validation failed: %v", err)
        job.Status = "failed"
        return
    }
    job.Progress = 10
    
    // Step 2: Download and extract (30%)
    pluginPath, err := pi.downloadPlugin(repoURL, ref)
    if err != nil {
        job.Error = fmt.Sprintf("Download failed: %v", err)
        job.Status = "failed"
        return
    }
    job.Progress = 30
    
    // Step 3: Validate manifest (50%)
    manifest, err := pi.validateManifest(pluginPath)
    if err != nil {
        job.Error = fmt.Sprintf("Manifest validation failed: %v", err)
        job.Status = "failed"
        return
    }
    job.Progress = 50
    
    // Step 4: Check dependencies (70%)
    if err := pi.checkDependencies(manifest); err != nil {
        job.Error = fmt.Sprintf("Dependency check failed: %v", err)
        job.Status = "failed"
        return
    }
    job.Progress = 70
    
    // Step 5: Install plugin (90%)
    if err := pi.installPlugin(manifest, pluginPath); err != nil {
        job.Error = fmt.Sprintf("Installation failed: %v", err)
        job.Status = "failed"
        return
    }
    job.Progress = 90
    
    // Step 6: Register in database (100%)
    if err := pi.registerPlugin(manifest, repoURL, ref, pluginPath); err != nil {
        job.Error = fmt.Sprintf("Database registration failed: %v", err)
        job.Status = "failed"
        return
    }
    
    job.Progress = 100
    job.Status = "completed"
    job.PluginName = manifest.Name
    job.Message = "Plugin installed successfully"
}
```

### Security Implementation Guide

#### Repository Validation Service
```go
type SecurityValidator struct {
    allowedOwners []string
    github        *github.Client
}

func (sv *SecurityValidator) ValidateRepository(repoURL string) error {
    // Parse repository URL
    owner, repo, err := parseGitHubURL(repoURL)
    if err != nil {
        return fmt.Errorf("invalid repository URL: %v", err)
    }
    
    // Check if owner is in curated list
    if !sv.isApprovedOwner(owner) {
        return fmt.Errorf("repository owner '%s' is not approved", owner)
    }
    
    // Verify repository exists and is accessible
    _, _, err = sv.github.Repositories.Get(context.Background(), owner, repo)
    if err != nil {
        return fmt.Errorf("repository not accessible: %v", err)
    }
    
    // Check for plugin.json in root
    _, _, _, err = sv.github.Repositories.GetContents(
        context.Background(), owner, repo, "plugin.json", nil)
    if err != nil {
        return fmt.Errorf("plugin.json not found in repository root")
    }
    
    return nil
}

func (sv *SecurityValidator) ValidateManifest(manifestPath string) (*PluginManifest, error) {
    data, err := ioutil.ReadFile(manifestPath)
    if err != nil {
        return nil, fmt.Errorf("failed to read manifest: %v", err)
    }
    
    var manifest PluginManifest
    if err := json.Unmarshal(data, &manifest); err != nil {
        return nil, fmt.Errorf("invalid JSON manifest: %v", err)
    }
    
    // Validate required fields
    if manifest.Name == "" {
        return nil, fmt.Errorf("plugin name is required")
    }
    
    if manifest.Version == "" {
        return nil, fmt.Errorf("plugin version is required")
    }
    
    // Validate permissions
    for _, perm := range manifest.Permissions {
        if !sv.isValidPermission(perm) {
            return nil, fmt.Errorf("invalid permission: %s", perm)
        }
    }
    
    // Calculate manifest hash for integrity
    manifest.Hash = calculateHash(data)
    
    return &manifest, nil
}

// Valid plugin permissions
var validPermissions = map[string]bool{
    "database:read":     true,
    "database:write":    true,
    "functions:deploy":  true,
    "functions:manage":  true,
    "webhooks:create":   true,
    "webhooks:manage":   true,
    "storage:read":      true,
    "storage:write":     true,
    "projects:read":     true,
    "projects:manage":   true,
    "users:read":        true,
    "users:manage":      true,
}
```

### Authentication and Authorization Implementation

#### JWT Middleware Enhancement
```go
// Enhanced authentication middleware for plugin management
func RequirePluginAdmin() gin.HandlerFunc {
    return gin.HandlerFunc(func(c *gin.Context) {
        userRole := c.GetString("user_role")
        userID := c.GetString("user_id")
        
        // Log authentication attempt
        log.Printf("Plugin admin access attempt: user_id=%s, role=%s", userID, userRole)
        
        // Check for valid admin roles
        if userRole != "admin" && userRole != "superadmin" {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error":   "Plugin management requires admin privileges",
                "code":    "INSUFFICIENT_PRIVILEGES",
            })
            c.Abort()
            return
        }
        
        // Log successful authorization
        log.Printf("Plugin admin access granted: user_id=%s, role=%s", userID, userRole)
        c.Next()
    })
}

// Enhanced plugin handler with proper error handling
func (h *PluginHandler) EnablePlugin(c *gin.Context) {
    pluginName := c.Param("pluginName")
    if pluginName == "" {
        c.JSON(http.StatusBadRequest, gin.H{
            "success": false,
            "error":   "Plugin name is required",
            "code":    "MISSING_PLUGIN_NAME",
        })
        return
    }
    
    // Start database transaction
    tx := h.db.Begin()
    defer func() {
        if r := recover(); r != nil {
            tx.Rollback()
            c.JSON(http.StatusInternalServerError, gin.H{
                "success": false,
                "error":   "Internal server error during plugin enable",
                "code":    "INTERNAL_ERROR",
            })
        }
    }()
    
    // Find plugin in database
    var plugin Plugin
    if err := tx.Where("name = ?", pluginName).First(&plugin).Error; err != nil {
        tx.Rollback()
        if errors.Is(err, gorm.ErrRecordNotFound) {
            c.JSON(http.StatusNotFound, gin.H{
                "success": false,
                "error":   fmt.Sprintf("Plugin '%s' not found", pluginName),
                "code":    "PLUGIN_NOT_FOUND",
            })
        } else {
            c.JSON(http.StatusInternalServerError, gin.H{
                "success": false,
                "error":   "Database error",
                "code":    "DATABASE_ERROR",
            })
        }
        return
    }
    
    // Check current status
    if plugin.Status == "enabled" {
        tx.Rollback()
        c.JSON(http.StatusConflict, gin.H{
            "success": false,
            "error":   fmt.Sprintf("Plugin '%s' is already enabled", pluginName),
            "code":    "ALREADY_ENABLED",
        })
        return
    }
    
    // Update plugin status
    if err := tx.Model(&plugin).Update("status", "enabled").Error; err != nil {
        tx.Rollback()
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error":   "Failed to update plugin status",
            "code":    "UPDATE_FAILED",
        })
        return
    }
    
    // Log the operation
    log := PluginLog{
        PluginID:  plugin.ID,
        Operation: "enable",
        Status:    "success",
        Message:   fmt.Sprintf("Plugin '%s' enabled successfully", pluginName),
        UserID:    c.GetString("user_id"),
    }
    
    if err := tx.Create(&log).Error; err != nil {
        // Log error but don't fail the operation
        fmt.Printf("Failed to create plugin log: %v\n", err)
    }
    
    // Commit transaction
    tx.Commit()
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": fmt.Sprintf("Plugin '%s' enabled successfully", pluginName),
        "plugin": gin.H{
            "name":   plugin.Name,
            "status": "enabled",
        },
    })
}
```

## For Frontend Engineers

### Component Architecture and State Management

#### Enhanced Plugin Store
```typescript
// lib/stores/plugins.ts
import { writable, derived } from 'svelte/store';
import { API_ENDPOINTS, createApiRequest } from '$lib/config';
import { auth } from './auth';
import { addToast } from './toast';

export interface Plugin {
  id: number;
  name: string;
  display_name: string;
  description: string;
  author: string;
  version: string;
  github_repo: string;
  status: 'enabled' | 'disabled' | 'installing' | 'error' | 'uninstalling';
  ui_config?: {
    dashboard_tab?: any;
    project_menu?: any;
  };
  permissions: string[];
  dependencies?: Record<string, string>;
  installed_at: string;
  updated_at: string;
}

export interface MarketplacePlugin {
  id: number;
  repo_url: string;
  repo_owner: string;
  repo_name: string;
  category: string;
  tags: string[];
  download_count: number;
  featured: boolean;
  last_updated: string;
  manifest?: {
    name: string;
    description: string;
    version: string;
    author: string;
  };
}

export interface InstallationJob {
  id: string;
  plugin_name: string;
  repo_url: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  message: string;
  started_at: string;
  completed_at?: string;
  error?: string;
}

// Core stores
export const plugins = writable<Plugin[]>([]);
export const marketplacePlugins = writable<MarketplacePlugin[]>([]);
export const installationJobs = writable<Map<string, InstallationJob>>(new Map());
export const isLoading = writable(false);
export const selectedPlugin = writable<Plugin | null>(null);

// Derived stores
export const enabledPlugins = derived(plugins, $plugins => 
  $plugins.filter(p => p.status === 'enabled')
);

export const disabledPlugins = derived(plugins, $plugins => 
  $plugins.filter(p => p.status === 'disabled')
);

export const pluginStats = derived(plugins, $plugins => ({
  total: $plugins.length,
  enabled: $plugins.filter(p => p.status === 'enabled').length,
  disabled: $plugins.filter(p => p.status === 'disabled').length,
  installing: $plugins.filter(p => p.status === 'installing').length,
  errors: $plugins.filter(p => p.status === 'error').length
}));

// Actions
export const pluginActions = {
  async loadPlugins() {
    isLoading.set(true);
    try {
      const response = await createApiRequest(`${API_ENDPOINTS.admin.plugins.list}`, {
        credentials: 'include',
        headers: {
          Authorization: `Bearer ${auth.getToken()}`
        }
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.success) {
        plugins.set(data.plugins || []);
      } else {
        throw new Error(data.error || 'Failed to load plugins');
      }
    } catch (error) {
      console.error('Failed to load plugins:', error);
      addToast(`Failed to load plugins: ${error.message}`, 'error');
    } finally {
      isLoading.set(false);
    }
  },

  async loadMarketplace() {
    try {
      const response = await createApiRequest(`${API_ENDPOINTS.admin.plugins.marketplace}`, {
        credentials: 'include',
        headers: {
          Authorization: `Bearer ${auth.getToken()}`
        }
      });

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.success) {
        marketplacePlugins.set(data.plugins || []);
      } else {
        throw new Error(data.error || 'Failed to load marketplace');
      }
    } catch (error) {
      console.error('Failed to load marketplace:', error);
      addToast(`Failed to load marketplace: ${error.message}`, 'error');
    }
  },

  async togglePlugin(plugin: Plugin) {
    const action = plugin.status === 'enabled' ? 'disable' : 'enable';
    const endpoint = action === 'enable' 
      ? API_ENDPOINTS.admin.plugins.enable(plugin.name)
      : API_ENDPOINTS.admin.plugins.disable(plugin.name);

    try {
      const response = await createApiRequest(endpoint, {
        method: 'POST',
        credentials: 'include',
        headers: {
          Authorization: `Bearer ${auth.getToken()}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.success) {
        // Update plugin status in store
        plugins.update(list => 
          list.map(p => 
            p.name === plugin.name 
              ? { ...p, status: action === 'enable' ? 'enabled' : 'disabled' }
              : p
          )
        );
        addToast(`Plugin ${action}d successfully`, 'success');
      } else {
        throw new Error(data.error || `Failed to ${action} plugin`);
      }
    } catch (error) {
      console.error(`Failed to ${action} plugin:`, error);
      addToast(`Failed to ${action} plugin: ${error.message}`, 'error');
    }
  },

  async installPlugin(repoUrl: string, ref: string = 'main') {
    try {
      const response = await createApiRequest(`${API_ENDPOINTS.admin.plugins.install}`, {
        method: 'POST',
        credentials: 'include',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${auth.getToken()}`
        },
        body: JSON.stringify({ repo_url: repoUrl, ref })
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.success && data.job) {
        // Add job to tracking
        installationJobs.update(jobs => {
          jobs.set(data.job.id, data.job);
          return jobs;
        });
        
        // Start polling for job status
        this.pollInstallationJob(data.job.id);
        
        addToast('Plugin installation started', 'info');
        return data.job;
      } else {
        throw new Error(data.error || 'Failed to start installation');
      }
    } catch (error) {
      console.error('Failed to install plugin:', error);
      addToast(`Failed to install plugin: ${error.message}`, 'error');
      throw error;
    }
  },

  async pollInstallationJob(jobId: string) {
    const pollInterval = setInterval(async () => {
      try {
        const response = await createApiRequest(`${API_ENDPOINTS.admin.plugins.jobStatus(jobId)}`, {
          credentials: 'include',
          headers: {
            Authorization: `Bearer ${auth.getToken()}`
          }
        });

        if (response.ok) {
          const data = await response.json();
          if (data.success && data.job) {
            installationJobs.update(jobs => {
              jobs.set(jobId, data.job);
              return jobs;
            });

            // Stop polling if job is complete
            if (data.job.status === 'completed' || data.job.status === 'failed') {
              clearInterval(pollInterval);
              
              if (data.job.status === 'completed') {
                addToast(`Plugin '${data.job.plugin_name}' installed successfully`, 'success');
                // Reload plugins list
                this.loadPlugins();
              } else {
                addToast(`Plugin installation failed: ${data.job.error}`, 'error');
              }
            }
          }
        }
      } catch (error) {
        console.error('Failed to poll installation job:', error);
        clearInterval(pollInterval);
      }
    }, 2000); // Poll every 2 seconds
  },

  async uninstallPlugin(plugin: Plugin) {
    try {
      const response = await createApiRequest(API_ENDPOINTS.admin.plugins.uninstall(plugin.name), {
        method: 'DELETE',
        credentials: 'include',
        headers: {
          Authorization: `Bearer ${auth.getToken()}`
        }
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.error || `HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.success) {
        // Remove plugin from store
        plugins.update(list => list.filter(p => p.name !== plugin.name));
        addToast('Plugin uninstalled successfully', 'success');
      } else {
        throw new Error(data.error || 'Failed to uninstall plugin');
      }
    } catch (error) {
      console.error('Failed to uninstall plugin:', error);
      addToast(`Failed to uninstall plugin: ${error.message}`, 'error');
    }
  }
};
```

#### GitHub Marketplace Modal Component
```svelte
<!-- lib/components/ui/plugin-marketplace-modal.svelte -->
<script lang="ts">
  import { onMount, createEventDispatcher } from 'svelte';
  import Button from './button.svelte';
  import Card from './card.svelte';
  import Icon from './icon.svelte';
  import Badge from './badge.svelte';
  import Modal from './modal.svelte';
  import { marketplacePlugins, pluginActions, installationJobs } from '$lib/stores/plugins';
  import { addToast } from '$lib/stores/toast';
  
  export let show = false;
  
  const dispatch = createEventDispatcher();
  
  let selectedCategory = 'all';
  let searchQuery = '';
  let installing = new Set<string>();
  
  const categories = [
    { value: 'all', label: 'All Plugins', icon: 'puzzle' },
    { value: 'dashboard', label: 'Dashboard', icon: 'layout-dashboard' },
    { value: 'project-tools', label: 'Project Tools', icon: 'folder' },
    { value: 'database', label: 'Database', icon: 'database' },
    { value: 'api-tools', label: 'API Tools', icon: 'code' },
    { value: 'deployment', label: 'Deployment', icon: 'upload' },
    { value: 'monitoring', label: 'Monitoring', icon: 'activity' },
    { value: 'security', label: 'Security', icon: 'shield' },
    { value: 'utility', label: 'Utility', icon: 'tool' }
  ];
  
  $: filteredPlugins = $marketplacePlugins.filter(plugin => {
    const matchesCategory = selectedCategory === 'all' || plugin.category === selectedCategory;
    const matchesSearch = !searchQuery || 
      plugin.repo_name.toLowerCase().includes(searchQuery.toLowerCase()) ||
      plugin.manifest?.description?.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesCategory && matchesSearch;
  });
  
  async function installPlugin(plugin: any) {
    if (installing.has(plugin.repo_url)) return;
    
    installing.add(plugin.repo_url);
    installing = installing; // Trigger reactivity
    
    try {
      await pluginActions.installPlugin(plugin.repo_url);
    } catch (error) {
      // Error already handled in store
    } finally {
      installing.delete(plugin.repo_url);
      installing = installing; // Trigger reactivity
    }
  }
  
  function getCategoryIcon(category: string) {
    const cat = categories.find(c => c.value === category);
    return cat?.icon || 'puzzle';
  }
  
  function getInstallationProgress(repoUrl: string) {
    for (const [jobId, job] of $installationJobs) {
      if (job.repo_url === repoUrl && job.status === 'running') {
        return job.progress;
      }
    }
    return 0;
  }
  
  onMount(() => {
    if (show) {
      pluginActions.loadMarketplace();
    }
  });
  
  $: if (show) {
    pluginActions.loadMarketplace();
  }
</script>

<Modal bind:show size="large" title="Plugin Marketplace">
  <div class="space-y-6">
    <!-- Search and Filter -->
    <div class="flex flex-col sm:flex-row gap-4">
      <div class="flex-1">
        <div class="relative">
          <Icon name="search" size={16} className="absolute left-3 top-1/2 transform -translate-y-1/2 text-muted-foreground" />
          <input
            type="text"
            placeholder="Search plugins..."
            bind:value={searchQuery}
            class="w-full pl-10 pr-4 py-2 border border-border rounded-md bg-background text-foreground placeholder-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
          />
        </div>
      </div>
      
      <div class="sm:w-48">
        <select
          bind:value={selectedCategory}
          class="w-full px-3 py-2 border border-border rounded-md bg-background text-foreground focus:outline-none focus:ring-2 focus:ring-primary focus:border-transparent"
        >
          {#each categories as category}
            <option value={category.value}>{category.label}</option>
          {/each}
        </select>
      </div>
    </div>
    
    <!-- Plugin Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 max-h-96 overflow-y-auto">
      {#each filteredPlugins as plugin (plugin.id)}
        <Card className="p-4 hover:shadow-md transition-shadow">
          <div class="space-y-3">
            <!-- Plugin Header -->
            <div class="flex items-start justify-between">
              <div class="flex items-center space-x-2">
                <Icon name={getCategoryIcon(plugin.category)} size={20} className="text-primary" />
                <div>
                  <h3 class="font-medium text-sm">{plugin.manifest?.name || plugin.repo_name}</h3>
                  <p class="text-xs text-muted-foreground">by {plugin.repo_owner}</p>
                </div>
              </div>
              
              {#if plugin.featured}
                <Badge variant="default" className="text-xs">Featured</Badge>
              {/if}
            </div>
            
            <!-- Plugin Description -->
            <p class="text-sm text-muted-foreground line-clamp-2">
              {plugin.manifest?.description || 'No description available'}
            </p>
            
            <!-- Plugin Metadata -->
            <div class="flex items-center justify-between text-xs text-muted-foreground">
              <span>v{plugin.manifest?.version || '1.0.0'}</span>
              <span>{plugin.download_count} downloads</span>
            </div>
            
            <!-- Tags -->
            {#if plugin.tags && plugin.tags.length > 0}
              <div class="flex flex-wrap gap-1">
                {#each plugin.tags.slice(0, 3) as tag}
                  <Badge variant="outline" className="text-xs">{tag}</Badge>
                {/each}
                {#if plugin.tags.length > 3}
                  <Badge variant="outline" className="text-xs">+{plugin.tags.length - 3}</Badge>
                {/if}
              </div>
            {/if}
            
            <!-- Install Button -->
            <div class="pt-2">
              {#if installing.has(plugin.repo_url)}
                <div class="space-y-2">
                  <div class="w-full bg-muted rounded-full h-2">
                    <div 
                      class="bg-primary h-2 rounded-full transition-all duration-300"
                      style="width: {getInstallationProgress(plugin.repo_url)}%"
                    ></div>
                  </div>
                  <p class="text-xs text-center text-muted-foreground">Installing...</p>
                </div>
              {:else}
                <Button 
                  on:click={() => installPlugin(plugin)}
                  variant="default" 
                  size="sm" 
                  className="w-full"
                >
                  <Icon name="download" size={14} className="mr-2" />
                  Install
                </Button>
              {/if}
            </div>
          </div>
        </Card>
      {:else}
        <div class="col-span-full text-center py-8">
          <Icon name="puzzle" size={48} className="mx-auto text-muted-foreground mb-4" />
          <h3 class="text-lg font-medium text-foreground mb-2">No plugins found</h3>
          <p class="text-muted-foreground">Try adjusting your search or category filter.</p>
        </div>
      {/each}
    </div>
    
    <!-- Footer -->
    <div class="flex justify-between items-center pt-4 border-t border-border">
      <p class="text-sm text-muted-foreground">
        {filteredPlugins.length} plugin{filteredPlugins.length !== 1 ? 's' : ''} available
      </p>
      
      <div class="flex gap-2">
        <Button on:click={() => show = false} variant="outline">
          Close
        </Button>
      </div>
    </div>
  </div>
</Modal>
```

### Routing and Navigation Architecture

#### Enhanced Plugin Management Page
```svelte
<!-- routes/dashboard/admin/plugins/+page.svelte -->
<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import Button from '$lib/components/ui/button.svelte';
  import Card from '$lib/components/ui/card.svelte';
  import Icon from '$lib/components/ui/icon.svelte';
  import Badge from '$lib/components/ui/badge.svelte';
  import PluginMarketplaceModal from '$lib/components/ui/plugin-marketplace-modal.svelte';
  import PluginDetailsModal from '$lib/components/ui/plugin-details-modal.svelte';
  import InstallationProgressModal from '$lib/components/ui/installation-progress-modal.svelte';
  
  import { 
    plugins, 
    pluginStats, 
    selectedPlugin, 
    installationJobs,
    pluginActions,
    isLoading 
  } from '$lib/stores/plugins';
  import { addToast } from '$lib/stores/toast';
  
  let showMarketplace = false;
  let showPluginDetails = false;
  let showInstallationProgress = false;
  let confirmDeletePlugin: Plugin | null = null;
  
  // Handle URL parameters for direct plugin access
  $: {
    const pluginName = $page.url.searchParams.get('plugin');
    if (pluginName && $plugins.length > 0) {
      const plugin = $plugins.find(p => p.name === pluginName);
      if (plugin) {
        selectedPlugin.set(plugin);
        showPluginDetails = true;
      }
    }
  }
  
  function getStatusColor(status: string) {
    switch (status) {
      case 'enabled': return 'bg-green-100 text-green-800 dark:bg-green-900/20 dark:text-green-300';
      case 'disabled': return 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-300';
      case 'installing': return 'bg-blue-100 text-blue-800 dark:bg-blue-900/20 dark:text-blue-300';
      case 'error': return 'bg-red-100 text-red-800 dark:bg-red-900/20 dark:text-red-300';
      case 'uninstalling': return 'bg-orange-100 text-orange-800 dark:bg-orange-900/20 dark:text-orange-300';
      default: return 'bg-gray-100 text-gray-800 dark:bg-gray-900/20 dark:text-gray-300';
    }
  }
  
  function getTypeIcon(type: string) {
    switch (type) {
      case 'dashboard-plugin': return 'layout-dashboard';
      case 'project-plugin': return 'folder';
      case 'system-plugin': return 'cog';
      default: return 'puzzle';
    }
  }
  
  function showDetails(plugin: Plugin) {
    selectedPlugin.set(plugin);
    showPluginDetails = true;
  }
  
  async function confirmUninstall(plugin: Plugin) {
    confirmDeletePlugin = plugin;
  }
  
  async function uninstallPlugin() {
    if (!confirmDeletePlugin) return;
    
    await pluginActions.uninstallPlugin(confirmDeletePlugin);
    confirmDeletePlugin = null;
    
    // Close details modal if it was showing the uninstalled plugin
    if ($selectedPlugin?.name === confirmDeletePlugin?.name) {
      showPluginDetails = false;
      selectedPlugin.set(null);
    }
  }
  
  // Check if there are active installation jobs
  $: hasActiveJobs = Array.from($installationJobs.values()).some(
    job => job.status === 'running' || job.status === 'pending'
  );
  
  onMount(() => {
    pluginActions.loadPlugins();
  });
</script>

<svelte:head>
  <title>Plugin Management - Admin - CloudBox</title>
</svelte:head>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-bold text-foreground">Plugin Management</h1>
      <p class="text-muted-foreground mt-1">
        Manage CloudBox plugins and extensions securely
      </p>
    </div>
    
    <div class="flex gap-3">
      {#if hasActiveJobs}
        <Button on:click={() => showInstallationProgress = true} variant="outline">
          <Icon name="activity" size={16} className="mr-2 animate-pulse" />
          Installation Progress
        </Button>
      {/if}
      
      <Button on:click={() => pluginActions.loadPlugins()} variant="outline">
        <Icon name="refresh" size={16} className="mr-2" />
        Refresh
      </Button>
      
      <Button on:click={() => showMarketplace = true} variant="default">
        <Icon name="plus" size={16} className="mr-2" />
        Browse Marketplace
      </Button>
    </div>
  </div>

  <!-- Plugin Stats -->
  <div class="grid grid-cols-1 md:grid-cols-5 gap-4">
    <Card className="p-4">
      <div class="flex items-center">
        <Icon name="puzzle" size={24} className="text-blue-500 mr-3" />
        <div>
          <p class="text-sm text-muted-foreground">Total Plugins</p>
          <p class="text-2xl font-bold text-foreground">{$pluginStats.total}</p>
        </div>
      </div>
    </Card>
    
    <Card className="p-4">
      <div class="flex items-center">
        <Icon name="check-circle" size={24} className="text-green-500 mr-3" />
        <div>
          <p class="text-sm text-muted-foreground">Enabled</p>
          <p class="text-2xl font-bold text-foreground">{$pluginStats.enabled}</p>
        </div>
      </div>
    </Card>
    
    <Card className="p-4">
      <div class="flex items-center">
        <Icon name="pause-circle" size={24} className="text-gray-500 mr-3" />
        <div>
          <p class="text-sm text-muted-foreground">Disabled</p>
          <p class="text-2xl font-bold text-foreground">{$pluginStats.disabled}</p>
        </div>
      </div>
    </Card>
    
    <Card className="p-4">
      <div class="flex items-center">
        <Icon name="download" size={24} className="text-blue-500 mr-3" />
        <div>
          <p class="text-sm text-muted-foreground">Installing</p>
          <p class="text-2xl font-bold text-foreground">{$pluginStats.installing}</p>
        </div>
      </div>
    </Card>
    
    <Card className="p-4">
      <div class="flex items-center">
        <Icon name="x-circle" size={24} className="text-red-500 mr-3" />
        <div>
          <p class="text-sm text-muted-foreground">Errors</p>
          <p class="text-2xl font-bold text-foreground">{$pluginStats.errors}</p>
        </div>
      </div>
    </Card>
  </div>

  <!-- Plugins List -->
  <Card>
    <div class="p-6 border-b border-border">
      <h2 class="text-lg font-semibold text-foreground">Installed Plugins</h2>
    </div>

    {#if $isLoading}
      <div class="p-12 text-center">
        <div class="animate-spin rounded-full h-8 w-8 border-b-2 border-primary mx-auto"></div>
        <p class="mt-4 text-muted-foreground">Loading plugins...</p>
      </div>
    {:else if $plugins.length === 0}
      <div class="p-12 text-center">
        <Icon name="puzzle" size={48} className="mx-auto text-muted-foreground mb-4" />
        <h3 class="text-lg font-medium text-foreground mb-2">No plugins installed</h3>
        <p class="text-muted-foreground mb-6">Install your first plugin to extend CloudBox functionality.</p>
        <Button on:click={() => showMarketplace = true}>
          <Icon name="plus" size={16} className="mr-2" />
          Browse Plugin Marketplace
        </Button>
      </div>
    {:else}
      <div class="divide-y divide-border">
        {#each $plugins as plugin (plugin.id)}
          <div class="p-6">
            <div class="flex items-center justify-between">
              <div class="flex items-center space-x-4">
                <div class="flex-shrink-0">
                  <Icon name={getTypeIcon(plugin.ui_config?.type || 'dashboard-plugin')} size={24} className="text-primary" />
                </div>
                
                <div class="flex-1 min-w-0">
                  <div class="flex items-center space-x-3">
                    <h3 class="text-lg font-medium text-foreground">{plugin.display_name || plugin.name}</h3>
                    <Badge className={getStatusColor(plugin.status)}>
                      {plugin.status}
                    </Badge>
                    <span class="text-sm text-muted-foreground">v{plugin.version}</span>
                  </div>
                  
                  <p class="text-sm text-muted-foreground mt-1">{plugin.description}</p>
                  
                  <div class="flex items-center space-x-4 mt-2 text-xs text-muted-foreground">
                    <span>By {plugin.author}</span>
                    <span>•</span>
                    <span>Installed {new Date(plugin.installed_at).toLocaleDateString('en-US')}</span>
                    <span>•</span>
                    <a 
                      href="https://github.com/{plugin.github_repo}" 
                      target="_blank" 
                      class="hover:text-primary"
                    >
                      {plugin.github_repo}
                    </a>
                    {#if plugin.ui_config?.project_menu}
                      <span>•</span>
                      <span>Project Integration</span>
                    {/if}
                  </div>
                </div>
              </div>

              <div class="flex items-center space-x-2">
                <Button 
                  on:click={() => showDetails(plugin)}
                  variant="outline"
                  size="sm"
                >
                  Details
                </Button>
                
                {#if plugin.status === 'enabled' || plugin.status === 'disabled'}
                  <Button 
                    on:click={() => pluginActions.togglePlugin(plugin)}
                    variant={plugin.status === 'enabled' ? 'outline' : 'default'}
                    size="sm"
                  >
                    {plugin.status === 'enabled' ? 'Disable' : 'Enable'}
                  </Button>
                {/if}
                
                <Button 
                  on:click={() => confirmUninstall(plugin)}
                  variant="outline"
                  size="sm"
                  className="text-red-600 hover:text-red-700"
                >
                  Uninstall
                </Button>
              </div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </Card>
</div>

<!-- Modals -->
<PluginMarketplaceModal bind:show={showMarketplace} />
<PluginDetailsModal bind:show={showPluginDetails} />
<InstallationProgressModal bind:show={showInstallationProgress} />

<!-- Confirmation Dialog -->
{#if confirmDeletePlugin}
  <div class="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
    <Card className="max-w-md w-full mx-4 p-6">
      <h3 class="text-lg font-semibold text-foreground mb-4">Confirm Uninstall</h3>
      <p class="text-muted-foreground mb-6">
        Are you sure you want to uninstall <strong>{confirmDeletePlugin.display_name}</strong>? 
        This action cannot be undone and will remove all plugin data.
      </p>
      <div class="flex justify-end gap-3">
        <Button on:click={() => confirmDeletePlugin = null} variant="outline">
          Cancel
        </Button>
        <Button on:click={uninstallPlugin} variant="destructive">
          Uninstall
        </Button>
      </div>
    </Card>
  </div>
{/if}
```

### Performance Optimization Strategies

#### Frontend Caching and State Management
```typescript
// lib/stores/plugin-cache.ts
import { writable } from 'svelte/store';

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  ttl: number;
}

class PluginCache {
  private cache = new Map<string, CacheEntry<any>>();
  private readonly DEFAULT_TTL = 5 * 60 * 1000; // 5 minutes

  set<T>(key: string, data: T, ttl = this.DEFAULT_TTL): void {
    this.cache.set(key, {
      data,
      timestamp: Date.now(),
      ttl
    });
  }

  get<T>(key: string): T | null {
    const entry = this.cache.get(key);
    if (!entry) return null;

    const now = Date.now();
    if (now - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  clear(): void {
    this.cache.clear();
  }

  has(key: string): boolean {
    const entry = this.cache.get(key);
    if (!entry) return false;

    const now = Date.now();
    if (now - entry.timestamp > entry.ttl) {
      this.cache.delete(key);
      return false;
    }

    return true;
  }
}

export const pluginCache = new PluginCache();

// Enhanced plugin actions with caching
export const cachedPluginActions = {
  async loadPlugins(forceRefresh = false) {
    const cacheKey = 'plugins-list';
    
    if (!forceRefresh && pluginCache.has(cacheKey)) {
      const cachedData = pluginCache.get(cacheKey);
      if (cachedData) {
        plugins.set(cachedData);
        return;
      }
    }

    // Load from API
    await pluginActions.loadPlugins();
    
    // Cache the result
    plugins.subscribe(value => {
      if (value.length > 0) {
        pluginCache.set(cacheKey, value);
      }
    })();
  },

  async loadMarketplace(forceRefresh = false) {
    const cacheKey = 'marketplace-plugins';
    
    if (!forceRefresh && pluginCache.has(cacheKey)) {
      const cachedData = pluginCache.get(cacheKey);
      if (cachedData) {
        marketplacePlugins.set(cachedData);
        return;
      }
    }

    // Load from API
    await pluginActions.loadMarketplace();
    
    // Cache the result
    marketplacePlugins.subscribe(value => {
      if (value.length > 0) {
        pluginCache.set(cacheKey, value, 10 * 60 * 1000); // 10 minutes for marketplace
      }
    })();
  }
};
```

## For QA Engineers

### Testable Component Boundaries

#### Plugin Management Test Scenarios
1. **Plugin Lifecycle Testing**
   - Install plugin from marketplace
   - Enable/disable plugin functionality
   - Uninstall plugin and cleanup verification
   - Plugin state persistence across sessions

2. **Authentication Testing**
   - Admin role verification for plugin operations
   - JWT token validation and refresh
   - Unauthorized access prevention
   - Session timeout handling

3. **GitHub Integration Testing**
   - Repository validation and access
   - Manifest parsing and validation
   - Download and extraction process
   - Error handling for invalid repositories

4. **UI/UX Testing**
   - Marketplace modal functionality
   - Installation progress tracking
   - Real-time status updates
   - Error message display and handling

### Data Validation Requirements

#### Plugin Manifest Validation
```go
// Test data for plugin manifest validation
type PluginManifestTest struct {
    Name         string `json:"name" validate:"required,min=3,max=50,alphanum"`
    Version      string `json:"version" validate:"required,semver"`
    Description  string `json:"description" validate:"required,min=10,max=500"`
    Author       string `json:"author" validate:"required,min=2,max=100"`
    Type         string `json:"type" validate:"required,oneof=dashboard-plugin project-plugin system-plugin"`
    Main         string `json:"main" validate:"required,endswith=.js"`
    Permissions  []string `json:"permissions" validate:"dive,oneof=database:read database:write functions:deploy"`
}

// Validation test cases
func TestPluginManifestValidation(t *testing.T) {
    validManifest := PluginManifestTest{
        Name:        "test-plugin",
        Version:     "1.0.0",
        Description: "A test plugin for CloudBox",
        Author:      "Test Author",
        Type:        "dashboard-plugin",
        Main:        "index.js",
        Permissions: []string{"database:read"},
    }
    
    // Test valid manifest
    assert.NoError(t, validator.Struct(validManifest))
    
    // Test invalid cases
    invalidCases := []struct {
        name     string
        modify   func(*PluginManifestTest)
        expected string
    }{
        {
            name: "empty name",
            modify: func(m *PluginManifestTest) { m.Name = "" },
            expected: "Name is required",
        },
        {
            name: "invalid version",
            modify: func(m *PluginManifestTest) { m.Version = "invalid" },
            expected: "Version must be valid semver",
        },
        {
            name: "invalid permission",
            modify: func(m *PluginManifestTest) { m.Permissions = []string{"invalid:permission"} },
            expected: "Invalid permission",
        },
    }
    
    for _, tc := range invalidCases {
        t.Run(tc.name, func(t *testing.T) {
            manifest := validManifest
            tc.modify(&manifest)
            err := validator.Struct(manifest)
            assert.Error(t, err)
            assert.Contains(t, err.Error(), tc.expected)
        })
    }
}
```

### Integration Points Testing

#### End-to-End Plugin Installation Test
```javascript
// e2e/plugin-installation.spec.js
import { test, expect } from '@playwright/test';

test.describe('Plugin Installation Flow', () => {
  test.beforeEach(async ({ page }) => {
    // Login as admin
    await page.goto('/login');
    await page.fill('[data-testid="email"]', 'admin@cloudbox.test');
    await page.fill('[data-testid="password"]', 'admin123');
    await page.click('[data-testid="login-button"]');
    await expect(page).toHaveURL('/dashboard');
    
    // Navigate to plugin management
    await page.goto('/dashboard/admin/plugins');
  });

  test('should install plugin from marketplace', async ({ page }) => {
    // Open marketplace modal
    await page.click('[data-testid="browse-marketplace-button"]');
    await expect(page.locator('[data-testid="marketplace-modal"]')).toBeVisible();
    
    // Search for test plugin
    await page.fill('[data-testid="marketplace-search"]', 'test-plugin');
    await page.waitForTimeout(500); // Debounce search
    
    // Install first plugin in results
    const firstPlugin = page.locator('[data-testid="marketplace-plugin"]').first();
    await expect(firstPlugin).toBeVisible();
    
    const installButton = firstPlugin.locator('[data-testid="install-button"]');
    await installButton.click();
    
    // Verify installation progress
    await expect(page.locator('[data-testid="installation-progress"]')).toBeVisible();
    
    // Wait for installation to complete (with timeout)
    await expect(page.locator('[data-testid="installation-success"]')).toBeVisible({ timeout: 30000 });
    
    // Close marketplace modal
    await page.click('[data-testid="marketplace-close"]');
    
    // Verify plugin appears in installed list
    await expect(page.locator('[data-testid="plugin-list"]')).toContainText('test-plugin');
    
    // Verify plugin is disabled by default
    const pluginRow = page.locator('[data-testid="plugin-row-test-plugin"]');
    await expect(pluginRow.locator('[data-testid="plugin-status"]')).toContainText('disabled');
  });

  test('should enable and disable plugin', async ({ page }) => {
    // Assuming test plugin is already installed
    const pluginRow = page.locator('[data-testid="plugin-row-test-plugin"]');
    await expect(pluginRow).toBeVisible();
    
    // Enable plugin
    const enableButton = pluginRow.locator('[data-testid="enable-button"]');
    await enableButton.click();
    
    // Verify status change
    await expect(pluginRow.locator('[data-testid="plugin-status"]')).toContainText('enabled');
    await expect(enableButton).toContainText('Disable');
    
    // Disable plugin
    await enableButton.click();
    
    // Verify status change
    await expect(pluginRow.locator('[data-testid="plugin-status"]')).toContainText('disabled');
    await expect(enableButton).toContainText('Enable');
  });

  test('should handle plugin installation errors', async ({ page }) => {
    // Mock invalid repository
    await page.route('**/api/v1/admin/plugins/install', route => {
      route.fulfill({
        status: 400,
        contentType: 'application/json',
        body: JSON.stringify({
          success: false,
          error: 'Repository validation failed: plugin.json not found'
        })
      });
    });
    
    // Try to install invalid plugin
    await page.click('[data-testid="browse-marketplace-button"]');
    const invalidPlugin = page.locator('[data-testid="marketplace-plugin"]').first();
    await invalidPlugin.locator('[data-testid="install-button"]').click();
    
    // Verify error handling
    await expect(page.locator('[data-testid="error-toast"]')).toBeVisible();
    await expect(page.locator('[data-testid="error-toast"]')).toContainText('Repository validation failed');
  });

  test('should uninstall plugin with confirmation', async ({ page }) => {
    // Assuming test plugin is installed
    const pluginRow = page.locator('[data-testid="plugin-row-test-plugin"]');
    await expect(pluginRow).toBeVisible();
    
    // Click uninstall
    await pluginRow.locator('[data-testid="uninstall-button"]').click();
    
    // Verify confirmation dialog
    const confirmDialog = page.locator('[data-testid="confirm-uninstall-dialog"]');
    await expect(confirmDialog).toBeVisible();
    await expect(confirmDialog).toContainText('test-plugin');
    
    // Confirm uninstall
    await confirmDialog.locator('[data-testid="confirm-uninstall-button"]').click();
    
    // Verify plugin is removed
    await expect(pluginRow).not.toBeVisible();
    await expect(page.locator('[data-testid="success-toast"]')).toContainText('Plugin uninstalled successfully');
  });
});
```

### Performance Benchmarks

#### Performance Testing Requirements
1. **API Response Times**
   - Plugin list retrieval: < 200ms
   - Plugin enable/disable: < 500ms
   - Plugin installation: < 30 seconds
   - Marketplace loading: < 1 second

2. **Database Performance**
   - Plugin queries with proper indexing
   - Bulk operations optimization
   - Connection pooling efficiency

3. **Frontend Performance**
   - Component rendering times
   - State update efficiency
   - Memory usage monitoring

## For Security Analysts

### Authentication Flow and Security Model

#### Multi-Layer Security Architecture
```go
// Security validation pipeline
type SecurityPipeline struct {
    validators []SecurityValidator
}

type SecurityValidator interface {
    Validate(ctx context.Context, plugin *PluginInstallRequest) error
}

// Repository whitelist validator
type RepositoryWhitelistValidator struct {
    allowedOwners map[string]bool
    allowedRepos  map[string]bool
}

func (v *RepositoryWhitelistValidator) Validate(ctx context.Context, req *PluginInstallRequest) error {
    owner, repo := parseGitHubURL(req.RepoURL)
    
    // Check owner whitelist
    if !v.allowedOwners[owner] {
        return &SecurityError{
            Code:    "OWNER_NOT_WHITELISTED",
            Message: fmt.Sprintf("Repository owner '%s' is not approved", owner),
            Severity: "HIGH",
        }
    }
    
    // Check specific repository if configured
    repoKey := fmt.Sprintf("%s/%s", owner, repo)
    if len(v.allowedRepos) > 0 && !v.allowedRepos[repoKey] {
        return &SecurityError{
            Code:    "REPO_NOT_WHITELISTED",
            Message: fmt.Sprintf("Repository '%s' is not approved", repoKey),
            Severity: "HIGH",
        }
    }
    
    return nil
}

// Manifest security validator
type ManifestSecurityValidator struct {
    maxPermissions int
    dangerousPerms map[string]bool
}

func (v *ManifestSecurityValidator) Validate(ctx context.Context, req *PluginInstallRequest) error {
    manifest, err := v.fetchManifest(req.RepoURL, req.Ref)
    if err != nil {
        return err
    }
    
    // Check permission count
    if len(manifest.Permissions) > v.maxPermissions {
        return &SecurityError{
            Code:    "TOO_MANY_PERMISSIONS",
            Message: fmt.Sprintf("Plugin requests %d permissions, maximum allowed: %d", 
                len(manifest.Permissions), v.maxPermissions),
            Severity: "MEDIUM",
        }
    }
    
    // Check for dangerous permissions
    for _, perm := range manifest.Permissions {
        if v.dangerousPerms[perm] {
            return &SecurityError{
                Code:    "DANGEROUS_PERMISSION",
                Message: fmt.Sprintf("Plugin requests dangerous permission: %s", perm),
                Severity: "HIGH",
            }
        }
    }
    
    return nil
}

// File integrity validator
type FileIntegrityValidator struct {
    maxFileSize   int64
    allowedExts   map[string]bool
    forbiddenPaths map[string]bool
}

func (v *FileIntegrityValidator) Validate(ctx context.Context, req *PluginInstallRequest) error {
    files, err := v.scanRepository(req.RepoURL, req.Ref)
    if err != nil {
        return err
    }
    
    for _, file := range files {
        // Check file size
        if file.Size > v.maxFileSize {
            return &SecurityError{
                Code:    "FILE_TOO_LARGE",
                Message: fmt.Sprintf("File %s exceeds maximum size limit", file.Path),
                Severity: "MEDIUM",
            }
        }
        
        // Check file extension
        ext := filepath.Ext(file.Path)
        if !v.allowedExts[ext] {
            return &SecurityError{
                Code:    "FORBIDDEN_FILE_TYPE",
                Message: fmt.Sprintf("File type %s is not allowed: %s", ext, file.Path),
                Severity: "HIGH",
            }
        }
        
        // Check for forbidden paths
        for forbiddenPath := range v.forbiddenPaths {
            if strings.Contains(file.Path, forbiddenPath) {
                return &SecurityError{
                    Code:    "FORBIDDEN_PATH",
                    Message: fmt.Sprintf("File in forbidden path: %s", file.Path),
                    Severity: "HIGH",
                }
            }
        }
    }
    
    return nil
}
```

### Threat Modeling and Vulnerability Assessment

#### Security Threat Matrix
| Threat | Likelihood | Impact | Mitigation |
|--------|------------|--------|------------|
| Malicious Plugin Installation | Medium | High | Repository whitelist, code review |
| Privilege Escalation | Low | Critical | Permission validation, sandboxing |
| Data Exfiltration | Medium | High | Network isolation, audit logging |
| Code Injection | Low | Critical | Input validation, content security policy |
| Dependency Confusion | Medium | Medium | Dependency verification, lock files |

#### Vulnerability Prevention Measures
```go
// Content Security Policy for plugin execution
type PluginCSP struct {
    AllowedOrigins []string
    AllowedActions []string
    BlockedAPIs    []string
}

func (csp *PluginCSP) GeneratePolicy() string {
    return fmt.Sprintf(`
        default-src 'self';
        script-src 'self' %s;
        connect-src 'self' %s;
        style-src 'self' 'unsafe-inline';
        img-src 'self' data: https:;
        font-src 'self';
        object-src 'none';
        base-uri 'self';
        form-action 'self';
    `, 
        strings.Join(csp.AllowedOrigins, " "),
        strings.Join(csp.AllowedOrigins, " "),
    )
}

// Plugin sandbox configuration
type PluginSandbox struct {
    NetworkAccess    bool
    FileSystemAccess []string
    DatabaseAccess   []string
    APIAccess        []string
    MemoryLimit      int64
    CPULimit         float64
    TimeoutLimit     time.Duration
}

func (s *PluginSandbox) CreateContainer(pluginPath string) (*docker.Container, error) {
    config := &container.Config{
        Image: "cloudbox/plugin-runtime:latest",
        WorkingDir: "/plugin",
        Env: []string{
            fmt.Sprintf("MEMORY_LIMIT=%d", s.MemoryLimit),
            fmt.Sprintf("CPU_LIMIT=%.2f", s.CPULimit),
            fmt.Sprintf("TIMEOUT_LIMIT=%d", int(s.TimeoutLimit.Seconds())),
        },
        NetworkDisabled: !s.NetworkAccess,
    }
    
    hostConfig := &container.HostConfig{
        Memory:    s.MemoryLimit,
        CPUQuota:  int64(s.CPULimit * 100000),
        CPUPeriod: 100000,
        ReadonlyRootfs: true,
        SecurityOpt: []string{
            "no-new-privileges:true",
            "seccomp:unconfined", // Will be restricted in production
        },
    }
    
    if len(s.FileSystemAccess) > 0 {
        for _, path := range s.FileSystemAccess {
            hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:ro", path, path))
        }
    }
    
    return docker.ContainerCreate(config, hostConfig, nil, nil, "")
}
```

### Security Testing Considerations

#### Penetration Testing Scenarios
1. **Authentication Bypass Attempts**
   - JWT token manipulation
   - Session hijacking
   - Privilege escalation

2. **Plugin Installation Attacks**
   - Malicious repository injection
   - Manifest tampering
   - Dependency poisoning

3. **Data Access Violations**
   - Unauthorized database access
   - Cross-project data leakage
   - Sensitive information exposure

4. **Injection Attacks**
   - SQL injection via plugin parameters
   - XSS through plugin UI components
   - Command injection in installation scripts

#### Security Audit Checklist
- [ ] Repository source validation
- [ ] Manifest integrity verification
- [ ] Permission scope limitation
- [ ] Network access restriction
- [ ] File system sandboxing
- [ ] Database access control
- [ ] API rate limiting
- [ ] Input sanitization
- [ ] Output encoding
- [ ] Audit logging
- [ ] Error handling security
- [ ] Session management
- [ ] HTTPS enforcement
- [ ] CSRF protection

---

## Critical Technical Constraints and Assumptions

### Assumptions
1. **Docker Environment**: CloudBox runs in containerized environment with orchestration capabilities
2. **PostgreSQL Database**: Primary database for plugin metadata and state management
3. **GitHub Access**: Internet connectivity for GitHub API and repository access
4. **Admin Authentication**: JWT-based authentication system already functional
5. **File System Access**: Writable plugin directory with appropriate permissions

### Technical Constraints
1. **Memory Limits**: Plugin execution limited to 512MB RAM per instance
2. **Network Restrictions**: Plugins run in isolated network namespace
3. **File System**: Read-only root file system with specific writable mounts
4. **Database Connections**: Shared connection pool with rate limiting
5. **API Rate Limits**: GitHub API rate limiting considerations
6. **Security Boundaries**: No direct system access or privilege escalation

### Implementation Phases
1. **Phase 1**: Database schema and basic CRUD operations (2-3 days)
2. **Phase 2**: GitHub integration and repository validation (3-4 days)
3. **Phase 3**: Secure installation engine with sandboxing (4-5 days)
4. **Phase 4**: Frontend marketplace and management UI (3-4 days)
5. **Phase 5**: Testing, security audit, and documentation (2-3 days)

**Total Estimated Timeline**: 14-19 days with 2-3 engineers

This architecture provides a robust, secure, and scalable foundation for the CloudBox plugin system that addresses all identified issues while maintaining high security standards and excellent user experience.