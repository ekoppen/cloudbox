# Plugin Management Architecture: Superadmin vs Project Admin Separation

## Executive Summary

This document outlines the architectural design for separating plugin management responsibilities between superadmins (global marketplace management) and project admins (per-project plugin usage). The design creates a clear separation of concerns while maintaining security and usability.

### Key Architectural Decisions

- **Superadmin Level**: Global plugin marketplace management, security approval, and system-wide availability
- **Project Admin Level**: Per-project plugin installation, activation, and configuration
- **Technology Stack**: Go backend with role-based permissions, PostgreSQL for data persistence
- **Security Model**: Hierarchical permissions with audit trails and validation workflows

## For Backend Engineers

### Database Schema Changes

```sql
-- New table for project-specific plugin configurations
CREATE TABLE project_plugin_settings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,
    
    project_id INTEGER NOT NULL REFERENCES projects(id),
    plugin_name VARCHAR(255) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT false,
    configuration JSONB DEFAULT '{}',
    environment_variables JSONB DEFAULT '{}',
    
    -- Installation metadata
    installed_by INTEGER NOT NULL REFERENCES users(id),
    installed_at TIMESTAMP NOT NULL DEFAULT NOW(),
    last_activated_at TIMESTAMP,
    last_deactivated_at TIMESTAMP,
    
    -- Versioning
    installed_version VARCHAR(100),
    available_version VARCHAR(100),
    
    -- Error tracking
    last_error TEXT,
    last_error_at TIMESTAMP,
    
    UNIQUE(project_id, plugin_name)
);

-- Index for performance
CREATE INDEX idx_project_plugin_settings_project_id ON project_plugin_settings(project_id);
CREATE INDEX idx_project_plugin_settings_active ON project_plugin_settings(project_id, is_active);

-- Enhanced marketplace table for better separation
ALTER TABLE plugin_marketplace ADD COLUMN approval_status VARCHAR(50) DEFAULT 'pending';
ALTER TABLE plugin_marketplace ADD COLUMN approved_by INTEGER REFERENCES users(id);
ALTER TABLE plugin_marketplace ADD COLUMN approved_at TIMESTAMP;
ALTER TABLE plugin_marketplace ADD COLUMN system_wide_available BOOLEAN DEFAULT false;

-- Project admin permissions tracking
CREATE TABLE project_admin_permissions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    project_id INTEGER NOT NULL REFERENCES projects(id),
    user_id INTEGER NOT NULL REFERENCES users(id),
    permission_type VARCHAR(100) NOT NULL, -- 'plugin_management', 'user_management', etc.
    granted_by INTEGER NOT NULL REFERENCES users(id),
    granted_at TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(project_id, user_id, permission_type)
);
```

### API Endpoint Specifications

#### Superadmin Endpoints (Global Marketplace Management)

```go
// Superadmin marketplace management
api.Group("/admin/marketplace")
{
    // Global plugin approval workflow
    POST   /admin/marketplace/plugins/submit          // Submit plugin for approval
    GET    /admin/marketplace/plugins/pending         // List pending approvals
    POST   /admin/marketplace/plugins/:id/approve     // Approve plugin
    POST   /admin/marketplace/plugins/:id/reject      // Reject plugin
    
    // Global marketplace management
    GET    /admin/marketplace/plugins                 // List all marketplace plugins
    PUT    /admin/marketplace/plugins/:id             // Update plugin metadata
    DELETE /admin/marketplace/plugins/:id             // Remove from marketplace
    
    // Repository management
    GET    /admin/marketplace/repositories            // List approved repositories
    POST   /admin/marketplace/repositories            // Add approved repository
    DELETE /admin/marketplace/repositories/:id        // Remove repository
    
    // System-wide availability
    PUT    /admin/marketplace/plugins/:id/availability // Set system-wide availability
}
```

#### Project Admin Endpoints (Per-Project Plugin Management)

```go
// Project-level plugin management
api.Group("/projects/:project_id/plugins")
{
    // Available plugins for this project
    GET    /projects/:project_id/plugins/available    // List available plugins from marketplace
    GET    /projects/:project_id/plugins/installed    // List installed plugins for project
    
    // Plugin lifecycle management
    POST   /projects/:project_id/plugins/:plugin_name/install    // Install plugin to project
    POST   /projects/:project_id/plugins/:plugin_name/activate   // Activate plugin
    POST   /projects/:project_id/plugins/:plugin_name/deactivate // Deactivate plugin
    DELETE /projects/:project_id/plugins/:plugin_name            // Uninstall plugin
    
    // Plugin configuration
    GET    /projects/:project_id/plugins/:plugin_name/config     // Get plugin config
    PUT    /projects/:project_id/plugins/:plugin_name/config     // Update plugin config
    
    // Plugin status and health
    GET    /projects/:project_id/plugins/:plugin_name/status     // Get plugin status
    GET    /projects/:project_id/plugins/:plugin_name/logs       // Get plugin logs
}
```

#### Enhanced API Request/Response Schemas

**Install Plugin Request:**
```json
{
    "plugin_name": "cloudbox-analytics",
    "version": "1.2.0",
    "configuration": {
        "api_endpoint": "https://analytics.example.com",
        "tracking_enabled": true
    },
    "environment_variables": {
        "ANALYTICS_API_KEY": "secret_key_here"
    }
}
```

**Plugin Status Response:**
```json
{
    "success": true,
    "plugin": {
        "name": "cloudbox-analytics",
        "version": "1.2.0",
        "status": "active",
        "health": "healthy",
        "installed_at": "2024-08-17T10:00:00Z",
        "last_activated_at": "2024-08-17T11:00:00Z",
        "configuration": {
            "api_endpoint": "https://analytics.example.com",
            "tracking_enabled": true
        },
        "metrics": {
            "uptime": "2h 30m",
            "memory_usage": "45MB",
            "cpu_usage": "2.1%"
        }
    }
}
```

### Authentication and Authorization Implementation

```go
// Enhanced middleware for project-level plugin management
func RequireProjectPluginPermission() gin.HandlerFunc {
    return func(c *gin.Context) {
        userRole := c.GetString("user_role")
        projectID := c.Param("project_id")
        userID := c.GetString("user_id")
        
        // Superadmins have access to all projects
        if userRole == "superadmin" {
            c.Next()
            return
        }
        
        // Check if user is project admin with plugin permissions
        hasPermission := checkProjectAdminPermission(userID, projectID, "plugin_management")
        if !hasPermission {
            c.JSON(http.StatusForbidden, gin.H{
                "success": false,
                "error": "Insufficient permissions for plugin management",
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Business logic for plugin installation
func (h *ProjectPluginHandler) InstallPlugin(c *gin.Context) {
    projectID := c.Param("project_id")
    pluginName := c.Param("plugin_name")
    
    // 1. Validate plugin exists in marketplace
    plugin, err := h.getMarketplacePlugin(pluginName)
    if err != nil {
        c.JSON(http.StatusNotFound, gin.H{
            "success": false,
            "error": "Plugin not found in marketplace",
        })
        return
    }
    
    // 2. Check if plugin is approved for system-wide use
    if !plugin.SystemWideAvailable {
        c.JSON(http.StatusForbidden, gin.H{
            "success": false,
            "error": "Plugin not approved for installation",
        })
        return
    }
    
    // 3. Check if already installed
    existing, _ := h.getProjectPlugin(projectID, pluginName)
    if existing != nil {
        c.JSON(http.StatusConflict, gin.H{
            "success": false,
            "error": "Plugin already installed",
        })
        return
    }
    
    // 4. Create project plugin record
    projectPlugin := models.ProjectPluginSetting{
        ProjectID:         projectID,
        PluginName:       pluginName,
        InstalledVersion: plugin.Version,
        InstalledBy:      c.GetString("user_id"),
        IsActive:         false, // Installed but not activated
        Configuration:    map[string]interface{}{},
        EnvironmentVariables: map[string]interface{}{},
    }
    
    if err := h.db.Create(&projectPlugin).Error; err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "success": false,
            "error": "Failed to install plugin",
        })
        return
    }
    
    // 5. Audit log
    h.auditPluginAction("install", pluginName, projectID, c)
    
    c.JSON(http.StatusOK, gin.H{
        "success": true,
        "message": "Plugin installed successfully",
        "plugin": projectPlugin,
    })
}
```

### Error Handling and Validation Strategies

```go
// Enhanced validation for plugin operations
type PluginValidator struct {
    db *gorm.DB
}

func (v *PluginValidator) ValidatePluginInstallation(projectID, pluginName string) error {
    // 1. Validate plugin exists in marketplace
    var plugin models.PluginMarketplace
    if err := v.db.Where("plugin_name = ? AND status = 'active'", pluginName).First(&plugin).Error; err != nil {
        return fmt.Errorf("plugin not found in marketplace: %s", pluginName)
    }
    
    // 2. Check if plugin is system-wide available
    if !plugin.SystemWideAvailable {
        return fmt.Errorf("plugin not approved for system-wide installation")
    }
    
    // 3. Validate project permissions
    // Additional validation logic here
    
    return nil
}

func (v *PluginValidator) ValidatePluginConfiguration(config map[string]interface{}, pluginName string) error {
    // Implement plugin-specific configuration validation
    // This could include schema validation, security checks, etc.
    return nil
}
```

## For Frontend Engineers

### Component Architecture and State Management

#### Project Plugin Management Dashboard

```typescript
// Project Plugin Settings Page Component Structure
components/
├── project/
│   ├── settings/
│   │   ├── plugins/
│   │   │   ├── PluginDashboard.tsx          // Main plugin management interface
│   │   │   ├── AvailablePluginsList.tsx     // Browse marketplace plugins
│   │   │   ├── InstalledPluginsList.tsx     // Manage installed plugins
│   │   │   ├── PluginCard.tsx               // Individual plugin display
│   │   │   ├── PluginConfigModal.tsx        // Plugin configuration interface
│   │   │   ├── PluginStatusIndicator.tsx    // Health and status display
│   │   │   └── PluginInstallModal.tsx       // Installation confirmation
│   │   └── PluginSettingsPage.tsx           // Settings page wrapper
│   └── dashboard/
│       └── ProjectNavigation.tsx             // Updated navigation with plugins
```

#### State Management Approach

```typescript
// Redux store structure for plugin management
interface PluginState {
  // Available plugins from marketplace
  marketplace: {
    plugins: MarketplacePlugin[];
    loading: boolean;
    error: string | null;
  };
  
  // Project-specific installed plugins
  installed: {
    [projectId: string]: {
      plugins: ProjectPlugin[];
      loading: boolean;
      error: string | null;
    };
  };
  
  // Plugin operation states
  operations: {
    installing: string[]; // plugin names being installed
    activating: string[]; // plugin names being activated
    configuring: string | null; // plugin being configured
  };
}

// Action creators for plugin management
const pluginSlice = createSlice({
  name: 'plugins',
  initialState,
  reducers: {
    // Marketplace actions
    fetchMarketplaceStart: (state) => {
      state.marketplace.loading = true;
      state.marketplace.error = null;
    },
    fetchMarketplaceSuccess: (state, action) => {
      state.marketplace.plugins = action.payload;
      state.marketplace.loading = false;
    },
    
    // Project plugin actions
    installPluginStart: (state, action) => {
      state.operations.installing.push(action.payload.pluginName);
    },
    installPluginSuccess: (state, action) => {
      const { projectId, plugin } = action.payload;
      if (!state.installed[projectId]) {
        state.installed[projectId] = { plugins: [], loading: false, error: null };
      }
      state.installed[projectId].plugins.push(plugin);
      state.operations.installing = state.operations.installing.filter(
        name => name !== plugin.name
      );
    },
    
    // Configuration actions
    updatePluginConfig: (state, action) => {
      const { projectId, pluginName, config } = action.payload;
      const plugin = state.installed[projectId]?.plugins.find(p => p.name === pluginName);
      if (plugin) {
        plugin.configuration = config;
      }
    }
  }
});
```

#### API Integration Patterns

```typescript
// API service for project plugin management
class ProjectPluginService {
  private baseUrl: string;
  
  constructor(baseUrl: string) {
    this.baseUrl = baseUrl;
  }
  
  // Get available plugins for a project
  async getAvailablePlugins(projectId: string): Promise<MarketplacePlugin[]> {
    const response = await fetch(`${this.baseUrl}/projects/${projectId}/plugins/available`, {
      headers: this.getAuthHeaders()
    });
    return response.json();
  }
  
  // Install plugin to project
  async installPlugin(projectId: string, pluginName: string, config: PluginInstallConfig): Promise<ProjectPlugin> {
    const response = await fetch(`${this.baseUrl}/projects/${projectId}/plugins/${pluginName}/install`, {
      method: 'POST',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(config)
    });
    return response.json();
  }
  
  // Activate/deactivate plugin
  async togglePlugin(projectId: string, pluginName: string, active: boolean): Promise<void> {
    const endpoint = active ? 'activate' : 'deactivate';
    const response = await fetch(`${this.baseUrl}/projects/${projectId}/plugins/${pluginName}/${endpoint}`, {
      method: 'POST',
      headers: this.getAuthHeaders()
    });
    
    if (!response.ok) {
      throw new Error(`Failed to ${endpoint} plugin`);
    }
  }
  
  // Update plugin configuration
  async updatePluginConfig(projectId: string, pluginName: string, config: PluginConfig): Promise<void> {
    const response = await fetch(`${this.baseUrl}/projects/${projectId}/plugins/${pluginName}/config`, {
      method: 'PUT',
      headers: this.getAuthHeaders(),
      body: JSON.stringify(config)
    });
    
    if (!response.ok) {
      throw new Error('Failed to update plugin configuration');
    }
  }
  
  private getAuthHeaders(): Record<string, string> {
    return {
      'Content-Type': 'application/json',
      'Authorization': `Bearer ${localStorage.getItem('auth_token')}`,
    };
  }
}
```

#### Routing and Navigation Strategy

```typescript
// Updated routing for project plugin management
const projectRoutes = [
  {
    path: '/dashboard/projects/:projectId',
    component: ProjectDashboard,
    children: [
      {
        path: 'settings',
        component: ProjectSettings,
        children: [
          {
            path: 'plugins',
            component: PluginSettingsPage,
            meta: { 
              requiresAuth: true,
              requiredPermission: 'project_admin'
            }
          }
        ]
      }
    ]
  }
];

// Navigation component updates
const ProjectNavigation: React.FC<{ projectId: string }> = ({ projectId }) => {
  const { hasPermission } = useAuth();
  
  return (
    <nav>
      <NavLink to={`/dashboard/projects/${projectId}`}>Dashboard</NavLink>
      <NavLink to={`/dashboard/projects/${projectId}/data`}>Data</NavLink>
      <NavLink to={`/dashboard/projects/${projectId}/functions`}>Functions</NavLink>
      
      {hasPermission('project_admin') && (
        <NavSection title="Settings">
          <NavLink to={`/dashboard/projects/${projectId}/settings/general`}>General</NavLink>
          <NavLink to={`/dashboard/projects/${projectId}/settings/plugins`}>Plugins</NavLink>
          <NavLink to={`/dashboard/projects/${projectId}/settings/team`}>Team</NavLink>
        </NavSection>
      )}
    </nav>
  );
};
```

### Performance Optimization Strategies

```typescript
// Optimized plugin loading with React Query
const useProjectPlugins = (projectId: string) => {
  return useQuery({
    queryKey: ['projectPlugins', projectId],
    queryFn: () => projectPluginService.getInstalledPlugins(projectId),
    staleTime: 5 * 60 * 1000, // 5 minutes
    cacheTime: 10 * 60 * 1000, // 10 minutes
  });
};

// Optimistic updates for plugin activation
const useTogglePlugin = (projectId: string) => {
  const queryClient = useQueryClient();
  
  return useMutation({
    mutationFn: ({ pluginName, active }: { pluginName: string; active: boolean }) =>
      projectPluginService.togglePlugin(projectId, pluginName, active),
    
    onMutate: async ({ pluginName, active }) => {
      // Cancel outgoing refetches
      await queryClient.cancelQueries(['projectPlugins', projectId]);
      
      // Snapshot previous value
      const previousPlugins = queryClient.getQueryData(['projectPlugins', projectId]);
      
      // Optimistically update
      queryClient.setQueryData(['projectPlugins', projectId], (old: ProjectPlugin[]) =>
        old?.map(plugin =>
          plugin.name === pluginName ? { ...plugin, is_active: active } : plugin
        )
      );
      
      return { previousPlugins };
    },
    
    onError: (err, newTodo, context) => {
      // Rollback on error
      queryClient.setQueryData(['projectPlugins', projectId], context?.previousPlugins);
    },
    
    onSettled: () => {
      // Refetch after error or success
      queryClient.invalidateQueries(['projectPlugins', projectId]);
    },
  });
};
```

## For QA Engineers

### Testable Component Boundaries and Interfaces

#### Unit Testing Strategy

```typescript
// Test plugin management components
describe('PluginDashboard', () => {
  const mockProjectId = 'project-123';
  const mockPlugins = [
    { name: 'analytics', status: 'active', version: '1.0.0' },
    { name: 'backup', status: 'inactive', version: '2.1.0' }
  ];
  
  beforeEach(() => {
    // Mock API responses
    jest.spyOn(projectPluginService, 'getInstalledPlugins')
      .mockResolvedValue(mockPlugins);
  });
  
  test('renders installed plugins correctly', async () => {
    render(<PluginDashboard projectId={mockProjectId} />);
    
    await waitFor(() => {
      expect(screen.getByText('analytics')).toBeInTheDocument();
      expect(screen.getByText('backup')).toBeInTheDocument();
    });
  });
  
  test('handles plugin activation', async () => {
    const toggleSpy = jest.spyOn(projectPluginService, 'togglePlugin')
      .mockResolvedValue();
    
    render(<PluginDashboard projectId={mockProjectId} />);
    
    const toggleButton = await screen.findByTestId('toggle-analytics');
    fireEvent.click(toggleButton);
    
    expect(toggleSpy).toHaveBeenCalledWith(mockProjectId, 'analytics', false);
  });
});
```

#### Integration Testing Focus Areas

```yaml
integration_tests:
  plugin_lifecycle:
    - test: "Install plugin flow"
      steps:
        - Navigate to project plugins page
        - Select plugin from marketplace
        - Configure plugin settings
        - Confirm installation
        - Verify plugin appears in installed list
    
    - test: "Plugin activation/deactivation"
      steps:
        - Install test plugin
        - Activate plugin via toggle
        - Verify status change in UI
        - Verify API call made correctly
        - Deactivate plugin
        - Verify status reverted
  
  permission_boundaries:
    - test: "Project admin access"
      users: ["project_admin"]
      expected: "Can access plugin management"
    
    - test: "Regular user access"
      users: ["regular_user"]
      expected: "Cannot access plugin management"
    
    - test: "Superadmin access"
      users: ["superadmin"]
      expected: "Can access all project plugin management"
```

#### Error Scenario Testing

```typescript
// Error handling test scenarios
describe('Plugin Error Handling', () => {
  test('handles plugin installation failure', async () => {
    jest.spyOn(projectPluginService, 'installPlugin')
      .mockRejectedValue(new Error('Installation failed'));
    
    render(<PluginInstallModal pluginName="test-plugin" projectId="123" />);
    
    const installButton = screen.getByRole('button', { name: /install/i });
    fireEvent.click(installButton);
    
    await waitFor(() => {
      expect(screen.getByText(/installation failed/i)).toBeInTheDocument();
    });
  });
  
  test('shows appropriate error for insufficient permissions', async () => {
    jest.spyOn(projectPluginService, 'installPlugin')
      .mockRejectedValue(new Error('Insufficient permissions'));
    
    // Test implementation
  });
});
```

### Data Validation Requirements

```yaml
validation_rules:
  plugin_configuration:
    - field: "plugin_name"
      rules: ["required", "alphanumeric_dash", "max_length:100"]
    - field: "configuration"
      rules: ["valid_json", "schema_validation"]
    - field: "environment_variables"
      rules: ["valid_json", "no_sensitive_keys_in_plain_text"]
  
  installation_request:
    - field: "project_id"
      rules: ["required", "exists:projects,id"]
    - field: "plugin_name"
      rules: ["required", "exists:plugin_marketplace,plugin_name"]
    - field: "version"
      rules: ["semantic_version"]
```

### Performance Benchmarks and Quality Metrics

```yaml
performance_targets:
  page_load:
    plugin_dashboard: "<2s initial load"
    marketplace_browse: "<1s with pagination"
  
  operations:
    plugin_install: "<10s end-to-end"
    plugin_activate: "<2s response time"
    config_update: "<1s response time"
  
  api_response_times:
    get_available_plugins: "<500ms"
    get_installed_plugins: "<300ms"
    install_plugin: "<5s"
    toggle_plugin: "<1s"

quality_metrics:
  test_coverage: ">90% for plugin-related code"
  error_rate: "<1% for plugin operations"
  accessibility: "WCAG 2.1 AA compliance"
  user_satisfaction: ">4.5/5 for plugin management UX"
```

## For Security Analysts

### Authentication Flow and Security Model

#### Role-Based Access Control (RBAC)

```yaml
roles:
  superadmin:
    permissions:
      - manage_marketplace_plugins
      - approve_plugin_submissions
      - manage_approved_repositories
      - access_all_project_plugins
      - view_security_audit_logs
  
  project_admin:
    permissions:
      - install_approved_plugins
      - configure_project_plugins
      - activate_deactivate_plugins
      - view_project_plugin_logs
  
  project_user:
    permissions:
      - view_active_plugins
      - use_plugin_features
```

#### Security Architecture

```go
// Security middleware for plugin operations
func PluginSecurityMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 1. Rate limiting for plugin operations
        if !rateLimiter.Allow(c.ClientIP()) {
            c.JSON(http.StatusTooManyRequests, gin.H{
                "error": "Rate limit exceeded",
            })
            c.Abort()
            return
        }
        
        // 2. CSRF protection for state-changing operations
        if c.Request.Method != "GET" {
            if !validateCSRFToken(c) {
                c.JSON(http.StatusForbidden, gin.H{
                    "error": "Invalid CSRF token",
                })
                c.Abort()
                return
            }
        }
        
        // 3. Input validation and sanitization
        if err := validatePluginInput(c); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{
                "error": "Invalid input: " + err.Error(),
            })
            c.Abort()
            return
        }
        
        c.Next()
    }
}

// Plugin security validator
type PluginSecurityValidator struct {
    db *gorm.DB
}

func (v *PluginSecurityValidator) ValidatePluginSecurity(pluginName string) error {
    // 1. Check plugin is from approved repository
    var plugin models.PluginMarketplace
    if err := v.db.Where("plugin_name = ?", pluginName).First(&plugin).Error; err != nil {
        return fmt.Errorf("plugin not found in marketplace")
    }
    
    // 2. Verify plugin signature
    if !v.verifyPluginSignature(plugin) {
        return fmt.Errorf("plugin signature verification failed")
    }
    
    // 3. Check for known vulnerabilities
    if vulnerabilities := v.checkVulnerabilities(plugin); len(vulnerabilities) > 0 {
        return fmt.Errorf("plugin has known vulnerabilities: %v", vulnerabilities)
    }
    
    return nil
}
```

#### Data Encryption Strategies

```go
// Encrypt sensitive plugin configuration
func (h *PluginHandler) encryptSensitiveConfig(config map[string]interface{}) error {
    sensitiveKeys := []string{"api_key", "secret", "password", "token"}
    
    for key, value := range config {
        if isSensitiveKey(key, sensitiveKeys) {
            encrypted, err := h.encryption.Encrypt(fmt.Sprintf("%v", value))
            if err != nil {
                return err
            }
            config[key] = encrypted
        }
    }
    
    return nil
}

// Plugin configuration encryption
type PluginConfigEncryption struct {
    key []byte
}

func (e *PluginConfigEncryption) Encrypt(data string) (string, error) {
    block, err := aes.NewCipher(e.key)
    if err != nil {
        return "", err
    }
    
    gcm, err := cipher.NewGCM(block)
    if err != nil {
        return "", err
    }
    
    nonce := make([]byte, gcm.NonceSize())
    if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
        return "", err
    }
    
    ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)
    return base64.StdEncoding.EncodeToString(ciphertext), nil
}
```

#### Security Headers and CORS Policies

```go
// Enhanced CORS for plugin API endpoints
func PluginCORS() gin.HandlerFunc {
    return cors.New(cors.Config{
        AllowOrigins: []string{
            "https://dashboard.cloudbox.com",
            "https://admin.cloudbox.com",
        },
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders: []string{
            "Origin",
            "Content-Type",
            "Accept",
            "Authorization",
            "X-CSRF-Token",
            "X-Request-ID",
        },
        ExposeHeaders: []string{
            "X-Rate-Limit-Remaining",
            "X-Rate-Limit-Reset",
        },
        AllowCredentials: true,
        MaxAge: 12 * time.Hour,
    })
}

// Security headers for plugin endpoints
func PluginSecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
        c.Header("Content-Security-Policy", "default-src 'self'; script-src 'self' 'unsafe-inline'")
        c.Next()
    }
}
```

#### Vulnerability Prevention Measures

```yaml
security_measures:
  input_validation:
    - plugin_name: "Alphanumeric + dash/underscore only, max 100 chars"
    - configuration: "JSON schema validation, no script injection"
    - repository_url: "HTTPS only, whitelist approved domains"
  
  access_control:
    - authentication: "JWT tokens with short expiry (15min)"
    - authorization: "Role-based with principle of least privilege"
    - session_management: "Secure cookie settings, CSRF protection"
  
  data_protection:
    - encryption_at_rest: "AES-256-GCM for sensitive config data"
    - encryption_in_transit: "TLS 1.3 for all API communications"
    - secret_management: "Separate encrypted storage for API keys"
  
  audit_logging:
    - security_events: "All plugin operations logged with user context"
    - retention_policy: "90 days for security logs, 30 days for operational"
    - monitoring: "Real-time alerts for suspicious activities"
```

### Monitoring and Observability Requirements

```go
// Security monitoring for plugin operations
type PluginSecurityMonitor struct {
    logger *logrus.Logger
    alerts chan SecurityAlert
}

func (m *PluginSecurityMonitor) LogPluginOperation(operation, pluginName, userID, projectID string, success bool) {
    logEntry := m.logger.WithFields(logrus.Fields{
        "operation":   operation,
        "plugin_name": pluginName,
        "user_id":     userID,
        "project_id":  projectID,
        "success":     success,
        "timestamp":   time.Now().UTC(),
        "ip_address":  getCurrentIP(),
    })
    
    if success {
        logEntry.Info("Plugin operation successful")
    } else {
        logEntry.Warn("Plugin operation failed")
        
        // Trigger security alert for failed operations
        m.alerts <- SecurityAlert{
            Type:        "PluginOperationFailed",
            Severity:    "Medium",
            UserID:      userID,
            Operation:   operation,
            PluginName:  pluginName,
            Timestamp:   time.Now(),
        }
    }
}
```

## Implementation Timeline and Milestones

### Phase 1: Database and Backend Foundation (Week 1-2)
- [ ] Create new database tables for project plugin settings
- [ ] Implement enhanced models and migrations
- [ ] Create new API endpoints for project plugin management
- [ ] Implement role-based authentication middleware
- [ ] Add audit logging system

### Phase 2: Core Plugin Management Logic (Week 3-4)
- [ ] Implement plugin installation/uninstallation logic
- [ ] Create plugin activation/deactivation functionality
- [ ] Build plugin configuration management
- [ ] Implement security validation and encryption
- [ ] Add comprehensive error handling

### Phase 3: Frontend Implementation (Week 5-6)
- [ ] Create project plugin management components
- [ ] Implement marketplace browsing interface
- [ ] Build plugin configuration modals
- [ ] Add real-time status indicators
- [ ] Implement optimistic updates and loading states

### Phase 4: Security and Testing (Week 7-8)
- [ ] Comprehensive security audit and penetration testing
- [ ] Unit and integration testing for all components
- [ ] Performance optimization and load testing
- [ ] Documentation and deployment guides
- [ ] User acceptance testing

### Critical Success Factors
1. **Clear separation of concerns** between superadmin and project admin roles
2. **Robust security measures** for plugin validation and execution
3. **Intuitive user experience** for plugin management workflows
4. **Comprehensive audit trails** for security monitoring
5. **Scalable architecture** for future plugin ecosystem growth

This architecture provides a solid foundation for separating plugin management responsibilities while maintaining security, usability, and scalability requirements.