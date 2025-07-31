-- Migration: Add system settings table for admin configuration
-- Created: 2025-07-31

BEGIN;

-- Create system_settings table for admin configuration
CREATE TABLE IF NOT EXISTS system_settings (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    
    -- Setting identification
    key VARCHAR(255) NOT NULL UNIQUE,
    category VARCHAR(100) NOT NULL DEFAULT 'general',
    
    -- Setting values
    value TEXT,
    value_type VARCHAR(50) NOT NULL DEFAULT 'string', -- string, boolean, integer, json
    
    -- Setting metadata
    name VARCHAR(255) NOT NULL,
    description TEXT,
    is_secret BOOLEAN DEFAULT FALSE, -- For sensitive values like secrets
    is_required BOOLEAN DEFAULT FALSE,
    default_value TEXT,
    validation_rules JSONB, -- For validation rules
    
    -- Setting organization
    sort_order INTEGER DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE
);

-- Insert GitHub OAuth settings
INSERT INTO system_settings (key, category, value, value_type, name, description, is_secret, is_required, sort_order) VALUES
('github_oauth_client_id', 'github', '', 'string', 'GitHub Client ID', 'OAuth Client ID from your GitHub OAuth App', FALSE, TRUE, 1),
('github_oauth_client_secret', 'github', '', 'string', 'GitHub Client Secret', 'OAuth Client Secret from your GitHub OAuth App', TRUE, TRUE, 2),
('github_oauth_enabled', 'github', 'false', 'boolean', 'Enable GitHub OAuth', 'Enable per-repository GitHub OAuth authorization', FALSE, FALSE, 0),

-- General system settings
('site_name', 'general', 'CloudBox', 'string', 'Site Name', 'The name of your CloudBox installation', FALSE, FALSE, 1),
('site_domain', 'general', 'localhost:3000', 'string', 'Site Domain', 'The domain where CloudBox is hosted (used for OAuth callbacks)', FALSE, TRUE, 2),
('site_protocol', 'general', 'http', 'string', 'Site Protocol', 'Protocol for your CloudBox installation (http or https)', FALSE, TRUE, 3);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_system_settings_category ON system_settings(category);
CREATE INDEX IF NOT EXISTS idx_system_settings_key ON system_settings(key);
CREATE INDEX IF NOT EXISTS idx_system_settings_active ON system_settings(is_active);

-- Add updated_at trigger
CREATE OR REPLACE FUNCTION update_system_settings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_system_settings_updated_at
    BEFORE UPDATE ON system_settings
    FOR EACH ROW
    EXECUTE FUNCTION update_system_settings_updated_at();

-- Add comments
COMMENT ON TABLE system_settings IS 'System-wide configuration settings manageable via admin interface';
COMMENT ON COLUMN system_settings.key IS 'Unique identifier for the setting';
COMMENT ON COLUMN system_settings.category IS 'Category for grouping settings (github, general, etc.)';
COMMENT ON COLUMN system_settings.value IS 'The actual setting value';
COMMENT ON COLUMN system_settings.is_secret IS 'Whether this setting contains sensitive information';

COMMIT;