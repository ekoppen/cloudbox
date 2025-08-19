-- Migration 007a: Create GitHub Repositories Table
-- This migration creates the github_repositories table that was missing but referenced by later migrations

CREATE TABLE IF NOT EXISTS github_repositories (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Repository information
    name VARCHAR(255) NOT NULL,
    full_name VARCHAR(255) NOT NULL, -- owner/repo format
    clone_url TEXT NOT NULL,
    branch VARCHAR(255) DEFAULT 'main',
    is_private BOOLEAN DEFAULT FALSE,
    description TEXT,
    
    -- GitHub webhook
    webhook_id BIGINT,
    webhook_secret VARCHAR(255),
    
    -- SSH Key for private repository access
    ssh_key_id INTEGER,
    
    -- GitHub OAuth for repository access
    access_token TEXT,
    token_expires_at TIMESTAMP WITH TIME ZONE,
    refresh_token TEXT,
    token_scopes VARCHAR(500), -- Comma-separated scopes
    authorized_at TIMESTAMP WITH TIME ZONE,
    authorized_by VARCHAR(255), -- GitHub username who authorized
    
    -- SDK Configuration
    sdk_version VARCHAR(50),
    app_port INTEGER DEFAULT 3000,
    build_command VARCHAR(500) DEFAULT 'npm run build',
    start_command VARCHAR(500) DEFAULT 'npm start',
    environment JSONB DEFAULT '{}',
    
    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- GitHub specific fields
    git_hub_id BIGINT, -- GitHub's internal ID for the repository
    default_branch VARCHAR(255),
    language VARCHAR(100),
    size INTEGER,
    stargazers_count INTEGER,
    forks_count INTEGER,
    
    -- Status and metadata
    status VARCHAR(50) DEFAULT 'active',
    last_sync_at TIMESTAMP WITH TIME ZONE,
    sync_status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    
    -- Indexes for performance
    UNIQUE(full_name, project_id),
    UNIQUE(git_hub_id, project_id)
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_github_repositories_project_id ON github_repositories(project_id);
CREATE INDEX IF NOT EXISTS idx_github_repositories_full_name ON github_repositories(full_name);
CREATE INDEX IF NOT EXISTS idx_github_repositories_git_hub_id ON github_repositories(git_hub_id);
CREATE INDEX IF NOT EXISTS idx_github_repositories_status ON github_repositories(status);
CREATE INDEX IF NOT EXISTS idx_github_repositories_deleted_at ON github_repositories(deleted_at);
CREATE INDEX IF NOT EXISTS idx_github_repositories_last_sync ON github_repositories(last_sync_at);

-- Foreign key constraints
ALTER TABLE github_repositories 
ADD CONSTRAINT fk_github_repositories_ssh_key 
FOREIGN KEY (ssh_key_id) REFERENCES ssh_keys(id) ON DELETE SET NULL;

-- Update trigger for updated_at
CREATE OR REPLACE FUNCTION update_github_repositories_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_github_repositories_updated_at
    BEFORE UPDATE ON github_repositories
    FOR EACH ROW
    EXECUTE FUNCTION update_github_repositories_updated_at();

-- Comments for documentation
COMMENT ON TABLE github_repositories IS 'GitHub repositories connected to CloudBox projects';
COMMENT ON COLUMN github_repositories.full_name IS 'Repository name in owner/repo format';
COMMENT ON COLUMN github_repositories.access_token IS 'OAuth access token for repository access (encrypted)';
COMMENT ON COLUMN github_repositories.git_hub_id IS 'GitHub internal ID for the repository';
COMMENT ON COLUMN github_repositories.environment IS 'JSON object containing repository-specific environment variables';

-- Migration complete
SELECT 'Migration 007a: GitHub Repositories table created successfully' as result;