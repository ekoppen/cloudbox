-- Migration: Add GitHub OAuth fields to github_repositories table
-- Created: 2025-07-31

BEGIN;

-- Add OAuth fields for per-repository GitHub access
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS access_token TEXT;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS token_expires_at TIMESTAMP;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS refresh_token TEXT;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS token_scopes VARCHAR(255);
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS authorized_at TIMESTAMP;
ALTER TABLE git_hub_repositories ADD COLUMN IF NOT EXISTS authorized_by VARCHAR(255);

-- Add index for OAuth token lookups
CREATE INDEX IF NOT EXISTS idx_github_repos_authorized_at ON git_hub_repositories(authorized_at);
CREATE INDEX IF NOT EXISTS idx_github_repos_authorized_by ON git_hub_repositories(authorized_by);

-- Add comments for documentation
COMMENT ON COLUMN git_hub_repositories.access_token IS 'GitHub OAuth access token for repository access';
COMMENT ON COLUMN git_hub_repositories.token_expires_at IS 'When the access token expires';
COMMENT ON COLUMN git_hub_repositories.refresh_token IS 'GitHub OAuth refresh token';
COMMENT ON COLUMN git_hub_repositories.token_scopes IS 'Comma-separated OAuth scopes granted';
COMMENT ON COLUMN git_hub_repositories.authorized_at IS 'When OAuth authorization was completed';
COMMENT ON COLUMN git_hub_repositories.authorized_by IS 'GitHub username who authorized access';

COMMIT;