-- Update projects table to add missing columns from the Project model

-- Add notes column for project notes
ALTER TABLE projects ADD COLUMN IF NOT EXISTS notes TEXT;

-- Update users table to add missing columns from User model
ALTER TABLE users ADD COLUMN IF NOT EXISTS role VARCHAR(20) DEFAULT 'admin';

-- Update the existing role column to match the model's default
UPDATE users SET role = 'admin' WHERE role = 'user';

-- Update api_keys table to match APIKey model structure
-- The key column should not exist in production (only key_hash should be stored)
-- But we'll keep it for now and add a comment about security

-- Add missing columns to messages table
ALTER TABLE messages ADD COLUMN IF NOT EXISTS parent_id VARCHAR(255) REFERENCES messages(id) ON DELETE SET NULL;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS thread_id VARCHAR(255);
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reply_count INTEGER DEFAULT 0;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_edited BOOLEAN DEFAULT false;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS edited_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS is_deleted BOOLEAN DEFAULT false;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS message_deleted_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE messages ADD COLUMN IF NOT EXISTS reaction_count INTEGER DEFAULT 0;

-- Add missing columns to channels table
ALTER TABLE channels ADD COLUMN IF NOT EXISTS topic TEXT;
ALTER TABLE channels ADD COLUMN IF NOT EXISTS max_members INTEGER DEFAULT 0; -- 0 = unlimited

-- Update deployments table with missing columns
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS subdomain VARCHAR(255);
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS deploy_path TEXT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS port_configuration JSONB DEFAULT '{}';
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS commit_message TEXT;
ALTER TABLE deployments ADD COLUMN IF NOT EXISTS commit_author VARCHAR(255);

-- Update web_servers table with missing columns
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS server_type VARCHAR(50) DEFAULT 'vps';
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS os VARCHAR(50) DEFAULT 'ubuntu';
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS docker_enabled BOOLEAN DEFAULT true;
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS nginx_enabled BOOLEAN DEFAULT true;
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS deploy_path TEXT DEFAULT '~/deploys';
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS backup_path TEXT DEFAULT '/var/backups';
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS log_path TEXT DEFAULT '/var/log/deployments';
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS last_connected_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE web_servers ADD COLUMN IF NOT EXISTS connection_status VARCHAR(50) DEFAULT 'unknown';

-- Update ssh_keys table with missing columns
ALTER TABLE ssh_keys ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE ssh_keys ADD COLUMN IF NOT EXISTS key_type VARCHAR(20) DEFAULT 'rsa';
ALTER TABLE ssh_keys ADD COLUMN IF NOT EXISTS key_size INTEGER DEFAULT 2048;
ALTER TABLE ssh_keys ADD COLUMN IF NOT EXISTS last_used_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE ssh_keys ADD COLUMN IF NOT EXISTS server_count INTEGER DEFAULT 0;

-- Update git_hub_repositories table with missing columns
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS webhook_id BIGINT;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS webhook_secret VARCHAR(255);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS access_token VARCHAR(500);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS refresh_token VARCHAR(500);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS token_scopes VARCHAR(255);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS authorized_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS authorized_by VARCHAR(255);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS sdk_version VARCHAR(50);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS environment JSONB DEFAULT '{}';
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS git_hub_id BIGINT;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS default_branch VARCHAR(255);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS language VARCHAR(100);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS size INTEGER;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS stargazers_count INTEGER;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS forks_count INTEGER;

-- Add new indexes for better performance
CREATE INDEX IF NOT EXISTS idx_messages_parent_id ON messages(parent_id);
CREATE INDEX IF NOT EXISTS idx_messages_thread_id ON messages(thread_id);
CREATE INDEX IF NOT EXISTS idx_messages_is_deleted ON messages(is_deleted);
CREATE INDEX IF NOT EXISTS idx_git_hub_repositories_git_hub_id ON git_hub_repositories(git_hub_id);
CREATE INDEX IF NOT EXISTS idx_git_hub_repositories_ssh_key_id ON git_hub_repositories(ssh_key_id);
CREATE INDEX IF NOT EXISTS idx_web_servers_connection_status ON web_servers(connection_status);
CREATE INDEX IF NOT EXISTS idx_ssh_keys_key_type ON ssh_keys(key_type);

-- Add comments for documentation
COMMENT ON COLUMN projects.notes IS 'Project notes and additional information';
COMMENT ON COLUMN messages.parent_id IS 'Parent message ID for replies';
COMMENT ON COLUMN messages.thread_id IS 'Thread identifier for message threads';
COMMENT ON COLUMN messages.reply_count IS 'Number of replies to this message';
COMMENT ON COLUMN deployments.port_configuration IS 'Port mappings: variable -> port';
COMMENT ON COLUMN git_hub_repositories.webhook_secret IS 'Secret for GitHub webhook verification';
COMMENT ON COLUMN git_hub_repositories.access_token IS 'GitHub OAuth access token (encrypted)';
COMMENT ON COLUMN git_hub_repositories.git_hub_id IS 'GitHub repository ID';
COMMENT ON COLUMN web_servers.server_type IS 'Type of server (vps, dedicated, cloud)';
COMMENT ON COLUMN web_servers.connection_status IS 'Current connection status to server';