-- Migration: Create plugin_audit_logs table for security audit trail
-- Created: 2024-08-17
-- Purpose: Track all plugin operations for security monitoring and compliance

CREATE TABLE IF NOT EXISTS plugin_audit_logs (
    id SERIAL PRIMARY KEY,
    user_id INTEGER NOT NULL,
    user_email VARCHAR(255) NOT NULL,
    action VARCHAR(50) NOT NULL, -- enable, disable, install, uninstall, list_all
    plugin_name VARCHAR(100),
    old_status VARCHAR(20),
    new_status VARCHAR(20),
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN NOT NULL DEFAULT false,
    error_msg TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for performance and security monitoring
CREATE INDEX idx_plugin_audit_logs_user_id ON plugin_audit_logs(user_id);
CREATE INDEX idx_plugin_audit_logs_action ON plugin_audit_logs(action);
CREATE INDEX idx_plugin_audit_logs_plugin_name ON plugin_audit_logs(plugin_name);
CREATE INDEX idx_plugin_audit_logs_created_at ON plugin_audit_logs(created_at);
CREATE INDEX idx_plugin_audit_logs_success ON plugin_audit_logs(success);
CREATE INDEX idx_plugin_audit_logs_ip_address ON plugin_audit_logs(ip_address);

-- Create partial index for failed operations (security monitoring)
CREATE INDEX idx_plugin_audit_logs_failures ON plugin_audit_logs(created_at, ip_address, user_id) 
WHERE success = false;

-- Add comments for documentation
COMMENT ON TABLE plugin_audit_logs IS 'Security audit trail for all plugin operations';
COMMENT ON COLUMN plugin_audit_logs.action IS 'Type of operation: enable, disable, install, uninstall, list_all';
COMMENT ON COLUMN plugin_audit_logs.plugin_name IS 'Name of the plugin being operated on';
COMMENT ON COLUMN plugin_audit_logs.old_status IS 'Previous status before operation';
COMMENT ON COLUMN plugin_audit_logs.new_status IS 'New status after operation';
COMMENT ON COLUMN plugin_audit_logs.ip_address IS 'Client IP address for security tracking';
COMMENT ON COLUMN plugin_audit_logs.success IS 'Whether the operation completed successfully';
COMMENT ON COLUMN plugin_audit_logs.error_msg IS 'Error message if operation failed';

-- Grant appropriate permissions
-- GRANT SELECT, INSERT ON plugin_audit_logs TO cloudbox_app;
-- GRANT USAGE ON SEQUENCE plugin_audit_logs_id_seq TO cloudbox_app;