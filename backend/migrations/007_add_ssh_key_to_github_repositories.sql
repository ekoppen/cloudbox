-- Add SSH key support to GitHub repositories
-- This allows users to associate SSH keys with GitHub repositories for private repository access

ALTER TABLE github_repositories 
ADD COLUMN ssh_key_id INTEGER REFERENCES ssh_keys(id) ON DELETE SET NULL;

-- Add index for better query performance
CREATE INDEX idx_github_repositories_ssh_key_id ON github_repositories(ssh_key_id);

-- Add comment for documentation
COMMENT ON COLUMN github_repositories.ssh_key_id IS 'Optional SSH key for private repository authentication';