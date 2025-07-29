-- Add pending update fields to github repositories
ALTER TABLE git_hub_repositories 
ADD COLUMN pending_commit_hash VARCHAR(255) DEFAULT '',
ADD COLUMN pending_commit_branch VARCHAR(255) DEFAULT '',
ADD COLUMN has_pending_update BOOLEAN DEFAULT FALSE;

-- Add index for efficient querying of repositories with pending updates
CREATE INDEX idx_github_repos_pending_update ON git_hub_repositories(has_pending_update) WHERE has_pending_update = true;