-- Create functions table for serverless functions

CREATE TABLE IF NOT EXISTS functions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Function identification
    name VARCHAR(255) NOT NULL,
    description TEXT,

    -- Function code
    runtime VARCHAR(50) DEFAULT 'nodejs18' NOT NULL, -- nodejs18, python3.9, go1.19
    language VARCHAR(50) DEFAULT 'javascript' NOT NULL, -- javascript, python, go
    code TEXT NOT NULL, -- The function code
    entry_point VARCHAR(255) DEFAULT 'index.handler', -- Entry point for the function

    -- Configuration
    timeout INTEGER DEFAULT 30, -- seconds
    memory INTEGER DEFAULT 128, -- MB
    environment JSONB DEFAULT '{}',
    commands JSONB DEFAULT '[]', -- Build commands
    dependencies JSONB DEFAULT '{}',

    -- Status and deployment
    status VARCHAR(50) DEFAULT 'draft', -- draft, building, deployed, error
    version INTEGER DEFAULT 1,
    last_deployed_at TIMESTAMP WITH TIME ZONE,

    -- Runtime info
    build_logs TEXT,
    deployment_logs TEXT,

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE,

    -- Unique constraint for function name per project
    CONSTRAINT uq_function_name_project UNIQUE (name, project_id)
);

-- Create function_executions table for tracking function invocations
CREATE TABLE IF NOT EXISTS function_executions (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Execution details
    function_id INTEGER NOT NULL REFERENCES functions(id) ON DELETE CASCADE,
    execution_id VARCHAR(255) NOT NULL UNIQUE, -- UUID for tracking

    -- Request/response data
    request_data JSONB DEFAULT '{}',
    response_data JSONB DEFAULT '{}',
    headers JSONB DEFAULT '{}',
    method VARCHAR(10) NOT NULL,
    path TEXT,

    -- Execution results
    status VARCHAR(50) NOT NULL, -- success, error, timeout
    status_code INTEGER DEFAULT 200,
    execution_time BIGINT, -- milliseconds
    memory_usage BIGINT, -- bytes
    started_at TIMESTAMP WITH TIME ZONE NOT NULL,
    completed_at TIMESTAMP WITH TIME ZONE,

    -- Logs and errors
    logs TEXT,
    error_message TEXT,

    -- Metadata
    user_agent TEXT,
    client_ip INET,
    source VARCHAR(50) DEFAULT 'http', -- http, webhook, cron, manual

    -- Project relation for easier querying
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Create function_domains table for custom domains
CREATE TABLE IF NOT EXISTS function_domains (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Domain info
    domain VARCHAR(255) NOT NULL UNIQUE,
    is_verified BOOLEAN DEFAULT false,
    certificate TEXT, -- SSL certificate

    -- Function mapping
    function_id INTEGER NOT NULL REFERENCES functions(id) ON DELETE CASCADE,

    -- Project relation
    project_id INTEGER NOT NULL REFERENCES projects(id) ON DELETE CASCADE
);

-- Create indexes for performance
CREATE INDEX IF NOT EXISTS idx_functions_project_id ON functions(project_id);
CREATE INDEX IF NOT EXISTS idx_functions_name ON functions(name);
CREATE INDEX IF NOT EXISTS idx_functions_status ON functions(status);
CREATE INDEX IF NOT EXISTS idx_functions_runtime ON functions(runtime);
CREATE INDEX IF NOT EXISTS idx_function_executions_function_id ON function_executions(function_id);
CREATE INDEX IF NOT EXISTS idx_function_executions_project_id ON function_executions(project_id);
CREATE INDEX IF NOT EXISTS idx_function_executions_execution_id ON function_executions(execution_id);
CREATE INDEX IF NOT EXISTS idx_function_executions_status ON function_executions(status);
CREATE INDEX IF NOT EXISTS idx_function_executions_started_at ON function_executions(started_at);
CREATE INDEX IF NOT EXISTS idx_function_domains_function_id ON function_domains(function_id);
CREATE INDEX IF NOT EXISTS idx_function_domains_project_id ON function_domains(project_id);
CREATE INDEX IF NOT EXISTS idx_function_domains_domain ON function_domains(domain);

-- Add triggers to update updated_at timestamp
CREATE TRIGGER update_functions_updated_at 
    BEFORE UPDATE ON functions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_function_executions_updated_at 
    BEFORE UPDATE ON function_executions 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_function_domains_updated_at 
    BEFORE UPDATE ON function_domains 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE functions IS 'Serverless functions for custom backend logic';
COMMENT ON TABLE function_executions IS 'Function execution logs and metrics';
COMMENT ON TABLE function_domains IS 'Custom domains for function access';
COMMENT ON COLUMN functions.runtime IS 'Function runtime environment';
COMMENT ON COLUMN functions.language IS 'Programming language used';
COMMENT ON COLUMN functions.code IS 'Function source code';
COMMENT ON COLUMN functions.entry_point IS 'Function entry point (e.g., index.handler)';
COMMENT ON COLUMN functions.timeout IS 'Function timeout in seconds';
COMMENT ON COLUMN functions.memory IS 'Function memory limit in MB';
COMMENT ON COLUMN function_executions.execution_id IS 'Unique UUID for tracking this execution';
COMMENT ON COLUMN function_executions.execution_time IS 'Function execution time in milliseconds';
COMMENT ON COLUMN function_executions.memory_usage IS 'Memory used during execution in bytes';