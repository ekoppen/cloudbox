-- Migration: Add repository analysis tables
-- Created: 2025-07-31

-- Create repository_analyses table
CREATE TABLE IF NOT EXISTS repository_analyses (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,
    
    -- Repository relation (unique - one analysis per repo)
    github_repository_id INTEGER NOT NULL UNIQUE REFERENCES git_hub_repositories(id) ON DELETE CASCADE,
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    
    -- Analysis metadata
    analyzed_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    analyzed_branch VARCHAR(255) NOT NULL DEFAULT 'main',
    analysis_status VARCHAR(50) DEFAULT 'completed',
    
    -- Project detection
    project_type VARCHAR(100),
    framework VARCHAR(100),
    language VARCHAR(100),
    package_manager VARCHAR(100),
    
    -- Build configuration
    build_command TEXT,
    start_command TEXT,
    dev_command TEXT,
    install_command TEXT,
    test_command TEXT,
    
    -- Runtime configuration
    port INTEGER DEFAULT 3000,
    environment JSONB DEFAULT '{}',
    
    -- Docker support
    has_docker BOOLEAN DEFAULT FALSE,
    docker_command TEXT,
    docker_port INTEGER,
    
    -- Dependencies and features
    dependencies JSONB DEFAULT '[]',
    dev_dependencies JSONB DEFAULT '[]',
    scripts JSONB DEFAULT '[]',
    
    -- File structure
    important_files JSONB DEFAULT '[]',
    config_files JSONB DEFAULT '[]',
    environment_files JSONB DEFAULT '[]',
    
    -- Installation options
    install_options JSONB DEFAULT '[]',
    
    -- Analysis insights
    insights JSONB DEFAULT '[]',
    warnings JSONB DEFAULT '[]',
    requirements JSONB DEFAULT '[]',
    
    -- Performance metrics
    estimated_build_time BIGINT DEFAULT 0,
    estimated_size BIGINT DEFAULT 0,
    complexity INTEGER DEFAULT 1,
    
    -- Analysis errors
    analysis_errors JSONB DEFAULT '[]'
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_repository_analyses_github_repository_id ON repository_analyses(github_repository_id);
CREATE INDEX IF NOT EXISTS idx_repository_analyses_project_id ON repository_analyses(project_id);
CREATE INDEX IF NOT EXISTS idx_repository_analyses_deleted_at ON repository_analyses(deleted_at);
CREATE INDEX IF NOT EXISTS idx_repository_analyses_analyzed_at ON repository_analyses(analyzed_at);
CREATE INDEX IF NOT EXISTS idx_repository_analyses_project_type ON repository_analyses(project_type);
CREATE INDEX IF NOT EXISTS idx_repository_analyses_analysis_status ON repository_analyses(analysis_status);

-- Add trigger to update the updated_at column
CREATE OR REPLACE FUNCTION update_repository_analyses_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_repository_analyses_updated_at
    BEFORE UPDATE ON repository_analyses
    FOR EACH ROW
    EXECUTE FUNCTION update_repository_analyses_updated_at();

-- Comment on table and important columns
COMMENT ON TABLE repository_analyses IS 'Stores detailed repository analysis results including project detection, build configuration, and install options';
COMMENT ON COLUMN repository_analyses.github_repository_id IS 'Reference to the analyzed GitHub repository (unique)';
COMMENT ON COLUMN repository_analyses.install_options IS 'JSON array of different installation/deployment options with commands and configurations';
COMMENT ON COLUMN repository_analyses.insights IS 'JSON array of helpful suggestions and recommendations for deployment';
COMMENT ON COLUMN repository_analyses.complexity IS 'Project complexity score from 1-10 based on framework, dependencies, and configuration';