-- Initial CloudBox database schema

-- Enable UUID extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Users table
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    is_active BOOLEAN DEFAULT true,
    last_login_at TIMESTAMP WITH TIME ZONE
);

-- Projects table
CREATE TABLE projects (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    slug VARCHAR(255) UNIQUE NOT NULL,
    is_active BOOLEAN DEFAULT true,
    
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE
);

-- API Keys table
CREATE TABLE api_keys (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    key VARCHAR(255) UNIQUE NOT NULL,
    key_hash VARCHAR(255) NOT NULL,
    is_active BOOLEAN DEFAULT true,
    last_used_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    
    permissions TEXT[],
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- CORS Configuration table
CREATE TABLE cors_configs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    allowed_origins TEXT[],
    allowed_methods TEXT[],
    allowed_headers TEXT[],
    exposed_headers TEXT[],
    allow_credentials BOOLEAN DEFAULT false,
    max_age INTEGER DEFAULT 3600,
    
    project_id INTEGER UNIQUE NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Deployments table
CREATE TABLE deployments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    version VARCHAR(255) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    build_logs TEXT,
    deployed_at TIMESTAMP WITH TIME ZONE,
    domain VARCHAR(255),
    environment JSONB,
    
    file_count BIGINT,
    total_size BIGINT,
    build_time BIGINT,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Backups table
CREATE TABLE backups (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    
    size BIGINT,
    file_path TEXT,
    checksum VARCHAR(255),
    completed_at TIMESTAMP WITH TIME ZONE,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Indexes for better performance
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_projects_user_id ON projects(user_id);
CREATE INDEX idx_projects_slug ON projects(slug);
CREATE INDEX idx_api_keys_project_id ON api_keys(project_id);
CREATE INDEX idx_api_keys_key ON api_keys(key);
CREATE INDEX idx_deployments_project_id ON deployments(project_id);
CREATE INDEX idx_backups_project_id ON backups(project_id);

-- Trigger function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- GitHub Repositories table
CREATE TABLE git_hub_repositories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL,
    clone_url TEXT NOT NULL,
    description TEXT,
    branch VARCHAR(255) DEFAULT 'main',
    is_private BOOLEAN DEFAULT false,
    
    -- Build configuration
    build_command VARCHAR(500),
    start_command VARCHAR(500),
    app_port INTEGER DEFAULT 3000,
    
    -- Status tracking
    is_active BOOLEAN DEFAULT true,
    last_sync_at TIMESTAMP WITH TIME ZONE,
    last_commit_hash VARCHAR(255) DEFAULT '',
    
    -- Pending updates (for notification system)
    pending_commit_hash VARCHAR(255) DEFAULT '',
    pending_commit_branch VARCHAR(255) DEFAULT '',
    has_pending_update BOOLEAN DEFAULT FALSE,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- SSH Keys table
CREATE TABLE ssh_keys (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    public_key TEXT NOT NULL,
    private_key TEXT NOT NULL,
    fingerprint VARCHAR(255),
    
    is_active BOOLEAN DEFAULT true,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Web Servers table
CREATE TABLE web_servers (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    hostname VARCHAR(255) NOT NULL,
    port INTEGER DEFAULT 22,
    username VARCHAR(255) NOT NULL,
    
    -- Connection status
    is_active BOOLEAN DEFAULT true,
    last_ping_at TIMESTAMP WITH TIME ZONE,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL
);

-- Update deployments table to include GitHub repository references
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS name VARCHAR(255);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS github_repository_id INTEGER REFERENCES git_hub_repositories(id) ON DELETE SET NULL;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS web_server_id INTEGER REFERENCES web_servers(id) ON DELETE SET NULL;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS port INTEGER DEFAULT 3000;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS branch VARCHAR(255) DEFAULT 'main';
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS commit_hash VARCHAR(255);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS build_command VARCHAR(500);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS start_command VARCHAR(500);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS deploy_logs TEXT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS error_logs TEXT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS deploy_time BIGINT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS is_auto_deploy_enabled BOOLEAN DEFAULT false;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS trigger_branch VARCHAR(255);

-- Channels table for messaging system
CREATE TABLE channels (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    name VARCHAR(255) NOT NULL,
    description TEXT,
    type VARCHAR(50) DEFAULT 'public',
    is_active BOOLEAN DEFAULT true,
    
    settings JSONB DEFAULT '{}',
    last_activity TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    message_count INTEGER DEFAULT 0,
    member_count INTEGER DEFAULT 0,
    
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    created_by VARCHAR(255) DEFAULT 'system'
);

-- Messages table for messaging system
CREATE TABLE messages (
    id VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    content TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'user',
    metadata JSONB DEFAULT '{}',
    
    channel_id VARCHAR(255) NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Add indexes for new tables
CREATE INDEX idx_git_hub_repositories_project_id ON git_hub_repositories(project_id);
CREATE INDEX idx_git_hub_repositories_full_name ON git_hub_repositories(full_name);
CREATE INDEX idx_web_servers_project_id ON web_servers(project_id);
CREATE INDEX idx_ssh_keys_project_id ON ssh_keys(project_id);
CREATE INDEX idx_channels_project_id ON channels(project_id);
CREATE INDEX idx_channels_type ON channels(type);
CREATE INDEX idx_messages_channel_id ON messages(channel_id);
CREATE INDEX idx_messages_project_id ON messages(project_id);
CREATE INDEX idx_messages_user_id ON messages(user_id);

-- Add triggers to new tables
CREATE TRIGGER update_git_hub_repositories_updated_at BEFORE UPDATE ON git_hub_repositories FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_web_servers_updated_at BEFORE UPDATE ON web_servers FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_ssh_keys_updated_at BEFORE UPDATE ON ssh_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_channels_updated_at BEFORE UPDATE ON channels FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_messages_updated_at BEFORE UPDATE ON messages FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Add triggers to existing tables
CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_projects_updated_at BEFORE UPDATE ON projects FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_api_keys_updated_at BEFORE UPDATE ON api_keys FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_cors_configs_updated_at BEFORE UPDATE ON cors_configs FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_deployments_updated_at BEFORE UPDATE ON deployments FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
CREATE TRIGGER update_backups_updated_at BEFORE UPDATE ON backups FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- Create default admin user
-- Password: 'admin123' (bcrypt hash with cost 12)
INSERT INTO users (email, name, password_hash, role, is_active) VALUES (
    'admin@cloudbox.dev',
    'Admin User',
    '$2b$12$XuhuQlzkg3wACdXfAU9mu.NMMtcvhwaf91N2S9saHt3mWjDArdAcG',
    'superadmin',
    true
);

-- Create default project
INSERT INTO projects (name, description, slug, is_active, user_id) VALUES (
    'Demo Project',
    'Default demo project for CloudBox',
    'demo-project',
    true,
    1
);

-- Create default API key for the project
INSERT INTO api_keys (name, key, key_hash, is_active, permissions, project_id) VALUES (
    'Default API Key',
    'cbx_demo_key_12345678901234567890',
    '$2a$10$demo_key_hash_placeholder_value_here',
    true,
    ARRAY['read', 'write', 'admin'],
    1
);

-- Create sample GitHub repository for webhook testing
INSERT INTO git_hub_repositories (
    name, full_name, clone_url, description, branch, is_private,
    build_command, start_command, app_port,
    is_active, project_id
) VALUES (
    'sample-app',
    'user/sample-app',
    'https://github.com/user/sample-app.git',
    'Sample application for CloudBox testing',
    'main',
    false,
    'npm run build',
    'npm start',
    3000,
    true,
    1
);