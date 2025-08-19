-- Migration: Create plugin registry and installation tracking tables
-- Created: 2024-08-17
-- Purpose: Complete plugin system infrastructure with registry, installations, and state persistence

-- Plugin registry - tracks all available plugins
CREATE TABLE IF NOT EXISTS plugin_registry (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    author VARCHAR(255) NOT NULL,
    repository VARCHAR(255) NOT NULL,
    license VARCHAR(50),
    type VARCHAR(50) NOT NULL DEFAULT 'dashboard-plugin', -- dashboard-plugin, api-plugin, service-plugin
    main_file VARCHAR(255),
    
    -- Security and validation
    checksum VARCHAR(64), -- SHA256 hash
    signature TEXT, -- Digital signature for verification
    is_verified BOOLEAN NOT NULL DEFAULT false,
    is_approved BOOLEAN NOT NULL DEFAULT false,
    
    -- Metadata
    permissions TEXT[], -- Array of permission strings
    dependencies JSONB DEFAULT '{}', -- JSON object of dependencies
    ui_config JSONB DEFAULT '{}', -- UI configuration for dashboard plugins
    
    -- Status and lifecycle
    status VARCHAR(20) NOT NULL DEFAULT 'available', -- available, deprecated, suspended
    download_count INTEGER NOT NULL DEFAULT 0,
    install_count INTEGER NOT NULL DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    deprecated_at TIMESTAMP,
    
    -- Registry metadata
    registry_source VARCHAR(100) NOT NULL DEFAULT 'github',
    source_metadata JSONB DEFAULT '{}' -- Additional source-specific data
);

-- Plugin installations - tracks installed plugins per project
CREATE TABLE IF NOT EXISTS plugin_installations (
    id SERIAL PRIMARY KEY,
    plugin_name VARCHAR(100) NOT NULL,
    plugin_version VARCHAR(50) NOT NULL,
    project_id INTEGER NOT NULL,
    
    -- Installation state
    status VARCHAR(20) NOT NULL DEFAULT 'disabled', -- enabled, disabled, installing, uninstalling, error
    installation_path VARCHAR(500),
    
    -- Installation metadata
    installed_by INTEGER NOT NULL, -- user_id who installed
    installed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_enabled_at TIMESTAMP,
    last_disabled_at TIMESTAMP,
    
    -- Configuration
    config JSONB DEFAULT '{}', -- Plugin-specific configuration
    environment JSONB DEFAULT '{}', -- Environment variables for this installation
    
    -- Error tracking
    error_message TEXT,
    last_error_at TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,
    
    -- Constraints
    UNIQUE(plugin_name, project_id) -- One installation per plugin per project
);

-- Plugin states - tracks real-time plugin states (for fast lookups)
CREATE TABLE IF NOT EXISTS plugin_states (
    id SERIAL PRIMARY KEY,
    plugin_name VARCHAR(100) NOT NULL,
    project_id INTEGER NOT NULL,
    
    -- Current state
    current_status VARCHAR(20) NOT NULL DEFAULT 'disabled',
    process_id INTEGER, -- System process ID if running
    port INTEGER, -- Port if plugin is running a service
    
    -- Health monitoring
    last_health_check TIMESTAMP,
    health_status VARCHAR(20) DEFAULT 'unknown', -- healthy, unhealthy, unknown
    health_details JSONB DEFAULT '{}',
    
    -- Performance metrics
    cpu_usage DECIMAL(5,2), -- Percentage
    memory_usage BIGINT, -- Bytes
    uptime_seconds INTEGER,
    
    -- State transitions
    state_changed_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    state_changed_by INTEGER, -- user_id who changed state
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    UNIQUE(plugin_name, project_id) -- One state per plugin per project
);

-- Approved repositories - dynamic whitelist management
CREATE TABLE IF NOT EXISTS approved_repositories (
    id SERIAL PRIMARY KEY,
    repository_url VARCHAR(255) NOT NULL UNIQUE,
    repository_owner VARCHAR(100) NOT NULL,
    repository_name VARCHAR(100) NOT NULL,
    
    -- Approval metadata
    approved_by INTEGER NOT NULL, -- user_id who approved
    approved_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    approval_reason TEXT,
    
    -- Repository metadata
    repository_type VARCHAR(50) NOT NULL DEFAULT 'github', -- github, gitlab, etc.
    is_official BOOLEAN NOT NULL DEFAULT false,
    trust_level INTEGER NOT NULL DEFAULT 1, -- 1=basic, 2=verified, 3=official
    
    -- Status
    is_active BOOLEAN NOT NULL DEFAULT true,
    last_validated_at TIMESTAMP,
    validation_status VARCHAR(20) DEFAULT 'pending', -- pending, valid, invalid, expired
    
    -- Security tracking
    security_issues JSONB DEFAULT '[]', -- Array of known security issues
    last_security_scan TIMESTAMP,
    
    -- Timestamps
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Plugin downloads - track plugin download/installation attempts
CREATE TABLE IF NOT EXISTS plugin_downloads (
    id SERIAL PRIMARY KEY,
    plugin_name VARCHAR(100) NOT NULL,
    plugin_version VARCHAR(50) NOT NULL,
    project_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    
    -- Download details
    download_source VARCHAR(255) NOT NULL, -- GitHub URL, registry URL, etc.
    download_status VARCHAR(20) NOT NULL DEFAULT 'started', -- started, completed, failed
    
    -- File information
    file_size BIGINT, -- Size in bytes
    download_time_ms INTEGER, -- Download time in milliseconds
    checksum_verified BOOLEAN DEFAULT false,
    signature_verified BOOLEAN DEFAULT false,
    
    -- Error tracking
    error_message TEXT,
    error_code VARCHAR(20),
    
    -- Network information
    client_ip INET,
    user_agent TEXT,
    
    -- Timestamps
    started_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    failed_at TIMESTAMP
);

-- Create plugin marketplace table for enhanced plugin discovery
CREATE TABLE IF NOT EXISTS plugin_marketplace (
    id SERIAL PRIMARY KEY,
    plugin_name VARCHAR(64) NOT NULL UNIQUE,
    repository VARCHAR(255) NOT NULL,
    version VARCHAR(32) NOT NULL,
    description TEXT,
    author VARCHAR(128),
    category VARCHAR(64),
    tags TEXT[], -- Array of tags for search and filtering
    license VARCHAR(32),
    website VARCHAR(255),
    support_email VARCHAR(128),
    screenshots TEXT[], -- Array of screenshot URLs
    demo_url VARCHAR(255),
    permissions TEXT[], -- Array of required permissions
    dependencies JSONB, -- Package dependencies as JSON
    pricing_model VARCHAR(32) DEFAULT 'free', -- free, paid, freemium, subscription
    price DECIMAL(10,2) DEFAULT 0.00,
    currency VARCHAR(3) DEFAULT 'USD',
    installation_count INTEGER DEFAULT 0,
    rating DECIMAL(3,2) DEFAULT 0.0, -- Average rating 0.0-5.0
    review_count INTEGER DEFAULT 0,
    featured BOOLEAN DEFAULT false,
    status VARCHAR(32) DEFAULT 'draft', -- draft, review, published, deprecated
    metadata JSONB, -- Additional metadata as JSON
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance
CREATE INDEX idx_plugin_registry_name ON plugin_registry(name);
CREATE INDEX idx_plugin_registry_status ON plugin_registry(status);
CREATE INDEX idx_plugin_registry_type ON plugin_registry(type);
CREATE INDEX idx_plugin_registry_created_at ON plugin_registry(created_at);
CREATE INDEX idx_plugin_registry_is_approved ON plugin_registry(is_approved);

CREATE INDEX idx_plugin_installations_project_id ON plugin_installations(project_id);
CREATE INDEX idx_plugin_installations_plugin_name ON plugin_installations(plugin_name);
CREATE INDEX idx_plugin_installations_status ON plugin_installations(status);
CREATE INDEX idx_plugin_installations_installed_by ON plugin_installations(installed_by);
CREATE INDEX idx_plugin_installations_unique_plugin_project ON plugin_installations(plugin_name, project_id);

CREATE INDEX idx_plugin_states_project_id ON plugin_states(project_id);
CREATE INDEX idx_plugin_states_plugin_name ON plugin_states(plugin_name);
CREATE INDEX idx_plugin_states_current_status ON plugin_states(current_status);
CREATE INDEX idx_plugin_states_health_status ON plugin_states(health_status);
CREATE INDEX idx_plugin_states_unique_plugin_project ON plugin_states(plugin_name, project_id);

CREATE INDEX idx_approved_repositories_url ON approved_repositories(repository_url);
CREATE INDEX idx_approved_repositories_owner ON approved_repositories(repository_owner);
CREATE INDEX idx_approved_repositories_active ON approved_repositories(is_active);
CREATE INDEX idx_approved_repositories_trust_level ON approved_repositories(trust_level);

CREATE INDEX idx_plugin_downloads_plugin_name ON plugin_downloads(plugin_name);
CREATE INDEX idx_plugin_downloads_project_id ON plugin_downloads(project_id);
CREATE INDEX idx_plugin_downloads_user_id ON plugin_downloads(user_id);
CREATE INDEX idx_plugin_downloads_status ON plugin_downloads(download_status);
CREATE INDEX idx_plugin_downloads_started_at ON plugin_downloads(started_at);

CREATE INDEX idx_plugin_marketplace_name ON plugin_marketplace(plugin_name);
CREATE INDEX idx_plugin_marketplace_category ON plugin_marketplace(category);
CREATE INDEX idx_plugin_marketplace_status ON plugin_marketplace(status);
CREATE INDEX idx_plugin_marketplace_featured ON plugin_marketplace(featured);
CREATE INDEX idx_plugin_marketplace_pricing_model ON plugin_marketplace(pricing_model);
CREATE INDEX idx_plugin_marketplace_rating ON plugin_marketplace(rating);

-- Create updated_at triggers for automatic timestamp management
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language plpgsql;

CREATE TRIGGER update_plugin_registry_updated_at BEFORE UPDATE ON plugin_registry FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugin_installations_updated_at BEFORE UPDATE ON plugin_installations FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugin_states_updated_at BEFORE UPDATE ON plugin_states FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_approved_repositories_updated_at BEFORE UPDATE ON approved_repositories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_plugin_marketplace_updated_at BEFORE UPDATE ON plugin_marketplace FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Insert default approved repositories
INSERT INTO approved_repositories (repository_url, repository_owner, repository_name, approved_by, approval_reason, is_official, trust_level)
VALUES 
    ('https://github.com/cloudbox/plugins', 'cloudbox', 'plugins', 1, 'Official CloudBox plugin repository', true, 3),
    ('https://github.com/cloudbox/official-plugins', 'cloudbox', 'official-plugins', 1, 'Official CloudBox plugins collection', true, 3),
    ('https://github.com/cloudbox/community-plugins', 'cloudbox', 'community-plugins', 1, 'Community-verified plugin repository', false, 2),
    ('https://github.com/cloudbox-org/plugins', 'cloudbox-org', 'plugins', 1, 'CloudBox organization plugin repository', true, 3)
ON CONFLICT (repository_url) DO NOTHING;

-- Add comments for documentation
COMMENT ON TABLE plugin_registry IS 'Registry of all available plugins with metadata and security information';
COMMENT ON TABLE plugin_installations IS 'Tracks plugin installations per project with configuration and state';
COMMENT ON TABLE plugin_states IS 'Real-time plugin states for fast status lookups and health monitoring';
COMMENT ON TABLE approved_repositories IS 'Dynamic whitelist of approved plugin source repositories';
COMMENT ON TABLE plugin_downloads IS 'Audit trail of plugin download and installation attempts';
COMMENT ON TABLE plugin_marketplace IS 'Enhanced marketplace entries for plugin discovery with rich metadata';

COMMENT ON COLUMN plugin_registry.permissions IS 'Array of permission strings required by the plugin';
COMMENT ON COLUMN plugin_registry.dependencies IS 'JSON object of plugin dependencies and versions';
COMMENT ON COLUMN plugin_registry.ui_config IS 'UI configuration for dashboard integration';
COMMENT ON COLUMN plugin_installations.config IS 'Plugin-specific configuration for this installation';
COMMENT ON COLUMN plugin_installations.environment IS 'Environment variables for this plugin installation';
COMMENT ON COLUMN plugin_states.health_details IS 'Detailed health check results and metrics';
COMMENT ON COLUMN approved_repositories.trust_level IS 'Repository trust level: 1=basic, 2=verified, 3=official';
COMMENT ON COLUMN approved_repositories.security_issues IS 'Array of known security issues for this repository';
COMMENT ON COLUMN plugin_marketplace.tags IS 'Array of tags for search and filtering';
COMMENT ON COLUMN plugin_marketplace.dependencies IS 'Package dependencies as JSON';
COMMENT ON COLUMN plugin_marketplace.metadata IS 'Additional metadata as JSON';

-- Grant appropriate permissions (uncomment when setting up production database)
-- GRANT SELECT, INSERT, UPDATE ON plugin_registry TO cloudbox_app;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON plugin_installations TO cloudbox_app;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON plugin_states TO cloudbox_app;
-- GRANT SELECT ON approved_repositories TO cloudbox_app;
-- GRANT SELECT, INSERT ON plugin_downloads TO cloudbox_app;
-- GRANT SELECT, INSERT, UPDATE ON plugin_marketplace TO cloudbox_app;
-- GRANT USAGE ON ALL SEQUENCES IN SCHEMA public TO cloudbox_app;