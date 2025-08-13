-- Migration 008: Template Deployment and Repository Compatibility System
-- This migration adds support for template-based deployments and repository compatibility checking

-- Template Deployments Table
-- Tracks deployments created from templates with customization variables
CREATE TABLE IF NOT EXISTS template_deployments (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Template information
    template_name VARCHAR(100) NOT NULL,
    variables JSONB DEFAULT '{}',
    
    -- Status and relationships
    status VARCHAR(50) DEFAULT 'created',
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Related resources
    github_repository_id INTEGER NOT NULL REFERENCES github_repositories(id) ON DELETE CASCADE,
    deployment_id INTEGER NOT NULL REFERENCES deployments(id) ON DELETE CASCADE,
    
    -- Indexes
    UNIQUE(github_repository_id, deployment_id)
);

-- CloudBox Compatibility Table
-- Stores compatibility check results for repositories
CREATE TABLE IF NOT EXISTS cloudbox_compatibilities (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Repository information
    github_repository_id INTEGER UNIQUE NOT NULL REFERENCES github_repositories(id) ON DELETE CASCADE,
    
    -- Compatibility check results
    is_compatible BOOLEAN DEFAULT FALSE,
    has_cloudbox_sdk BOOLEAN DEFAULT FALSE,
    sdk_version VARCHAR(50),
    package_manager VARCHAR(20), -- npm, yarn, pnpm
    framework_type VARCHAR(50),  -- react, vue, nextjs, etc.
    
    -- CloudBox configuration detected
    has_cloudbox_config BOOLEAN DEFAULT FALSE,
    config_file VARCHAR(255),
    detected_config JSONB DEFAULT '{}',
    
    -- Environment variables detected
    env_variables TEXT[],
    required_env_vars TEXT[],
    
    -- Installation requirements
    install_commands TEXT[],
    build_commands TEXT[],
    start_commands TEXT[],
    
    -- Compatibility issues
    issues JSONB DEFAULT '[]',
    recommendations JSONB DEFAULT '[]',
    
    -- Check metadata
    checked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    check_version VARCHAR(20) DEFAULT '1.0.0',
    error_message TEXT
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_template_deployments_project_id ON template_deployments(project_id);
CREATE INDEX IF NOT EXISTS idx_template_deployments_template_name ON template_deployments(template_name);
CREATE INDEX IF NOT EXISTS idx_template_deployments_status ON template_deployments(status);
CREATE INDEX IF NOT EXISTS idx_template_deployments_deleted_at ON template_deployments(deleted_at);

CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_github_repo_id ON cloudbox_compatibilities(github_repository_id);
CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_is_compatible ON cloudbox_compatibilities(is_compatible);
CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_has_sdk ON cloudbox_compatibilities(has_cloudbox_sdk);
CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_framework ON cloudbox_compatibilities(framework_type);
CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_checked_at ON cloudbox_compatibilities(checked_at);
CREATE INDEX IF NOT EXISTS idx_cloudbox_compatibilities_deleted_at ON cloudbox_compatibilities(deleted_at);

-- Add sample data for testing
INSERT INTO template_deployments (template_name, variables, status, project_id, github_repository_id, deployment_id)
SELECT 
    'photoportfolio',
    '{"site_name": "Example Portfolio", "theme": "modern", "language": "en"}',
    'created',
    p.id,
    gr.id,
    d.id
FROM projects p
CROSS JOIN github_repositories gr
CROSS JOIN deployments d
WHERE p.slug = 'example-project' 
  AND gr.name = 'example-repo'
  AND d.name = 'example-deployment'
  AND NOT EXISTS (
    SELECT 1 FROM template_deployments 
    WHERE project_id = p.id 
      AND github_repository_id = gr.id 
      AND deployment_id = d.id
  )
LIMIT 1;

-- Add compatibility check example
INSERT INTO cloudbox_compatibilities (
    github_repository_id,
    is_compatible,
    has_cloudbox_sdk,
    sdk_version,
    package_manager,
    framework_type,
    has_cloudbox_config,
    config_file,
    detected_config,
    env_variables,
    required_env_vars,
    install_commands,
    build_commands,
    start_commands,
    issues,
    recommendations,
    checked_at,
    check_version
)
SELECT 
    gr.id,
    true,
    true,
    '^1.0.0',
    'npm',
    'react',
    true,
    '.env.example',
    '{"endpoint": "http://localhost:8080", "project_id": "1"}',
    ARRAY['CLOUDBOX_ENDPOINT', 'CLOUDBOX_API_KEY', 'CLOUDBOX_PROJECT_ID', 'NODE_ENV'],
    ARRAY['CLOUDBOX_ENDPOINT', 'CLOUDBOX_API_KEY', 'CLOUDBOX_PROJECT_ID'],
    ARRAY['npm install'],
    ARRAY['npm run build'],
    ARRAY['npm start'],
    '[]',
    '["Consider adding TypeScript for better development experience", "Add proper error handling for CloudBox API calls"]',
    NOW(),
    '1.0.0'
FROM github_repositories gr
WHERE gr.name = 'example-repo'
  AND NOT EXISTS (
    SELECT 1 FROM cloudbox_compatibilities 
    WHERE github_repository_id = gr.id
  )
LIMIT 1;

-- Update updated_at trigger for new tables
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_template_deployments_updated_at
    BEFORE UPDATE ON template_deployments
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_cloudbox_compatibilities_updated_at
    BEFORE UPDATE ON cloudbox_compatibilities
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Comments for documentation
COMMENT ON TABLE template_deployments IS 'Tracks deployments created from CloudBox templates with customization variables';
COMMENT ON TABLE cloudbox_compatibilities IS 'Stores CloudBox SDK compatibility analysis results for GitHub repositories';

COMMENT ON COLUMN template_deployments.template_name IS 'Name of the template used (photoportfolio, blog, ecommerce, etc.)';
COMMENT ON COLUMN template_deployments.variables IS 'JSON object containing template customization variables';
COMMENT ON COLUMN cloudbox_compatibilities.is_compatible IS 'Overall compatibility assessment (true if score >= 50%)';
COMMENT ON COLUMN cloudbox_compatibilities.has_cloudbox_sdk IS 'Whether CloudBox SDK dependency was detected';
COMMENT ON COLUMN cloudbox_compatibilities.detected_config IS 'CloudBox configuration found in repository';
COMMENT ON COLUMN cloudbox_compatibilities.issues IS 'Array of compatibility issues found';
COMMENT ON COLUMN cloudbox_compatibilities.recommendations IS 'Array of recommended improvements';

-- Migration complete
SELECT 'Migration 008: Template Deployment and Repository Compatibility System completed successfully' as result;