-- Create audit_logs table for tracking system activities and changes

CREATE TABLE IF NOT EXISTS audit_logs (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE,

    -- Action details
    action VARCHAR(100) NOT NULL,
    resource VARCHAR(100) NOT NULL,
    resource_id VARCHAR(255),
    description TEXT,

    -- Actor (who performed the action)
    actor_id INTEGER NOT NULL,
    actor_name VARCHAR(255) NOT NULL,
    actor_role VARCHAR(50) NOT NULL,

    -- Request context
    ip_address INET,
    user_agent TEXT,
    method VARCHAR(10),
    path TEXT,

    -- Additional data (JSON)
    metadata TEXT, -- JSON string for additional context

    -- Project context (if applicable)
    project_id INTEGER REFERENCES projects(id) ON DELETE SET NULL,

    -- Success/failure tracking
    success BOOLEAN DEFAULT true NOT NULL,
    error_msg TEXT
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource ON audit_logs(resource);
CREATE INDEX IF NOT EXISTS idx_audit_logs_resource_id ON audit_logs(resource_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_actor_id ON audit_logs(actor_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_project_id ON audit_logs(project_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_success ON audit_logs(success);

-- Add trigger to update updated_at timestamp
CREATE TRIGGER update_audit_logs_updated_at 
    BEFORE UPDATE ON audit_logs 
    FOR EACH ROW 
    EXECUTE FUNCTION update_updated_at_column();

-- Add comments for documentation
COMMENT ON TABLE audit_logs IS 'Audit trail for tracking system activities and changes';
COMMENT ON COLUMN audit_logs.action IS 'Type of action performed (e.g., project.create, user.update)';
COMMENT ON COLUMN audit_logs.resource IS 'Resource type affected (e.g., project, user, organization)';
COMMENT ON COLUMN audit_logs.resource_id IS 'ID of the affected resource';
COMMENT ON COLUMN audit_logs.description IS 'Human-readable description of the action';
COMMENT ON COLUMN audit_logs.actor_id IS 'ID of user who performed the action';
COMMENT ON COLUMN audit_logs.actor_name IS 'Name of user who performed the action';
COMMENT ON COLUMN audit_logs.actor_role IS 'Role of user who performed the action';
COMMENT ON COLUMN audit_logs.metadata IS 'Additional context data as JSON string';
COMMENT ON COLUMN audit_logs.success IS 'Whether the action was successful';
COMMENT ON COLUMN audit_logs.error_msg IS 'Error message if action failed';